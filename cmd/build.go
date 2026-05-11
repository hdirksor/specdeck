package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/spec"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Resolve container refs and write flat specs to dist/",
	Args:  cobra.NoArgs,
	RunE:  runBuild,
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func runBuild(cmd *cobra.Command, args []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}

	containersRoot := filepath.Join(root, "containers")
	containers, err := spec.LoadContainers(containersRoot)
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	byPath := make(map[string]spec.Container, len(containers))
	for _, c := range containers {
		byPath[c.Path] = c
	}

	distRoot := filepath.Join(root, "dist")

	for _, c := range containers {
		ownStates, imports := spec.ResolveContainerSections(c, containersRoot, byPath)
		outPath := filepath.Join(distRoot, c.Path)
		if err := spec.WriteBuiltContainer(outPath, c, ownStates, imports); err != nil {
			return fmt.Errorf("\U0001F7E5  %s: %w", c.Path, err)
		}
		fmt.Printf("\U0001F7E2  dist/%s\n", c.Path)
	}

	return nil
}
