package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/spec"
	"github.com/spf13/cobra"
)

var generateStatesCmd = &cobra.Command{
	Use:   "states",
	Short: "Generate states from fact scopes",
	Long: `Generates states/states.lock from facts defined in states/facts/.

Facts with scope "cross" are fully cartesian-producted together.
Facts with scope "isolated" each contribute one state per non-default value.
Facts with scope "manual" (the default) are ignored.

Any state in states.lock with the same name as a manually defined state
will be overridden by the manual definition.`,
	RunE: runGenerateStates,
}

func init() {
	generateCmd.AddCommand(generateStatesCmd)
}

func runGenerateStates(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	facts, err := spec.LoadFacts(filepath.Join(projectRoot, "states", "facts"))
	if err != nil {
		return fmt.Errorf("loading facts: %w", err)
	}

	states := spec.GenerateStates(facts)
	if len(states) == 0 {
		fmt.Println("No states generated — set scope to 'cross' or 'isolated' on your facts.")
		return nil
	}

	outPath := filepath.Join(projectRoot, "states", "states.lock")
	if err := spec.WriteStates(outPath, states); err != nil {
		return err
	}

	fmt.Printf("Generated %d states → states/states.lock\n", len(states))
	for _, s := range states {
		fmt.Printf("  %s\n", s.Name)
	}
	return nil
}
