package cmd

import (
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

type Application struct {
	Init        *usecases.InitUseCase
	Start       *usecases.StartWorkItemUseCase
	Status      *usecases.StatusUseCase
	Next        *usecases.NextUseCase
	Begin       *usecases.BeginPhaseUseCase
	Deliver     *usecases.DeliverPhaseUseCase
	Approve     *usecases.ApproveUseCase
	Complete    *usecases.CompleteUseCase
	RecordEvent *usecases.RecordEventUseCase
}

func NewProductionApplication() Application {
	workItems := infra.NewFSWorkItemRepository()
	workflows := infra.NewFSWorkflowRepository()
	config := infra.NewFSConfigRepository()
	artifacts := infra.NewArtifactManager()
	clock := infra.NewSystemClock()
	ids := infra.NewCryptoIDGenerator()

	return Application{
		Init: usecases.NewInitUseCase(
			infra.NewFSProjectInitializer(),
		),
		Start: usecases.NewStartWorkItemUseCase(
			workItems,
			workflows,
			config,
			artifacts,
			clock,
			ids,
		),
		Status: usecases.NewStatusUseCase(workItems, workflows),
		Next:   usecases.NewNextUseCase(workItems, workflows),
		Begin: usecases.NewBeginPhaseUseCase(
			workItems,
			workflows,
			clock,
			ids,
		),
		Deliver: usecases.NewDeliverPhaseUseCase(
			workItems,
			workflows,
			artifacts,
			clock,
			ids,
		),
		Approve: usecases.NewApproveUseCase(
			workItems,
			workflows,
			artifacts,
			clock,
			ids,
		),
		Complete: usecases.NewCompleteUseCase(
			workItems,
			workflows,
			artifacts,
			clock,
			ids,
		),
		RecordEvent: usecases.NewRecordEventUseCase(
			workItems,
			clock,
			ids,
		),
	}
}
