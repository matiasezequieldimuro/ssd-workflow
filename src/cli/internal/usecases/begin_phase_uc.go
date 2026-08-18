package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type BeginPhaseInput struct {
	WorkItemID string
	PhaseID    string
	Actor      domain.Actor
}

type BeginPhaseUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewBeginPhaseUseCase(workItemRepo ports.WorkItemRepository, workflowRepo ports.WorkflowRepository) *BeginPhaseUseCase {
	return &BeginPhaseUseCase{workItemRepo: workItemRepo, workflowRepo: workflowRepo}
}

func (uc *BeginPhaseUseCase) Execute(baseDir string, input BeginPhaseInput) (*domain.WorkItem, error) {
	if err := domain.ValidateActor(input.Actor); err != nil {
		return nil, err
	}
	item, workflow, err := loadWorkItemAndWorkflow(baseDir, input.WorkItemID, uc.workItemRepo, uc.workflowRepo)
	if err != nil {
		return nil, err
	}

	mutation, err := item.BeginPhase(workflow, input.PhaseID)
	if err != nil {
		return nil, err
	}
	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}

	event := domain.NewEvent(input.WorkItemID, "phase.transitioned", input.Actor, map[string]interface{}{
		"phase": input.PhaseID,
		"from":  mutation.Transition.From,
		"to":    mutation.Transition.To,
	})
	if err := uc.workItemRepo.AppendEvent(baseDir, input.WorkItemID, event); err != nil {
		return nil, fmt.Errorf("failed to append phase transition event: %w", err)
	}

	return item, nil
}
