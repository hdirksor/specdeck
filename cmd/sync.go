package cmd

import (
	"fmt"
	"os"

	"github.com/hdickson/specdeck/internal/link"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Update Claude Code skills to the current specdeck version",
	Long:  `Re-writes all Claude Code skill files without changing specdeck.yml configuration.`,
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := link.Sync(dir); err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Skills updated.")
	return nil
}
