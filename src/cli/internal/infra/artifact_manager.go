package infra

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"sdd-cli/embeds"
	"sdd-cli/internal/domain"
)

// ArtifactManager maneja la creación de archivos de artifacts basados en templates
type ArtifactManager struct{}

func NewArtifactManager() *ArtifactManager {
	return &ArtifactManager{}
}

// CreateArtifactsForPhase crea archivos Markdown para los artifacts producidos por una fase
func (am *ArtifactManager) CreateArtifactsForPhase(
	baseDir string,
	workflow *domain.Workflow,
	phaseID string,
	workItemID string,
	templateVars map[string]string,
) error {
	// Encontrar la fase en el workflow
	var phase *domain.WorkflowPhase
	for i := range workflow.Phases {
		if workflow.Phases[i].ID == phaseID {
			phase = &workflow.Phases[i]
			break
		}
	}

	if phase == nil {
		return fmt.Errorf("phase %s not found in workflow %s", phaseID, workflow.ID)
	}

	// Para cada artifact que produce esta fase
	for _, artifactID := range phase.Produces {
		artifactConfig, exists := workflow.Artifacts[artifactID]
		if !exists {
			return fmt.Errorf("artifact %s defined in phase but not in workflow artifacts", artifactID)
		}

		// Leer el template desde embeds
		templateContent, err := am.readTemplateFromEmbeds(artifactConfig.Template)
		if err != nil {
			return fmt.Errorf("failed to read template for artifact %s: %w", artifactID, err)
		}

		// Renderizar template con variables
		renderedContent := domain.RenderTemplate(templateContent, templateVars)

		// Crear archivo en work-item
		artifactPath := filepath.Join(baseDir, ".sdd", "work-items", "active", workItemID, artifactConfig.Path)
		if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
			return fmt.Errorf("failed to create artifact directory: %w", err)
		}

		if err := os.WriteFile(artifactPath, []byte(renderedContent), 0644); err != nil {
			return fmt.Errorf("failed to write artifact file %s: %w", artifactPath, err)
		}
	}

	return nil
}

// readTemplateFromEmbeds lee un template desde los recursos embebidos
func (am *ArtifactManager) readTemplateFromEmbeds(templateName string) (string, error) {
	subFS, err := fs.Sub(embeds.DefaultSDDResources, "default_sdd/templates")
	if err != nil {
		return "", fmt.Errorf("failed to access embedded templates: %w", err)
	}

	templateFile := templateName + ".md"
	data, err := fs.ReadFile(subFS, templateFile)
	if err != nil {
		return "", fmt.Errorf("failed to read template file %s: %w", templateFile, err)
	}

	return string(data), nil
}
