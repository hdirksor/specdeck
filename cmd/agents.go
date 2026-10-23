package cmd

import (
	"fmt"
	"os"

	"github.com/hdirksor/specdeck/internal/agents"
	"github.com/spf13/cobra"
)

var agentsCmd = &cobra.Command{
	Use:   "agents [agent...]",
	Short: "Install or update specdeck's commands for AI coding agents",
	Long: `Installs or updates specdeck's commands for the named agents (supported: claude) and records them in specdeck.yml.
With no names, updates every agent already recorded. Run it after upgrading specdeck.`,
	RunE: runAgents,
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}

func runAgents(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	installed, err := agents.Update(dir, args)
	if err != nil {
		return err
	}
	for _, a := range installed {
		fmt.Fprintf(cmd.OutOrStdout(), "Updated %s commands in %s/\n", a.Name, a.Dir)
	}
	return nil
}
