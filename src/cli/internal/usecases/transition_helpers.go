package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/ports"
)

func loadWorkItemAndWorkflow(
	baseDir, workItemID string,
	workItemRepo ports.WorkItemRepository,
	workflowRepo ports.WorkflowRepository,
) (*domain.WorkItem, *domain.Workflow, error) {
	item, err := workItemRepo.GetWorkItem(baseDir, workItemID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get work item: %w", err)
	}

	workflow, err := workflowRepo.GetWorkflow(baseDir, item.Workflow.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return item, workflow, nil
}

func createArtifactsForTransitions(
	baseDir string,
	item *domain.WorkItem,
	workflow *domain.Workflow,
	transitions []domain.PhaseTransition,
) error {
	artifactManager := infra.NewArtifactManager()
	templateVars := map[string]string{
		"title":      item.Title,
		"id":         item.ID,
		"created_at": item.CreatedAt,
		"type":       item.Type,
	}

	for _, transition := range transitions {
		if transition.To != domain.PhaseReady {
			continue
		}
		if err := artifactManager.CreateArtifactsForPhase(baseDir, workflow, transition.Phase, item.ID, templateVars); err != nil {
			return fmt.Errorf("failed to create artifacts for phase %s: %w", transition.Phase, err)
		}
	}

	return nil
}
