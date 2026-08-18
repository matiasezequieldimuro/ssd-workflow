package usecases

import (
	"errors"
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

func prepareArtifactsForTransitions(
	baseDir string,
	item *domain.WorkItem,
	workflow *domain.Workflow,
	transitions []domain.PhaseTransition,
) ([]ports.ArtifactWrite, error) {
	artifactManager := infra.NewArtifactManager()
	templateVars := map[string]string{
		"title":      item.Title,
		"id":         item.ID,
		"created_at": item.CreatedAt,
		"type":       item.Type,
	}

	var writes []ports.ArtifactWrite
	for _, transition := range transitions {
		if transition.To != domain.PhaseReady {
			continue
		}
		phaseWrites, err := artifactManager.PrepareArtifactsForPhase(baseDir, workflow, transition.Phase, item.ID, templateVars)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare artifacts for phase %s: %w", transition.Phase, err)
		}
		writes = append(writes, phaseWrites...)
	}

	return writes, nil
}

func operationApplied(
	baseDir, workItemID, operationID string,
	repository ports.WorkItemRepository,
) (bool, error) {
	if err := domain.ValidateOperationID(operationID); err != nil {
		return false, err
	}
	if operationID == "" {
		return false, nil
	}
	applied, err := repository.OperationApplied(baseDir, workItemID, operationID)
	if err != nil {
		return false, fmt.Errorf("failed to check operation id: %w", err)
	}
	return applied, nil
}

func commitWorkItem(
	baseDir string,
	repository ports.WorkItemRepository,
	item *domain.WorkItem,
	artifacts []ports.ArtifactWrite,
	events []domain.Event,
	operationID string,
) (*domain.WorkItem, error) {
	err := repository.CommitWorkItem(baseDir, ports.WorkItemCommit{
		Item:        item,
		Artifacts:   artifacts,
		Events:      events,
		OperationID: operationID,
	})
	if errors.Is(err, domain.ErrOperationAlreadyApplied) {
		persisted, loadErr := repository.GetWorkItem(baseDir, item.ID)
		if loadErr != nil {
			return nil, errors.Join(err, fmt.Errorf("failed to load idempotent result: %w", loadErr))
		}
		return persisted, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func newOperationEvent(
	workItemID, eventType string,
	actor domain.Actor,
	data map[string]interface{},
	operationID string,
) domain.Event {
	event := domain.NewEvent(workItemID, eventType, actor, data)
	event.CorrelationID = operationID
	return event
}
