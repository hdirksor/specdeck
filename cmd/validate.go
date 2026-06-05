package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/spec"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate cross-references in the specdeck project",
	Args:  cobra.NoArgs,
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	var allErrs []spec.ValidationError

	factsDir := filepath.Join(root, "states", "facts")
	validFacts, factErrs := spec.ValidateFactFiles(factsDir)
	for _, name := range validFacts {
		fmt.Printf("\U0001F7E2  %s\n", filepath.Join("states", "facts", name))
	}
	for i := range factErrs {
		factErrs[i].File = filepath.Join("states", "facts", factErrs[i].File)
	}
	allErrs = append(allErrs, factErrs...)

	statesDir := filepath.Join(root, "states")
	validStates, stateErrs := spec.ValidateStateFiles(statesDir)
	for _, name := range validStates {
		fmt.Printf("\U0001F7E2  %s\n", filepath.Join("states", name))
	}
	for i := range stateErrs {
		stateErrs[i].File = filepath.Join("states", stateErrs[i].File)
	}
	allErrs = append(allErrs, stateErrs...)

	containersRoot := filepath.Join(root, "containers")
	containers, err := spec.LoadContainers(containersRoot)
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}
	for _, c := range containers {
		fmt.Printf("\U0001F7E2  %s\n", filepath.Join("containers", c.Path))
	}

	if len(allErrs) == 0 {
		facts, err := spec.LoadFacts(factsDir)
		if err != nil {
			return err
		}
		states, err := spec.LoadStates(statesDir)
		if err != nil {
			return err
		}
		allErrs = append(allErrs, spec.Validate(facts, states, containers)...)
	}

	for _, e := range allErrs {
		fmt.Fprintf(os.Stderr, "\U0001F7E5  %s  %s\n", e.File, e.Message)
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("%d validation error(s)", len(allErrs))
	}
	return nil
}
