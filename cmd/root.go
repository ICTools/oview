package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "oview",
	Short: "oview - Local Software Factory Environment Manager",
	Long: `oview is a CLI tool that bootstraps a local Software Factory environment
for multiple projects with shared infrastructure (Postgres+pgvector, n8n).`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// CheckError prints error and exits
func CheckError(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
