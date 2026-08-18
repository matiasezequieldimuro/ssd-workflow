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
	OperationID             string
}

type DeliverPhaseUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewDeliverPhaseUseCase(workItemRepo ports.WorkItemRepository, workflowRepo ports.WorkflowRepository) *DeliverPhaseUseCase {
	return &DeliverPhaseUseCase{workItemRepo: workItemRepo, workflowRepo: workflowRepo}
}

func (uc *DeliverPhaseUseCase) Execute(baseDir string, input DeliverPhaseInput) (*domain.WorkItem, error) {
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

	mutation, err := item.DeliverPhase(workflow, input.PhaseID, input.RequestOptionalApproval)
	if err != nil {
		return nil, err
	}
	artifacts, err := prepareArtifactsForTransitions(baseDir, item, workflow, mutation.Unblocked)
	if err != nil {
		return nil, err
	}

	events := []domain.Event{newOperationEvent(input.WorkItemID, "phase.transitioned", input.Actor, map[string]interface{}{
		"phase": input.PhaseID,
		"from":  mutation.Transition.From,
		"to":    mutation.Transition.To,
	}, input.OperationID)}

	if mutation.Transition.To == domain.PhaseAwaitingApproval {
		events = append(events, newOperationEvent(input.WorkItemID, "approval.requested", input.Actor, map[string]interface{}{
			"phase": input.PhaseID,
		}, input.OperationID))
	}

	persisted, err := commitWorkItem(baseDir, uc.workItemRepo, item, artifacts, events, input.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to commit phase delivery: %w", err)
	}
	return persisted, nil
}
