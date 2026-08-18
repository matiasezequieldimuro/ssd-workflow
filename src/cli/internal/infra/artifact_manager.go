package infra

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"sdd-cli/internal/domain"
)

type ArtifactManager struct{}

func NewArtifactManager() *ArtifactManager {
	return &ArtifactManager{}
}

func (am *ArtifactManager) CreateArtifactsForPhase(
	baseDir string,
	workflow *domain.Workflow,
	phaseID string,
	workItemID string,
	templateVars map[string]string,
) error {
	if err := domain.ValidateIdentifier("work item id", workItemID); err != nil {
		return err
	}
	phase, exists := workflow.Phase(phaseID)
	if !exists {
		return fmt.Errorf("phase %s not found in workflow %s", phaseID, workflow.ID)
	}

	for _, artifactID := range phase.Produces {
		artifactConfig, exists := workflow.Artifacts[artifactID]
		if !exists {
			return fmt.Errorf("artifact %s defined in phase but not in workflow artifacts", artifactID)
		}

		templateContent, err := am.readLocalTemplate(baseDir, artifactConfig.Template)
		if err != nil {
			return fmt.Errorf("failed to read template for artifact %s: %w", artifactID, err)
		}

		vars := copyTemplateVars(templateVars)
		vars["artifact_id"] = artifactID
		vars["phase"] = phaseID
		vars["created_at"] = time.Now().UTC().Format(time.RFC3339)
		vars["created_by_kind"] = string(domain.ActorCLI)
		vars["created_by_id"] = "sdd"
		vars["sources"] = artifactSources(workflow, phase)

		renderedContent := domain.RenderTemplate(templateContent, vars)
		if strings.Contains(renderedContent, "{{") {
			return fmt.Errorf("%w: template %s contains unresolved placeholders", domain.ErrSchemaValidation, artifactConfig.Template)
		}
		metadata, err := extractFrontMatter(renderedContent)
		if err != nil {
			return fmt.Errorf("%w: artifact %s: %v", domain.ErrSchemaValidation, artifactID, err)
		}
		if err := NewSchemaValidator().ValidateYAML(baseDir, "artifact.schema.json", metadata); err != nil {
			return err
		}
		if err := validateArtifactMetadata(metadata, artifactID, phaseID, workItemID); err != nil {
			return err
		}

		artifactPath, err := containedPath(
			filepath.Join(baseDir, ".sdd"),
			"work-items",
			"active",
			workItemID,
			artifactConfig.Path,
		)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory: %w", err)
		}

		if err := os.WriteFile(artifactPath, []byte(renderedContent), 0644); err != nil {
			return fmt.Errorf("failed to write artifact file %s: %w", artifactPath, err)
		}
	}

	return nil
}

func (am *ArtifactManager) ImportExternalArtifact(
	baseDir string,
	workflow *domain.Workflow,
	phaseID string,
	workItemID string,
	artifactID string,
	sourcePath string,
) error {
	phase, exists := workflow.Phase(phaseID)
	if !exists {
		return domain.ErrPhaseNotFound
	}
	producesArtifact := false
	for _, producedID := range phase.Produces {
		if producedID == artifactID {
			producesArtifact = true
			break
		}
	}
	if !producesArtifact {
		return fmt.Errorf("%w: phase %s does not produce artifact %s", domain.ErrInvalidExternalArtifact, phaseID, artifactID)
	}

	artifactConfig := workflow.Artifacts[artifactID]
	targetPath, err := containedPath(
		filepath.Join(baseDir, ".sdd"),
		"work-items",
		"active",
		workItemID,
		artifactConfig.Path,
	)
	if err != nil {
		return err
	}
	generatedContent, err := os.ReadFile(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read generated artifact %s: %w", targetPath, err)
	}
	metadata, err := extractFrontMatter(string(generatedContent))
	if err != nil {
		return fmt.Errorf("%w: generated artifact %s: %v", domain.ErrSchemaValidation, artifactID, err)
	}

	externalContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("%w: read %s: %v", domain.ErrInvalidExternalArtifact, sourcePath, err)
	}
	importedContent := "---\n" + string(metadata) + "\n---\n\n" + stripFrontMatter(string(externalContent))
	if err := os.WriteFile(targetPath, []byte(importedContent), 0644); err != nil {
		return fmt.Errorf("failed to import external artifact %s: %w", targetPath, err)
	}
	return nil
}

func (am *ArtifactManager) readLocalTemplate(baseDir, templateName string) (string, error) {
	if err := domain.ValidateIdentifier("template id", templateName); err != nil {
		return "", err
	}
	templatePath, err := containedPath(filepath.Join(baseDir, ".sdd"), "templates", templateName+".md")
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templatePath, err)
	}

	return string(data), nil
}

func copyTemplateVars(source map[string]string) map[string]string {
	result := make(map[string]string, len(source)+6)
	for key, value := range source {
		result[key] = value
	}
	return result
}

func artifactSources(workflow *domain.Workflow, phase domain.WorkflowPhase) string {
	if len(phase.Requires) == 0 {
		return "[]"
	}
	var lines []string
	for _, requiredPhaseID := range phase.Requires {
		sourcePath := workflow.ArtifactPathForPhase(requiredPhaseID)
		if sourcePath != "" {
			lines = append(lines, "  - "+sourcePath)
		}
	}
	if len(lines) == 0 {
		return "[]"
	}
	return "\n" + strings.Join(lines, "\n")
}

func extractFrontMatter(content string) ([]byte, error) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("missing YAML front matter")
	}
	remaining := content[len("---\n"):]
	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return nil, fmt.Errorf("unterminated YAML front matter")
	}
	return []byte(remaining[:end]), nil
}

func stripFrontMatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return strings.TrimSpace(content) + "\n"
	}
	remaining := content[len("---\n"):]
	end := strings.Index(remaining, "\n---")
	if end < 0 {
		return strings.TrimSpace(content) + "\n"
	}
	return strings.TrimSpace(remaining[end+len("\n---"):]) + "\n"
}

func validateArtifactMetadata(data []byte, artifactID, phaseID, workItemID string) error {
	var metadata struct {
		ID       string `yaml:"id"`
		Phase    string `yaml:"phase"`
		WorkItem string `yaml:"work_item"`
	}
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return fmt.Errorf("%w: parse artifact metadata: %v", domain.ErrSchemaValidation, err)
	}
	if metadata.ID != artifactID || metadata.Phase != phaseID || metadata.WorkItem != workItemID {
		return fmt.Errorf(
			"%w: artifact metadata mismatch: id=%s phase=%s work_item=%s",
			domain.ErrSchemaValidation,
			metadata.ID,
			metadata.Phase,
			metadata.WorkItem,
		)
	}
	return nil
}
