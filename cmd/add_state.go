package cmd

import (
	"fmt"
	"os"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/spf13/cobra"
)

var addStateCmd = &cobra.Command{
	Use:   "state <name>",
	Short: "Propagate a new state into all existing leaf containers",
	Long: `Adds the named state to every leaf container that does not already have it,
copying the container's default specs as a starting point.

The state must already be defined in states/ before running this command.

Example: specdeck add state dark-mode`,
	Args: cobra.ExactArgs(1),
	RunE: runAddState,
}

func init() {
	addCmd.AddCommand(addStateCmd)
}

func runAddState(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	stateName := args[0]
	if err := scaffold.PropagateState(projectRoot, stateName); err != nil {
		return err
	}

	fmt.Printf("State %q propagated to all leaf containers.\n", stateName)
	return nil
}
