package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var nextCmd = &cobra.Command{
	Use:   "next <work-item-id>",
	Short: "Get the next active phase and required procedure",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wiRepo := infra.NewFSWorkItemRepository()
		wfRepo := infra.NewFSWorkflowRepository()
		uc := usecases.NewNextUseCase(wiRepo, wfRepo)

		action, err := uc.Execute(targetDir, id)

		outputResult(action, err, func() {
			fmt.Printf("%s\n", action.Message)
			if action.PhaseID != "" {
				fmt.Printf("Phase: %s (Status: %s)\n", action.PhaseID, action.Status)
				fmt.Printf("Procedure: %s\n", action.Procedure)
				fmt.Printf("Artifact: %s\n", action.Artifact)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(nextCmd)
}
