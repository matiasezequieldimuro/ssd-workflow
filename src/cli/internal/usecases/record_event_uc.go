package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type RecordEventInput struct {
	WorkItemID string
	EventType  string
	Message    string
	Actor      domain.Actor
	Data       map[string]interface{}
}

type RecordEventUseCase struct {
	workItemRepo ports.WorkItemRepository
}

func NewRecordEventUseCase(repo ports.WorkItemRepository) *RecordEventUseCase {
	return &RecordEventUseCase{workItemRepo: repo}
}

func (uc *RecordEventUseCase) Execute(baseDir string, in RecordEventInput) error {
	if !uc.workItemRepo.WorkItemExists(baseDir, in.WorkItemID) {
		return domain.ErrWorkItemNotFound
	}

	if in.Data == nil {
		in.Data = make(map[string]interface{})
	}
	if in.Message != "" {
		in.Data["message"] = in.Message
	}

	event := domain.NewEvent(in.WorkItemID, in.EventType, in.Actor, in.Data)
	if err := uc.workItemRepo.AppendEvent(baseDir, in.WorkItemID, event); err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	return nil
}
