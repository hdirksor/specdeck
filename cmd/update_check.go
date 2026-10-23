package cmd

import (
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hdirksor/specdeck/internal/update"
	"github.com/spf13/cobra"
)

var updateCheckCmd = &cobra.Command{
	Use:    "__update-check",
	Short:  "Refresh the cached latest-release information",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE:   runUpdateCheck,
}

// The release API and the background refresh; tests replace these.
var (
	latestReleaseURL = update.LatestReleaseURL
	startRefresh     = startUpdateCheck
)

func init() {
	rootCmd.AddCommand(updateCheckCmd)
}

func runUpdateCheck(cmd *cobra.Command, args []string) error {
	path, err := updateCachePath()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), update.Timeout)
	defer cancel()
	return update.Refresh(ctx, http.DefaultClient, latestReleaseURL, path)
}

// notifyUpdates prints a new-release notice and refreshes the release cache
// in the background. The update check must never break a command, so its
// failures are dropped.
func notifyUpdates(executed *cobra.Command) {
	if skipUpdateCheck(executed) || !update.Enabled(currentBuild, os.Getenv, stderrIsTerminal()) {
		return
	}
	path, err := updateCachePath()
	if err != nil {
		return
	}
	_ = update.Notify(stderr, currentBuild, path, time.Now(), startRefresh)
}

// skipUpdateCheck reports whether executed is a command that should never
// show a notice: help, --version, shell completion, or the refresh itself.
func skipUpdateCheck(executed *cobra.Command) bool {
	if executed == nil {
		return true
	}
	for _, name := range []string{"help", "version"} {
		if flag := executed.Flags().Lookup(name); flag != nil && flag.Changed {
			return true
		}
	}
	for c := executed; c != nil; c = c.Parent() {
		switch c.Name() {
		case "help", "completion", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd, updateCheckCmd.Name():
			return true
		}
	}
	return false
}

// startUpdateCheck runs `specdeck __update-check` as a separate process so
// the current command can exit without waiting for the network.
func startUpdateCheck() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	child := exec.Command(exe, updateCheckCmd.Name())
	if err := child.Start(); err != nil {
		return err
	}
	return child.Process.Release()
}

func updateCachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "specdeck", "update-check.json"), nil
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
