package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	completePhaseID   string
	completeActorKind string
	completeActorID   string
)

var completeCmd = &cobra.Command{
	Use:   "complete <work-item-id>",
	Short: "Complete an approved phase or the work item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workItemRepo := infra.NewFSWorkItemRepository()
		workflowRepo := infra.NewFSWorkflowRepository()
		uc := usecases.NewCompleteUseCase(workItemRepo, workflowRepo)

		item, err := uc.Execute(targetDir, usecases.CompleteInput{
			WorkItemID: args[0],
			PhaseID:    completePhaseID,
			Actor: domain.Actor{
				Kind: domain.ActorKind(completeActorKind),
				ID:   completeActorID,
			},
		})

		outputResult(item, err, func() {
			if completePhaseID == "" {
				fmt.Printf("Work item '%s' completed.\n", args[0])
				return
			}
			fmt.Printf("Phase '%s' completed for work item '%s'.\n", completePhaseID, args[0])
		})
	},
}

func init() {
	completeCmd.Flags().StringVarP(&completePhaseID, "phase", "p", "", "Approved or accepted phase to complete; omit to complete the work item")
	completeCmd.Flags().StringVar(&completeActorKind, "actor-kind", "cli", "Actor kind (human, agent, cli, system)")
	completeCmd.Flags().StringVar(&completeActorID, "actor-id", "sdd", "Actor ID")
	RootCmd.AddCommand(completeCmd)
}
