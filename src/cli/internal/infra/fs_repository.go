package infra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type FSWorkItemRepository struct{}

func NewFSWorkItemRepository() ports.WorkItemRepository {
	return &FSWorkItemRepository{}
}

func (r *FSWorkItemRepository) getManifestPath(baseDir, id string) string {
	return filepath.Join(baseDir, ".sdd", "work-items", "active", id, "manifest.yaml")
}

func (r *FSWorkItemRepository) getEventsPath(baseDir, id string) string {
	return filepath.Join(baseDir, ".sdd", "work-items", "active", id, "events.jsonl")
}

func (r *FSWorkItemRepository) WorkItemExists(baseDir, id string) bool {
	path := r.getManifestPath(baseDir, id)
	_, err := os.Stat(path)
	return err == nil
}

func (r *FSWorkItemRepository) SaveWorkItem(baseDir string, item *domain.WorkItem) error {
	dir := filepath.Join(baseDir, ".sdd", "work-items", "active", item.ID)
	artifactsDir := filepath.Join(dir, "artifacts")
	evidenceDir := filepath.Join(dir, "evidence")

	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}
	if err := os.MkdirAll(evidenceDir, 0755); err != nil {
		return fmt.Errorf("failed to create evidence directory: %w", err)
	}

	manifestPath := r.getManifestPath(baseDir, item.ID)
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
	manifestPath := r.getManifestPath(baseDir, id)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrWorkItemNotFound
		}
		return nil, fmt.Errorf("failed to read manifest.yaml: %w", err)
	}

	var item domain.WorkItem
	if err := yaml.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("failed to parse manifest.yaml: %w", err)
	}

	return &item, nil
}

func (r *FSWorkItemRepository) AppendEvent(baseDir, id string, event domain.Event) error {
	eventsPath := r.getEventsPath(baseDir, id)

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
	path := filepath.Join(baseDir, ".sdd", "workflows", workflowID+".workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var wf domain.Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("failed to parse workflow YAML: %w", err)
	}

	return &wf, nil
}
