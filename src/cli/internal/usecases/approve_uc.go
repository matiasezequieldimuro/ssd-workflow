package usecases

import (
	"fmt"
	"time"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type ApproveInput struct {
	WorkItemID  string
	PhaseID     string
	ApprovedBy  domain.Actor
	Comment     string
	OperationID string
}

type ApproveUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewApproveUseCase(wiRepo ports.WorkItemRepository, wfRepo ports.WorkflowRepository) *ApproveUseCase {
	return &ApproveUseCase{
		workItemRepo: wiRepo,
		workflowRepo: wfRepo,
	}
}

func (uc *ApproveUseCase) Execute(baseDir string, in ApproveInput) (*domain.WorkItem, error) {
	item, err := uc.workItemRepo.GetWorkItem(baseDir, in.WorkItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to find work item: %w", err)
	}

	wf, err := uc.workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}
	applied, err := operationApplied(baseDir, in.WorkItemID, in.OperationID, uc.workItemRepo)
	if err != nil {
		return nil, err
	}
	if applied {
		return item, nil
	}

	mutation, err := item.ApprovePhase(
		wf,
		in.PhaseID,
		in.ApprovedBy,
		time.Now().UTC().Format(time.RFC3339),
		in.Comment,
	)
	if err != nil {
		return nil, err
	}

	artifacts, err := prepareArtifactsForTransitions(baseDir, item, wf, mutation.Unblocked)
	if err != nil {
		return nil, err
	}

	event := newOperationEvent(in.WorkItemID, "approval.recorded", in.ApprovedBy, map[string]interface{}{
		"phase":   in.PhaseID,
		"status":  domain.ApprovalApproved,
		"comment": in.Comment,
	}, in.OperationID)

	persisted, err := commitWorkItem(baseDir, uc.workItemRepo, item, artifacts, []domain.Event{event}, in.OperationID)
	if err != nil {
		return nil, fmt.Errorf("failed to commit approval: %w", err)
	}
	return persisted, nil
}
