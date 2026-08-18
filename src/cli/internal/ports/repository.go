package ports

import "sdd-cli/internal/domain"

type WorkItemRepository interface {
	SaveWorkItem(baseDir string, item *domain.WorkItem) error
	GetWorkItem(baseDir string, id string) (*domain.WorkItem, error)
	WorkItemExists(baseDir string, id string) (bool, error)
	AppendEvent(baseDir string, id string, event domain.Event) error
}

type WorkflowRepository interface {
	GetWorkflow(baseDir string, workflowID string) (*domain.Workflow, error)
}

type ConfigRepository interface {
	GetConfig(baseDir string) (*domain.Config, error)
}
