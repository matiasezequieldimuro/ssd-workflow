package usecases

import (
	"fmt"
	"time"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
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

	phaseState, exists := item.Phases[in.PhaseID]
	if !exists {
		return nil, domain.ErrPhaseNotFound
	}

	// Also allow approving when in_progress if artifact exists (for CLI flexibility)
	if phaseState.Status != "awaiting_approval" && phaseState.Status != "in_progress" {
		return nil, fmt.Errorf("%w: current status is %s", domain.ErrPhaseNotAwaitingApproval, phaseState.Status)
	}

	// Update phase status
	phaseState.Status = "approved"
	item.Phases[in.PhaseID] = phaseState

	// Append approval record
	approvalRecord := domain.Approval{
		Phase:  in.PhaseID,
		Status: "approved",
		By:     &in.ApprovedBy,
		At:     time.Now().UTC().Format(time.RFC3339),
		Comment: in.Comment,
	}
	item.Approvals = append(item.Approvals, approvalRecord)

	// Load workflow to unblock dependent phases
	wf, err := uc.workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
	if err == nil {
		unblockDependentPhases(item, wf)
	}

	if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
		return nil, fmt.Errorf("failed to save work item: %w", err)
	}

	// Create artifacts for newly unblocked phases
	if err == nil && wf != nil {
		artifactMgr := infra.NewArtifactManager()
		templateVars := map[string]string{
			"title":      item.Title,
			"id":         item.ID,
			"created_at": item.CreatedAt,
			"type":       item.Type,
		}

		for _, ph := range wf.Phases {
			if state, exists := item.Phases[ph.ID]; exists && state.Status == "ready" && state.Artifact != "" {
				// Try to create artifacts for this phase if not already done
				_ = artifactMgr.CreateArtifactsForPhase(baseDir, wf, ph.ID, in.WorkItemID, templateVars)
			}
		}
	}

	event := domain.NewEvent(in.WorkItemID, "approval.recorded", in.ApprovedBy, map[string]interface{}{
		"phase":   in.PhaseID,
		"status":  "approved",
		"comment": in.Comment,
	})
	_ = uc.workItemRepo.AppendEvent(baseDir, in.WorkItemID, event)

	return item, nil
}

func unblockDependentPhases(item *domain.WorkItem, wf *domain.Workflow) {
	for _, ph := range wf.Phases {
		currentState, exists := item.Phases[ph.ID]
		if !exists || currentState.Status != "blocked" {
			continue
		}

		// Check if all requires are satisfied
		allRequiresSatisfied := true
		for _, reqPhaseID := range ph.Requires {
			reqState, reqExists := item.Phases[reqPhaseID]
			if !reqExists {
				allRequiresSatisfied = false
				break
			}
			s := reqState.Status
			if s != "approved" && s != "completed" && s != "accepted" && s != "not_applicable" {
				allRequiresSatisfied = false
				break
			}
		}

		if allRequiresSatisfied {
			artPath := fmt.Sprintf("artifacts/%s.md", ph.ID)
			item.Phases[ph.ID] = domain.PhaseState{
				Status:   "ready",
				Artifact: artPath,
			}
		}
	}
}
