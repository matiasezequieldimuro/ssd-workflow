package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"sdd-cli/internal/usecases"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize .sdd framework structure in the project",
	Run: func(cmd *cobra.Command, args []string) {
		uc := usecases.NewInitUseCase()
		err := uc.Execute(targetDir)

		outputResult("Successfully initialized .sdd directory framework", err, func() {
			fmt.Printf("Successfully initialized .sdd directory framework in '%s'\n", targetDir)
		})
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
