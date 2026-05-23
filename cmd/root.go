package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cligpt",
	Short: "CLIGpt is a minimal terminal chatbot powered by OpenAI",
	Long:  "CLIGpt is a small Go CLI that streams AI chat responses in the terminal.",
}

// Execute is the entrypoint called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(chatCmd)
}