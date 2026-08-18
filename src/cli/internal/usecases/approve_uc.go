package usecases

import (
	"fmt"
	"time"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type ApproveInput struct {
	WorkItemID string
	PhaseID    string
	ApprovedBy domain.Actor
	Comment    string
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

	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}

	if err := createArtifactsForTransitions(baseDir, item, wf, mutation.Unblocked); err != nil {
		return nil, err
	}

	event := domain.NewEvent(in.WorkItemID, "approval.recorded", in.ApprovedBy, map[string]interface{}{
		"phase":   in.PhaseID,
		"status":  domain.ApprovalApproved,
		"comment": in.Comment,
	})
	if err := uc.workItemRepo.AppendEvent(baseDir, in.WorkItemID, event); err != nil {
		return nil, fmt.Errorf("failed to append approval event: %w", err)
	}

	return item, nil
}
