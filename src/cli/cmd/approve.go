package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	approvePhase   string
	approveBy      string
	approveComment string
)

var approveCmd = &cobra.Command{
	Use:   "approve <work-item-id>",
	Short: "Record human approval for a phase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wiRepo := infra.NewFSWorkItemRepository()
		wfRepo := infra.NewFSWorkflowRepository()
		uc := usecases.NewApproveUseCase(wiRepo, wfRepo)

		actor := domain.Actor{
			Kind: "human",
			ID:   approveBy,
		}

		item, err := uc.Execute(targetDir, usecases.ApproveInput{
			WorkItemID: id,
			PhaseID:    approvePhase,
			ApprovedBy: actor,
			Comment:    approveComment,
		})

		outputResult(item, err, func() {
			fmt.Printf("Successfully approved phase '%s' for work item '%s' by %s.\n", approvePhase, id, approveBy)
		})
	},
}

func init() {
	approveCmd.Flags().StringVarP(&approvePhase, "phase", "p", "", "Phase ID to approve")
	approveCmd.Flags().StringVarP(&approveBy, "by", "b", "human", "Human approver ID")
	approveCmd.Flags().StringVarP(&approveComment, "comment", "c", "", "Optional approval comment")

	_ = approveCmd.MarkFlagRequired("phase")
	RootCmd.AddCommand(approveCmd)
}
