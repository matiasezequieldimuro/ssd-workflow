package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type DeliverPhaseInput struct {
	WorkItemID              string
	PhaseID                 string
	RequestOptionalApproval bool
	Actor                   domain.Actor
}

type DeliverPhaseUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewDeliverPhaseUseCase(workItemRepo ports.WorkItemRepository, workflowRepo ports.WorkflowRepository) *DeliverPhaseUseCase {
	return &DeliverPhaseUseCase{workItemRepo: workItemRepo, workflowRepo: workflowRepo}
}

func (uc *DeliverPhaseUseCase) Execute(baseDir string, input DeliverPhaseInput) (*domain.WorkItem, error) {
	item, workflow, err := loadWorkItemAndWorkflow(baseDir, input.WorkItemID, uc.workItemRepo, uc.workflowRepo)
	if err != nil {
		return nil, err
	}

	mutation, err := item.DeliverPhase(workflow, input.PhaseID, input.RequestOptionalApproval)
	if err != nil {
		return nil, err
	}
	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}
	if err := createArtifactsForTransitions(baseDir, item, workflow, mutation.Unblocked); err != nil {
		return nil, err
	}

	event := domain.NewEvent(input.WorkItemID, "phase.transitioned", input.Actor, map[string]interface{}{
		"phase": input.PhaseID,
		"from":  mutation.Transition.From,
		"to":    mutation.Transition.To,
	})
	if err := uc.workItemRepo.AppendEvent(baseDir, input.WorkItemID, event); err != nil {
		return nil, fmt.Errorf("failed to append phase transition event: %w", err)
	}

	if mutation.Transition.To == domain.PhaseAwaitingApproval {
		approvalEvent := domain.NewEvent(input.WorkItemID, "approval.requested", input.Actor, map[string]interface{}{
			"phase": input.PhaseID,
		})
		if err := uc.workItemRepo.AppendEvent(baseDir, input.WorkItemID, approvalEvent); err != nil {
			return nil, fmt.Errorf("failed to append approval request event: %w", err)
		}
	}

	return item, nil
}
