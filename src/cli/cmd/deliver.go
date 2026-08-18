package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	deliverPhaseID         string
	deliverRequestApproval bool
	deliverActorKind       string
	deliverActorID         string
)

var deliverCmd = &cobra.Command{
	Use:   "deliver <work-item-id>",
	Short: "Deliver the output of an in-progress phase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workItemRepo := infra.NewFSWorkItemRepository()
		workflowRepo := infra.NewFSWorkflowRepository()
		uc := usecases.NewDeliverPhaseUseCase(workItemRepo, workflowRepo)

		item, err := uc.Execute(targetDir, usecases.DeliverPhaseInput{
			WorkItemID:              args[0],
			PhaseID:                 deliverPhaseID,
			RequestOptionalApproval: deliverRequestApproval,
			Actor: domain.Actor{
				Kind: domain.ActorKind(deliverActorKind),
				ID:   deliverActorID,
			},
		})

		outputResult(item, err, func() {
			fmt.Printf("Phase '%s' delivered for work item '%s'.\n", deliverPhaseID, args[0])
		})
	},
}

func init() {
	deliverCmd.Flags().StringVarP(&deliverPhaseID, "phase", "p", "", "Phase ID to deliver")
	deliverCmd.Flags().BoolVar(&deliverRequestApproval, "request-approval", false, "Request human approval for an optional gate")
	deliverCmd.Flags().StringVar(&deliverActorKind, "actor-kind", "agent", "Actor kind (human, agent, cli, system)")
	deliverCmd.Flags().StringVar(&deliverActorID, "actor-id", "agent", "Actor ID")
	_ = deliverCmd.MarkFlagRequired("phase")
	RootCmd.AddCommand(deliverCmd)
}
