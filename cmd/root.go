package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/hdirksor/specdeck/internal/update"
	"github.com/spf13/cobra"
)

// version is set by goreleaser at link time.
var version string

var currentBuild = detectBuild()

var rootCmd = &cobra.Command{
	Use:           "specdeck",
	Short:         "Manage and export project specifications",
	Long:          `specdeck manages a local directory of specifications and builds complete export artifacts from them.`,
	SilenceErrors: true,
	SilenceUsage:  true,
	Version:       currentBuild.Version,
}

func detectBuild() update.Build {
	info, _ := debug.ReadBuildInfo()
	return update.DetectBuild(version, info)
}

func Execute() {
	executed, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	notifyStaleCommands(executed)
	notifyUpdates(executed)
	if err != nil {
		os.Exit(1)
	}
}
