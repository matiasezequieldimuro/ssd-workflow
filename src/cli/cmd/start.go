package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	startWorkflow string
	startTitle    string
	startSummary  string
	startFromArt  string
	startPhase    string
	actorKind     string
	actorID       string
)

var startCmd = &cobra.Command{
	Use:   "start <work-item-id>",
	Short: "Start a new SDD work item",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wiRepo := infra.NewFSWorkItemRepository()
		wfRepo := infra.NewFSWorkflowRepository()
		configRepo := infra.NewFSConfigRepository()
		uc := usecases.NewStartWorkItemUseCase(wiRepo, wfRepo, configRepo)

		actor := domain.Actor{
			Kind: domain.ActorKind(actorKind),
			ID:   actorID,
		}

		item, err := uc.Execute(targetDir, usecases.StartWorkItemInput{
			ID:           id,
			WorkflowID:   startWorkflow,
			Title:        startTitle,
			Summary:      startSummary,
			FromArtifact: startFromArt,
			Phase:        startPhase,
			Actor:        actor,
		})

		outputResult(item, err, func() {
			fmt.Printf("Work item '%s' successfully started with workflow '%s'.\n", item.ID, item.Workflow.ID)
		})
	},
}

func init() {
	startCmd.Flags().StringVarP(&startWorkflow, "workflow", "w", "", "Workflow ID; defaults to .sdd/config.yaml")
	startCmd.Flags().StringVarP(&startTitle, "title", "t", "", "Work item title")
	startCmd.Flags().StringVarP(&startSummary, "summary", "s", "", "Initial input summary")
	startCmd.Flags().StringVar(&startFromArt, "from-artifact", "", "Path to pre-existing artifact")
	startCmd.Flags().StringVar(&startPhase, "phase", "", "Entry phase when using --from-artifact")
	startCmd.Flags().StringVar(&actorKind, "actor-kind", "human", "Actor kind (human, agent, cli)")
	startCmd.Flags().StringVar(&actorID, "actor-id", "user", "Actor ID")

	_ = startCmd.MarkFlagRequired("title")
	RootCmd.AddCommand(startCmd)
}
