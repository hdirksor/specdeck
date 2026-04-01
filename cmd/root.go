package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "specdeck",
	Short: "Manage and export project specifications",
	Long:  `specdeck manages a local directory of specifications and builds complete export artifacts from them.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
