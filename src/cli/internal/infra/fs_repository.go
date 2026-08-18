package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type FSWorkItemRepository struct{}

func NewFSWorkItemRepository() ports.WorkItemRepository {
	return &FSWorkItemRepository{}
}

func (r *FSWorkItemRepository) getManifestPath(baseDir, id string) (string, error) {
	if err := domain.ValidateIdentifier("work item id", id); err != nil {
		return "", err
	}
	return containedPath(filepath.Join(baseDir, ".sdd"), "work-items", "active", id, "manifest.yaml")
}

func (r *FSWorkItemRepository) getEventsPath(baseDir, id string) (string, error) {
	if err := domain.ValidateIdentifier("work item id", id); err != nil {
		return "", err
	}
	return containedPath(filepath.Join(baseDir, ".sdd"), "work-items", "active", id, "events.jsonl")
}

func (r *FSWorkItemRepository) WorkItemExists(baseDir, id string) (bool, error) {
	path, err := r.getManifestPath(baseDir, id)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to inspect manifest.yaml: %w", err)
}

func (r *FSWorkItemRepository) SaveWorkItem(baseDir string, item *domain.WorkItem) error {
	if err := domain.ValidateIdentifier("work item id", item.ID); err != nil {
		return err
	}
	workflow, err := NewFSWorkflowRepository().GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return fmt.Errorf("failed to validate work item workflow: %w", err)
	}
	if err := item.ValidateAgainst(workflow); err != nil {
		return err
	}
	if err := NewSchemaValidator().ValidateValue(baseDir, "work-item.schema.json", item); err != nil {
		return err
	}

	sddRoot := filepath.Join(baseDir, ".sdd")
	artifactsDir, err := containedPath(sddRoot, "work-items", "active", item.ID, "artifacts")
	if err != nil {
		return err
	}
	evidenceDir, err := containedPath(sddRoot, "work-items", "active", item.ID, "evidence")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}
	if err := os.MkdirAll(evidenceDir, 0755); err != nil {
		return fmt.Errorf("failed to create evidence directory: %w", err)
	}

	manifestPath, err := r.getManifestPath(baseDir, item.ID)
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal work item manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest.yaml: %w", err)
	}

	return nil
}

func (r *FSWorkItemRepository) GetWorkItem(baseDir, id string) (*domain.WorkItem, error) {
	manifestPath, err := r.getManifestPath(baseDir, id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrWorkItemNotFound
		}
		return nil, fmt.Errorf("failed to read manifest.yaml: %w", err)
	}
	if err := NewSchemaValidator().ValidateYAML(baseDir, "work-item.schema.json", data); err != nil {
		return nil, err
	}

	var item domain.WorkItem
	if err := yaml.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.yaml: %w", err)
	}
	if item.ID != id {
		return nil, fmt.Errorf("%w: manifest id %s does not match requested id %s", domain.ErrInvalidWorkItem, item.ID, id)
	}
	workflow, err := NewFSWorkflowRepository().GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate work item workflow: %w", err)
	}
	if err := item.ValidateAgainst(workflow); err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *FSWorkItemRepository) AppendEvent(baseDir, id string, event domain.Event) error {
	if err := domain.ValidateIdentifier("work item id", id); err != nil {
		return err
	}
	if event.WorkItem != id {
		return fmt.Errorf("%w: event work item %s does not match %s", domain.ErrSchemaValidation, event.WorkItem, id)
	}
	if err := domain.ValidateActor(event.Actor); err != nil {
		return err
	}
	if err := NewSchemaValidator().ValidateValue(baseDir, "event.schema.json", event); err != nil {
		return err
	}
	eventsPath, err := r.getEventsPath(baseDir, id)
	if err != nil {
		return err
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event to JSON: %w", err)
	}

	f, err := os.OpenFile(eventsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open events.jsonl: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to append to events.jsonl: %w", err)
	}

	return nil
}

type FSWorkflowRepository struct{}

func NewFSWorkflowRepository() ports.WorkflowRepository {
	return &FSWorkflowRepository{}
}

func (r *FSWorkflowRepository) GetWorkflow(baseDir, workflowID string) (*domain.Workflow, error) {
	if err := domain.ValidateIdentifier("workflow id", workflowID); err != nil {
		return nil, err
	}
	path, err := containedPath(filepath.Join(baseDir, ".sdd"), "workflows", workflowID+".workflow.yaml")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}
	if err := NewSchemaValidator().ValidateYAML(baseDir, "workflow.schema.json", data); err != nil {
		return nil, err
	}

	var wf domain.Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}
	if wf.ID != workflowID {
		return nil, fmt.Errorf("%w: workflow id %s does not match filename %s", domain.ErrInvalidWorkflow, wf.ID, workflowID)
	}
	if err := wf.ValidateSemantics(); err != nil {
		return nil, err
	}
	if err := validateWorkflowTemplates(baseDir, &wf); err != nil {
		return nil, err
	}

	return &wf, nil
}

func validateWorkflowTemplates(baseDir string, workflow *domain.Workflow) error {
	sddRoot := filepath.Join(baseDir, ".sdd")
	for artifactID, artifact := range workflow.Artifacts {
		templatePath, err := containedPath(sddRoot, "templates", artifact.Template+".md")
		if err != nil {
			return err
		}
		info, err := os.Stat(templatePath)
		if err != nil {
			return fmt.Errorf("%w: artifact %s template %s: %v", domain.ErrInvalidWorkflow, artifactID, artifact.Template, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%w: artifact %s template %s is not a regular file", domain.ErrInvalidWorkflow, artifactID, artifact.Template)
		}
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("%w: read template %s: %v", domain.ErrInvalidWorkflow, artifact.Template, err)
		}
		phaseID := ""
		var phase domain.WorkflowPhase
		for _, candidate := range workflow.Phases {
			for _, producedID := range candidate.Produces {
				if producedID == artifactID {
					phaseID = candidate.ID
					phase = candidate
					break
				}
			}
			if phaseID != "" {
				break
			}
		}
		rendered := domain.RenderTemplate(string(content), map[string]string{
			"title":           "Template validation",
			"id":              "template-validation",
			"type":            workflow.WorkItemType,
			"artifact_id":     artifactID,
			"phase":           phaseID,
			"created_at":      "2026-08-18T00:00:00Z",
			"created_by_kind": string(domain.ActorCLI),
			"created_by_id":   "sdd",
			"sources":         artifactSources(workflow, phase),
		})
		if strings.Contains(rendered, "{{") {
			return fmt.Errorf("%w: template %s contains unresolved placeholders", domain.ErrInvalidWorkflow, artifact.Template)
		}
		metadata, err := extractFrontMatter(rendered)
		if err != nil {
			return fmt.Errorf("%w: template %s: %v", domain.ErrInvalidWorkflow, artifact.Template, err)
		}
		if err := NewSchemaValidator().ValidateYAML(baseDir, "artifact.schema.json", metadata); err != nil {
			return fmt.Errorf("%w: template %s: %v", domain.ErrInvalidWorkflow, artifact.Template, err)
		}
		if err := validateArtifactMetadata(metadata, artifactID, phaseID, "template-validation"); err != nil {
			return fmt.Errorf("%w: template %s: %v", domain.ErrInvalidWorkflow, artifact.Template, err)
		}
	}
	return nil
}
