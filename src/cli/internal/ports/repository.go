package ports

import (
	"io/fs"
	"time"

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

type WorkItemReader interface {
	GetWorkItem(baseDir string, id string) (*domain.WorkItem, error)
}

type WorkItemExistenceChecker interface {
	WorkItemExists(baseDir string, id string) (bool, error)
}

type OperationTracker interface {
	OperationApplied(baseDir string, id string, operationID string) (bool, error)
}

type WorkItemCommitter interface {
	CommitWorkItem(baseDir string, commit WorkItemCommit) error
}

type WorkItemMutationRepository interface {
	WorkItemReader
	OperationTracker
	WorkItemCommitter
}

type WorkItemCreationRepository interface {
	WorkItemMutationRepository
	WorkItemExistenceChecker
}

type WorkflowRepository interface {
	GetWorkflow(baseDir string, workflowID string) (*domain.Workflow, error)
}

type ConfigRepository interface {
	GetConfig(baseDir string) (*domain.Config, error)
}

type ExternalArtifact struct {
	Path    string
	SHA256  string
	Content []byte
}

type ArtifactPreparer interface {
	PrepareArtifactsForPhase(
		baseDir string,
		workflow *domain.Workflow,
		phaseID string,
		workItemID string,
		templateVars map[string]string,
	) ([]ArtifactWrite, error)
}

type ExternalArtifactImporter interface {
	ResolveExternalArtifact(path string) (ExternalArtifact, error)
	ImportExternalArtifact(
		workflow *domain.Workflow,
		phaseID string,
		artifactID string,
		source ExternalArtifact,
		writes []ArtifactWrite,
	) ([]ArtifactWrite, error)
}

type ArtifactService interface {
	ArtifactPreparer
	ExternalArtifactImporter
}

type ProjectInitializer interface {
	Initialize(targetDir string) error
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}
