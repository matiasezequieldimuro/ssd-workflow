package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	beginPhaseID   string
	beginActorKind string
	beginActorID   string
)

var beginCmd = &cobra.Command{
	Use:   "begin <work-item-id>",
	Short: "Begin a ready, rejected, or superseded phase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workItemRepo := infra.NewFSWorkItemRepository()
		workflowRepo := infra.NewFSWorkflowRepository()
		uc := usecases.NewBeginPhaseUseCase(workItemRepo, workflowRepo)

		item, err := uc.Execute(targetDir, usecases.BeginPhaseInput{
			WorkItemID: args[0],
			PhaseID:    beginPhaseID,
			Actor: domain.Actor{
				Kind: domain.ActorKind(beginActorKind),
				ID:   beginActorID,
			},
		})

		outputResult(item, err, func() {
			fmt.Printf("Phase '%s' started for work item '%s'.\n", beginPhaseID, args[0])
		})
	},
}

func init() {
	beginCmd.Flags().StringVarP(&beginPhaseID, "phase", "p", "", "Phase ID to begin")
	beginCmd.Flags().StringVar(&beginActorKind, "actor-kind", "agent", "Actor kind (human, agent, cli, system)")
	beginCmd.Flags().StringVar(&beginActorID, "actor-id", "agent", "Actor ID")
	_ = beginCmd.MarkFlagRequired("phase")
	RootCmd.AddCommand(beginCmd)
}
