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

	containerErrs := spec.ValidateImports(containersRoot, containers)
	for i := range containerErrs {
		containerErrs[i].File = filepath.Join("containers", containerErrs[i].File)
	}
	containerErrFiles := make(map[string]bool, len(containerErrs))
	for _, e := range containerErrs {
		containerErrFiles[e.File] = true
	}
	for _, c := range containers {
		path := filepath.Join("containers", c.Path)
		if !containerErrFiles[path] {
			fmt.Printf("\U0001F7E2  %s\n", path)
		}
	}
	allErrs = append(allErrs, containerErrs...)

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
		printValidationError(e)
	}

	if len(allErrs) > 0 {
		return fmt.Errorf("%d validation error(s)", len(allErrs))
	}
	return nil
}

func printValidationError(e spec.ValidationError) {
	loc := e.File
	if e.Line > 0 {
		loc = fmt.Sprintf("%s:%d", e.File, e.Line)
	}
	fmt.Fprintf(os.Stderr, "\U0001F7E5  %s  %s\n", loc, e.Message)
}
