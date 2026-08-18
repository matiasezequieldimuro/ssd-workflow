package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type StatusPhase struct {
	ID       string
	Status   domain.PhaseStatus
	Artifact string
}

type StatusResult struct {
	*domain.WorkItem
	OrderedPhases []StatusPhase `json:"-" yaml:"-"`
}

type StatusUseCase struct {
	workItemRepo ports.WorkItemReader
	workflowRepo ports.WorkflowRepository
}

func NewStatusUseCase(repo ports.WorkItemReader, workflowRepo ports.WorkflowRepository) *StatusUseCase {
	return &StatusUseCase{
		workItemRepo: repo,
		workflowRepo: workflowRepo,
	}
}

func (uc *StatusUseCase) Execute(baseDir, id string) (*StatusResult, error) {
	item, err := uc.workItemRepo.GetWorkItem(baseDir, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get status for work item %s: %w", id, err)
	}
	workflow, err := uc.workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status workflow: %w", err)
	}
	ordered, err := workflow.OrderedPhases()
	if err != nil {
		return nil, err
	}
	phases := make([]StatusPhase, 0, len(ordered))
	for _, phase := range ordered {
		state := item.Phases[phase.ID]
		phases = append(phases, StatusPhase{
			ID:       phase.ID,
			Status:   state.Status,
			Artifact: state.Artifact,
		})
	}
	return &StatusResult{
		WorkItem:      item,
		OrderedPhases: phases,
	}, nil
}
