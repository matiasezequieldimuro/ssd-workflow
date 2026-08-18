package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	targetDir  string
)

var RootCmd = &cobra.Command{
	Use:   "sdd",
	Short: "SDD Engine CLI - Spec-Driven Development Framework Engine",
	Long:  "CLI tool for managing Spec-Driven Development workflows, work items, phase state transitions, and event tracking.",
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output results in JSON format")
	RootCmd.PersistentFlags().StringVar(&targetDir, "dir", ".", "Target project directory path")
}

type JSONResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *JSONError  `json:"error,omitempty"`
}

type JSONError struct {
	Message string `json:"message"`
}

func outputResult(data interface{}, err error, textPrinter func()) {
	if jsonOutput {
		resp := JSONResponse{Success: err == nil}
		if err != nil {
			resp.Error = &JSONError{Message: err.Error()}
		} else {
			resp.Data = data
		}

		out, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(out))
		if err != nil {
			os.Exit(1)
		}
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if textPrinter != nil {
		textPrinter()
	}
}
