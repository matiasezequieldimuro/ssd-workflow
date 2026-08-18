package usecases_test

import (
	"io/fs"
	"strconv"
	"testing"
	"time"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
	"sdd-cli/internal/usecases"
)

func TestStartUseCaseUsesInjectedServicesWithoutFilesystem(t *testing.T) {
	workflow := memoryWorkflow()
	repository := &memoryWorkItemRepository{}
	artifacts := &memoryArtifactService{}
	clock := fixedClock{value: time.Date(2026, time.August, 18, 19, 0, 0, 0, time.UTC)}
	ids := &sequenceIDGenerator{}
	useCase := usecases.NewStartWorkItemUseCase(
		repository,
		staticWorkflowRepository{workflow: workflow},
		staticConfigRepository{workflowID: workflow.ID},
		artifacts,
		clock,
		ids,
	)

	item, err := useCase.Execute("unused", usecases.StartWorkItemInput{
		ID:      "memory-item",
		Title:   "Memory item",
		Summary: "No filesystem required",
		Actor:   domain.Actor{Kind: domain.ActorHuman, ID: "matias"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if item.CreatedAt != "2026-08-18T19:00:00Z" {
		t.Fatalf("CreatedAt = %s", item.CreatedAt)
	}
	if repository.commits != 1 || len(repository.events) != 2 {
		t.Fatalf("commits = %d, events = %d", repository.commits, len(repository.events))
	}
	if repository.events[0].Type != "work_item.created" ||
		repository.events[1].Type != "phase.transitioned" {
		t.Fatalf("event order = %s, %s", repository.events[0].Type, repository.events[1].Type)
	}
	if repository.events[0].ID == repository.events[1].ID {
		t.Fatalf("event IDs are not unique: %s", repository.events[0].ID)
	}
	if repository.events[1].Data["from"] != domain.PhaseReady ||
		repository.events[1].Data["to"] != domain.PhaseInProgress {
		t.Fatalf("transition data = %#v", repository.events[1].Data)
	}
	if len(repository.artifacts) != 1 || artifacts.prepared != 1 {
		t.Fatalf("artifact writes = %d, prepare calls = %d", len(repository.artifacts), artifacts.prepared)
	}
}

func TestInitUseCaseDelegatesToInjectedInitializer(t *testing.T) {
	initializer := &memoryInitializer{}
	if err := usecases.NewInitUseCase(initializer).Execute("project"); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if initializer.target != "project" {
		t.Fatalf("initializer target = %s", initializer.target)
	}
}

func TestExternalStartKeepsLogicalEventOrderWithFixedClock(t *testing.T) {
	workflow := memoryWorkflow()
	repository := &memoryWorkItemRepository{}
	artifacts := &memoryArtifactService{
		resolved: ports.ExternalArtifact{
			Path:    "/external/first.md",
			SHA256:  "hash",
			Content: []byte("external"),
		},
	}
	useCase := usecases.NewStartWorkItemUseCase(
		repository,
		staticWorkflowRepository{workflow: workflow},
		staticConfigRepository{workflowID: workflow.ID},
		artifacts,
		fixedClock{value: time.Date(2026, time.August, 18, 19, 0, 0, 0, time.UTC)},
		&sequenceIDGenerator{},
	)

	if _, err := useCase.Execute("unused", usecases.StartWorkItemInput{
		ID:           "external-item",
		Title:        "External item",
		FromArtifact: "first.md",
		Phase:        "first",
		Actor:        domain.Actor{Kind: domain.ActorHuman, ID: "matias"},
	}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := []string{
		"work_item.created",
		"phase.bypassed_by_external_input",
		"phase.transitioned",
	}
	if len(repository.events) != len(want) {
		t.Fatalf("events = %#v", repository.events)
	}
	seen := make(map[string]struct{}, len(repository.events))
	for index, event := range repository.events {
		if event.Type != want[index] {
			t.Fatalf("event %d type = %s, want %s", index, event.Type, want[index])
		}
		if _, exists := seen[event.ID]; exists {
			t.Fatalf("duplicate event id = %s", event.ID)
		}
		seen[event.ID] = struct{}{}
	}
}

func TestStatusUseCaseOrdersPhasesByWorkflowGraph(t *testing.T) {
	workflow := memoryWorkflow()
	workflow.Phases = []domain.WorkflowPhase{
		{ID: "second", Requires: []string{"first"}, Produces: []string{"second"}, Approval: domain.ApprovalNone},
		{ID: "first", Produces: []string{"first"}, Approval: domain.ApprovalNone},
	}
	workflow.Artifacts["second"] = domain.WorkflowArtifact{Path: "artifacts/second.md", Template: "second"}
	item := &domain.WorkItem{
		ID:       "ordered-item",
		Workflow: domain.WorkItemWorkflow{ID: workflow.ID},
		Phases: map[string]domain.PhaseState{
			"first":  {Status: domain.PhaseCompleted},
			"second": {Status: domain.PhaseReady},
		},
	}
	repository := &memoryWorkItemRepository{item: item}
	result, err := usecases.NewStatusUseCase(
		repository,
		staticWorkflowRepository{workflow: workflow},
	).Execute("unused", item.ID)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.OrderedPhases[0].ID != "first" || result.OrderedPhases[1].ID != "second" {
		t.Fatalf("phase order = %#v", result.OrderedPhases)
	}
}

type memoryWorkItemRepository struct {
	item      *domain.WorkItem
	events    []domain.Event
	artifacts []ports.ArtifactWrite
	commits   int
}

func (repository *memoryWorkItemRepository) GetWorkItem(_ string, _ string) (*domain.WorkItem, error) {
	if repository.item == nil {
		return nil, domain.ErrWorkItemNotFound
	}
	return repository.item, nil
}

func (repository *memoryWorkItemRepository) WorkItemExists(_ string, _ string) (bool, error) {
	return repository.item != nil, nil
}

func (repository *memoryWorkItemRepository) OperationApplied(_ string, _ string, _ string) (bool, error) {
	return false, nil
}

func (repository *memoryWorkItemRepository) CommitWorkItem(_ string, commit ports.WorkItemCommit) error {
	repository.item = commit.Item
	repository.events = append(repository.events, commit.Events...)
	repository.artifacts = append(repository.artifacts, commit.Artifacts...)
	repository.commits++
	return nil
}

type staticWorkflowRepository struct {
	workflow *domain.Workflow
}

func (repository staticWorkflowRepository) GetWorkflow(_ string, _ string) (*domain.Workflow, error) {
	return repository.workflow, nil
}

type staticConfigRepository struct {
	workflowID string
}

func (repository staticConfigRepository) GetConfig(_ string) (*domain.Config, error) {
	return &domain.Config{
		SchemaVersion: "0.1",
		Defaults: domain.ConfigDefaults{
			Workflow: repository.workflowID,
		},
	}, nil
}

type memoryArtifactService struct {
	prepared int
	resolved ports.ExternalArtifact
}

func (service *memoryArtifactService) PrepareArtifactsForPhase(
	_ string,
	_ *domain.Workflow,
	_ string,
	_ string,
	_ map[string]string,
) ([]ports.ArtifactWrite, error) {
	service.prepared++
	return []ports.ArtifactWrite{{
		Path:    "artifacts/first.md",
		Content: []byte("artifact"),
		Mode:    fs.FileMode(0644),
	}}, nil
}

func (service *memoryArtifactService) ResolveExternalArtifact(_ string) (ports.ExternalArtifact, error) {
	return service.resolved, nil
}

func (service *memoryArtifactService) ImportExternalArtifact(
	_ *domain.Workflow,
	_ string,
	_ string,
	_ ports.ExternalArtifact,
	writes []ports.ArtifactWrite,
) ([]ports.ArtifactWrite, error) {
	return writes, nil
}

type fixedClock struct {
	value time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.value
}

type sequenceIDGenerator struct {
	next int
}

func (generator *sequenceIDGenerator) NewID() (string, error) {
	generator.next++
	return "evt-sequence-" + strconv.Itoa(generator.next), nil
}

type memoryInitializer struct {
	target string
}

func (initializer *memoryInitializer) Initialize(targetDir string) error {
	initializer.target = targetDir
	return nil
}

func memoryWorkflow() *domain.Workflow {
	return &domain.Workflow{
		SchemaVersion: "0.1",
		Kind:          "workflow",
		ID:            "memory-workflow",
		Title:         "Memory workflow",
		WorkItemType:  "test",
		EntryPoints: []domain.EntryPoint{{
			Phase:   "first",
			Accepts: []string{"user_prompt", "first"},
		}},
		Phases: []domain.WorkflowPhase{{
			ID:       "first",
			Produces: []string{"first"},
			Approval: domain.ApprovalNone,
		}},
		Artifacts: map[string]domain.WorkflowArtifact{
			"first": {Path: "artifacts/first.md", Template: "first"},
		},
	}
}
