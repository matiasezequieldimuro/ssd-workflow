package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type CompleteInput struct {
	WorkItemID  string
	PhaseID     string
	Actor       domain.Actor
	OperationID string
}

type CompleteUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewCompleteUseCase(workItemRepo ports.WorkItemRepository, workflowRepo ports.WorkflowRepository) *CompleteUseCase {
	return &CompleteUseCase{workItemRepo: workItemRepo, workflowRepo: workflowRepo}
}

func (uc *CompleteUseCase) Execute(baseDir string, input CompleteInput) (*domain.WorkItem, error) {
	if err := domain.ValidateActor(input.Actor); err != nil {
		return nil, err
	}
	item, workflow, err := loadWorkItemAndWorkflow(baseDir, input.WorkItemID, uc.workItemRepo, uc.workflowRepo)
	if err != nil {
		return nil, err
	}
	applied, err := operationApplied(baseDir, input.WorkItemID, input.OperationID, uc.workItemRepo)
	if err != nil {
		return nil, err
	}
	if applied {
		return item, nil
	}

	eventData := map[string]interface{}{}
	eventType := "work_item.completed"
	var artifacts []ports.ArtifactWrite

	if input.PhaseID == "" {
		if err := item.Complete(workflow); err != nil {
			return nil, err
		}
	} else {
		mutation, err := item.CompletePhase(workflow, input.PhaseID)
		if err != nil {
			return nil, err
		}
		artifacts, err = prepareArtifactsForTransitions(baseDir, item, workflow, mutation.Unblocked)
		if err != nil {
			return nil, err
		}
		eventType = "phase.transitioned"
		eventData = map[string]interface{}{
			"phase": input.PhaseID,
			"from":  mutation.Transition.From,
			"to":    mutation.Transition.To,
		}
	}

	event := newOperationEvent(input.WorkItemID, eventType, input.Actor, eventData, input.OperationID)
	persisted, err := commitWorkItem(baseDir, uc.workItemRepo, item, artifacts, []domain.Event{event}, input.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to commit completion: %w", err)
	}

	return persisted, nil
}
