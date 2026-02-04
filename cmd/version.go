package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of oview",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("oview version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
