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

	for _, e := range containerErrs {
		printValidationError(e)
	}

	if len(containerErrs) > 0 {
		return fmt.Errorf("%d validation error(s)", len(containerErrs))
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
