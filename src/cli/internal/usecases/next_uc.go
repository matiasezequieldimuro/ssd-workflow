package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
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
			// Si la fase está lista (ready), buscar la siguiente y desbloquearlo
			if state.Status == "ready" {
				// Buscar el índice de la fase actual
				currentIdx := -1
				for i, p := range wf.Phases {
					if p.ID == ph.ID {
						currentIdx = i
						break
					}
				}

				// Si hay una fase siguiente, desbloquearla y crear sus artifacts
				if currentIdx >= 0 && currentIdx+1 < len(wf.Phases) {
					nextPhase := wf.Phases[currentIdx+1]
					item.Phases[nextPhase.ID] = domain.PhaseState{
						Status: "ready",
					}
					// Guardar el item actualizado
					if err := uc.workItemRepo.SaveWorkItem(baseDir, item); err != nil {
						return nil, fmt.Errorf("failed to update work item: %w", err)
					}

					// Crear artifacts para la siguiente fase
					artifactMgr := infra.NewArtifactManager()
					templateVars := map[string]string{
						"title":      item.Title,
						"id":         item.ID,
						"created_at": item.CreatedAt,
						"type":       item.Type,
					}
					if err := artifactMgr.CreateArtifactsForPhase(baseDir, wf, nextPhase.ID, item.ID, templateVars); err != nil {
						return nil, fmt.Errorf("failed to create artifacts for phase %s: %w", nextPhase.ID, err)
					}
				}
			}

			return &NextAction{
				PhaseID:       ph.ID,
				Status:        state.Status,
				Procedure:     ph.Procedure,
				Artifact:      state.Artifact,
				NeedsApproval: ph.Approval == "required",
				Message:       fmt.Sprintf("Next active phase is '%s' (%s). Follow procedure '%s'.", ph.ID, state.Status, ph.Procedure),
			}, nil
		}
	}

	return &NextAction{
		Message: "No active phases pending. Work item may be completed or archived.",
	}, nil
}
