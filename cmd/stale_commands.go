package cmd

import (
	"fmt"
	"os"

	"github.com/hdirksor/specdeck/internal/agents"
	"github.com/hdirksor/specdeck/internal/update"
	"github.com/spf13/cobra"
)

// notifyStaleCommands tells the user when the agent commands installed in the
// current directory are older than this version of specdeck.
func notifyStaleCommands(executed *cobra.Command) {
	if skipStaleCheck(executed) || !update.Interactive(os.Getenv, stderrIsTerminal()) {
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	fmt.Fprint(stderr, staleCommandsNotice(dir))
}

// skipStaleCheck reports whether executed should not check for stale
// commands: everything skipUpdateCheck skips, plus agents, which updates them.
func skipStaleCheck(executed *cobra.Command) bool {
	return skipUpdateCheck(executed) || executed == agentsCmd
}

// staleCommandsNotice returns the message to show for dir, or "" when no
// agents are recorded or their commands are current. Errors are dropped
// because the check must never break a command.
func staleCommandsNotice(dir string) string {
	stale, err := agents.Stale(dir)
	if err != nil || len(stale) == 0 {
		return ""
	}
	return "Agent commands are out of date — run `specdeck agents` to update.\n"
}
