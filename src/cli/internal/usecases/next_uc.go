package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type NextAction struct {
	PhaseID       string `json:"phase_id" yaml:"phase_id"`
	Status        string `json:"status" yaml:"status"`
	Procedure     string `json:"procedure,omitempty" yaml:"procedure,omitempty"`
	Artifact      string `json:"artifact,omitempty" yaml:"artifact,omitempty"`
	NeedsApproval bool   `json:"needs_approval" yaml:"needs_approval"`
	Message       string `json:"message" yaml:"message"`
}

type NextUseCase struct {
	workItemRepo ports.WorkItemRepository
	workflowRepo ports.WorkflowRepository
}

func NewNextUseCase(wiRepo ports.WorkItemRepository, wfRepo ports.WorkflowRepository) *NextUseCase {
	return &NextUseCase{
		workItemRepo: wiRepo,
		workflowRepo: wfRepo,
	}
}

func (uc *NextUseCase) Execute(baseDir, id string) (*NextAction, error) {
	item, err := uc.workItemRepo.GetWorkItem(baseDir, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get work item: %w", err)
	}

	wf, err := uc.workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	// Look for phases awaiting approval
	for _, ph := range wf.Phases {
		state, exists := item.Phases[ph.ID]
		if exists && state.Status == "awaiting_approval" {
			return &NextAction{
				PhaseID:       ph.ID,
				Status:        state.Status,
				Procedure:     ph.Procedure,
				Artifact:      state.Artifact,
				NeedsApproval: true,
				Message:       fmt.Sprintf("Phase '%s' is awaiting human approval before proceeding.", ph.ID),
			}, nil
		}
	}

	// Look for phases in_progress or ready
	for _, ph := range wf.Phases {
		state, exists := item.Phases[ph.ID]
		if exists && (state.Status == "ready" || state.Status == "in_progress") {
			// Si la fase está lista (ready), cambiar a in_progress
			if state.Status == "ready" {
				// Cambiar fase actual de ready → in_progress
				item.Phases[ph.ID] = domain.PhaseState{
					Status:   "in_progress",
					Artifact: state.Artifact,
				}

				// Guardar el item actualizado
				if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
					return nil, fmt.Errorf("failed to update work item: %w", err)
				}
			}

			return &NextAction{
				PhaseID:       ph.ID,
				Status:        state.Status,
				Procedure:     ph.Procedure,
				Artifact:      state.Artifact,
				NeedsApproval: ph.Approval == "required",
				Message:       fmt.Sprintf("Next active phase is '%s' (in_progress). Follow procedure '%s'.", ph.ID, ph.Procedure),
			}, nil
		}
	}

	return &NextAction{
		Message: "No active phases pending. Work item may be completed or archived.",
	}, nil
}
