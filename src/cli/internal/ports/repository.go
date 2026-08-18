package ports

import (
	"io/fs"

	"sdd-cli/internal/domain"
)

type ArtifactWrite struct {
	Path    string
	Content []byte
	Mode    fs.FileMode
}

type WorkItemCommit struct {
	Item        *domain.WorkItem
	Artifacts   []ArtifactWrite
	Events      []domain.Event
	OperationID string
}

type WorkItemRepository interface {
	GetWorkItem(baseDir string, id string) (*domain.WorkItem, error)
	WorkItemExists(baseDir string, id string) (bool, error)
	OperationApplied(baseDir string, id string, operationID string) (bool, error)
	CommitWorkItem(baseDir string, commit WorkItemCommit) error
}

type WorkflowRepository interface {
	GetWorkflow(baseDir string, workflowID string) (*domain.Workflow, error)
}

type ConfigRepository interface {
	GetConfig(baseDir string) (*domain.Config, error)
}
