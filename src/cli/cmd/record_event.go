package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/domain"
	"sdd-cli/internal/infra"
	"sdd-cli/internal/usecases"
)

var (
	eventType        string
	eventMsg         string
	eventActKind     string
	eventActID       string
	eventOperationID string
)

var recordEventCmd = &cobra.Command{
	Use:   "record-event <work-item-id>",
	Short: "Append a custom event to the work item event log",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		wiRepo := infra.NewFSWorkItemRepository()
		uc := usecases.NewRecordEventUseCase(wiRepo)

		actor := domain.Actor{
			Kind: domain.ActorKind(eventActKind),
			ID:   eventActID,
		}

		err := uc.Execute(targetDir, usecases.RecordEventInput{
			WorkItemID:  id,
			EventType:   eventType,
			Message:     eventMsg,
			Actor:       actor,
			OperationID: eventOperationID,
		})

		outputResult("Event recorded successfully", err, func() {
			fmt.Printf("Successfully recorded event '%s' for work item '%s'.\n", eventType, id)
		})
	},
}

func init() {
	recordEventCmd.Flags().StringVarP(&eventType, "type", "t", "", "Event type")
	recordEventCmd.Flags().StringVarP(&eventMsg, "message", "m", "", "Event message description")
	recordEventCmd.Flags().StringVar(&eventActKind, "actor-kind", "agent", "Actor kind (human, agent, cli, system)")
	recordEventCmd.Flags().StringVar(&eventActID, "actor-id", "agent", "Actor ID")
	recordEventCmd.Flags().StringVar(&eventOperationID, "operation-id", "", "Stable idempotency key for safe retries")

	mustMarkFlagRequired(recordEventCmd, "type")
	RootCmd.AddCommand(recordEventCmd)
}
