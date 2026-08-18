package usecases

import (
	"fmt"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/ports"
)

type StatusUseCase struct {
	workItemRepo ports.WorkItemRepository
}

func NewStatusUseCase(repo ports.WorkItemRepository) *StatusUseCase {
	return &StatusUseCase{workItemRepo: repo}
}

func (uc *StatusUseCase) Execute(baseDir, id string) (*domain.WorkItem, error) {
	item, err := uc.workItemRepo.GetWorkItem(baseDir, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get status for work item %s: %w", id, err)
	}
	return item, nil
}
