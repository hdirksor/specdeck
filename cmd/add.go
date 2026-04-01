package cmd

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add containers or states to the project",
}

func init() {
	rootCmd.AddCommand(addCmd)
}
