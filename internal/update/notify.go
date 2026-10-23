package update

import (
	"fmt"
	"io"
	"time"
)

// Enabled reports whether release checks should run: only for released builds
// and only when Interactive.
func Enabled(build Build, getenv func(string) string, stderrIsTerminal bool) bool {
	return build.Method != MethodDev && Interactive(getenv, stderrIsTerminal)
}

// Interactive reports whether notices may be printed: only when a person is
// watching stderr, and never in CI or when opted out.
func Interactive(getenv func(string) string, stderrIsTerminal bool) bool {
	return stderrIsTerminal &&
		getenv("CI") == "" &&
		getenv("SPECDECK_NO_UPDATE_CHECK") == ""
}

// Notify writes any due notice from the cache at path to w, then calls
// startRefresh if the cache is stale. The attempt is recorded before the
// refresh starts so that runs in quick succession start only one refresh.
func Notify(w io.Writer, build Build, path string, now time.Time, startRefresh func() error) error {
	state := LoadState(path)
	msg, state := Notice(state, build, now)
	fmt.Fprint(w, msg)

	refresh := RefreshDue(state, now)
	if refresh {
		state.CheckedAt = now
	}
	if msg == "" && !refresh {
		return nil
	}
	if err := SaveState(path, state); err != nil {
		return err
	}
	if refresh {
		return startRefresh()
	}
	return nil
}
