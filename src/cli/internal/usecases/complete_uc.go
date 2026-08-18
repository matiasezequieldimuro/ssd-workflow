package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type CompleteInput struct {
	WorkItemID string
	PhaseID    string
	Actor      domain.Actor
}

type CompleteUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewCompleteUseCase(workItemRepo ports.WorkItemRepository, workflowRepo ports.WorkflowRepository) *CompleteUseCase {
	return &CompleteUseCase{workItemRepo: workItemRepo, workflowRepo: workflowRepo}
}

func (uc *CompleteUseCase) Execute(baseDir string, input CompleteInput) (*domain.WorkItem, error) {
	item, workflow, err := loadWorkItemAndWorkflow(baseDir, input.WorkItemID, uc.workItemRepo, uc.workflowRepo)
	if err != nil {
		return nil, err
	}

	eventData := map[string]interface{}{}
	eventType := "work_item.completed"

	if input.PhaseID == "" {
		if err := item.Complete(workflow); err != nil {
			return nil, err
		}
	} else {
		mutation, err := item.CompletePhase(workflow, input.PhaseID)
		if err != nil {
			return nil, err
		}
		if err := createArtifactsForTransitions(baseDir, item, workflow, mutation.Unblocked); err != nil {
			return nil, err
		}
		eventType = "phase.transitioned"
		eventData = map[string]interface{}{
			"phase": input.PhaseID,
			"from":  mutation.Transition.From,
			"to":    mutation.Transition.To,
		}
	}

	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}
	if err := uc.workItemRepo.AppendEvent(baseDir, input.WorkItemID, domain.NewEvent(input.WorkItemID, eventType, input.Actor, eventData)); err != nil {
		return nil, fmt.Errorf("failed to append completion event: %w", err)
	}

	return item, nil
}
