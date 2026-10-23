package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"

	"github.com/hdirksor/specdeck/internal/update"
	"github.com/spf13/cobra"
)

// version is set by goreleaser at link time.
var version string

var currentBuild = detectBuild()

// Where notices and errors go; tests replace these.
var (
	stderr           io.Writer = os.Stderr
	stderrIsTerminal           = func() bool { return isTerminal(os.Stderr) }
)

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
	os.Exit(execute())
}

// execute runs the command line, prints any notices, and returns the exit code.
func execute() int {
	executed, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(stderr, err)
	}
	notifyStaleCommands(executed)
	notifyUpdates(executed)
	if err != nil {
		return 1
	}
	return 0
}
