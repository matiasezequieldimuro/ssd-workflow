package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type RecordEventInput struct {
	WorkItemID  string
	EventType   string
	Message     string
	Actor       domain.Actor
	Data        map[string]interface{}
	OperationID string
}

type RecordEventUseCase struct {
	workItemRepo ports.WorkItemRepository
}

func NewRecordEventUseCase(repo ports.WorkItemRepository) *RecordEventUseCase {
	return &RecordEventUseCase{workItemRepo: repo}
}

func (uc *RecordEventUseCase) Execute(baseDir string, in RecordEventInput) error {
	if err := domain.ValidateActor(in.Actor); err != nil {
		return err
	}
	item, err := uc.workItemRepo.GetWorkItem(baseDir, in.WorkItemID)
	if err != nil {
		return err
	}
	applied, err := operationApplied(baseDir, in.WorkItemID, in.OperationID, uc.workItemRepo)
	if err != nil {
		return err
	}
	if applied {
		return nil
	}

	if in.Data == nil {
		in.Data = make(map[string]interface{})
	}
	if in.Message != "" {
		in.Data["message"] = in.Message
	}

	event := newOperationEvent(in.WorkItemID, in.EventType, in.Actor, in.Data, in.OperationID)
	if _, err := commitWorkItem(baseDir, uc.workItemRepo, item, nil, []domain.Event{event}, in.OperationID); err != nil {
		return fmt.Errorf("failed to commit event: %w", err)
	}

	return nil
}
