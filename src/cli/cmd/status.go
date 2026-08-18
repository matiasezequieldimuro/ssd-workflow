package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var statusCmd = &cobra.Command{
	Use:   "status <work-item-id>",
	Short: "Get work item status and phase details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		repo := infra.NewFSWorkItemRepository()
		uc := usecases.NewStatusUseCase(repo)

		item, err := uc.Execute(targetDir, id)

		outputResult(item, err, func() {
			fmt.Printf("Work Item: %s [%s]\n", item.ID, item.Status)
			fmt.Printf("Title: %s\n", item.Title)
			fmt.Printf("Workflow: %s\n", item.Workflow.ID)
			fmt.Println("-------------------------------------------------------------")
			fmt.Printf("%-20s %-20s %-20s\n", "PHASE", "STATUS", "ARTIFACT")
			fmt.Println("-------------------------------------------------------------")
			for ph, st := range item.Phases {
				fmt.Printf("%-20s %-20s %-20s\n", ph, st.Status, st.Artifact)
			}
		})
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
