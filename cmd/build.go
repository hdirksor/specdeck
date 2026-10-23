package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdirksor/specdeck/internal/build"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Resolve container refs and write specs to dist/",
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

	containers, err := build.Load(filepath.Join(root, "containers"))
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	if err := build.Validate(containers); err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}

	written, err := build.WriteDist(containers, filepath.Join(root, "dist"))
	for _, path := range written {
		fmt.Fprintf(cmd.OutOrStdout(), "\033[32m↳\033[0m  dist/%s\n", path)
	}
	if err != nil {
		return fmt.Errorf("\U0001F7E5  %w", err)
	}
	return nil
}
