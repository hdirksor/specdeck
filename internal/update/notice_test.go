package update_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hdirksor/specdeck/internal/update"
)

var now = time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC)

const releaseURL = "https://github.com/hdirksor/specdeck/releases/tag/v0.3.0"

func newerState() update.State {
	return update.State{LatestVersion: "v0.3.0", LatestURL: releaseURL}
}

func TestNotice_GoInstallShowsVersionLinkAndUpgradeCommand(t *testing.T) {
	msg, _ := update.Notice(newerState(), update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}, now)

	want := "specdeck v0.3.0 is available (you have v0.2.1) — " + releaseURL + "\n" +
		"Upgrade: go install github.com/hdirksor/specdeck@latest\n"
	if msg != want {
		t.Errorf("got:\n%s\nwant:\n%s", msg, want)
	}
}

func TestNotice_ReleaseDownloadShowsLinkOnly(t *testing.T) {
	msg, _ := update.Notice(newerState(), update.Build{Version: "v0.2.1", Method: update.MethodRelease}, now)

	want := "specdeck v0.3.0 is available (you have v0.2.1) — " + releaseURL + "\n"
	if msg != want {
		t.Errorf("got:\n%s\nwant:\n%s", msg, want)
	}
}

func TestNotice_RecordsWhenShown(t *testing.T) {
	_, state := update.Notice(newerState(), update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}, now)

	if !state.NotifiedAt.Equal(now) {
		t.Errorf("NotifiedAt = %v, want %v", state.NotifiedAt, now)
	}
}

func TestNotice_NothingWhenUpToDate(t *testing.T) {
	state := newerState()
	msg, got := update.Notice(state, update.Build{Version: "v0.3.0", Method: update.MethodGoInstall}, now)

	if msg != "" {
		t.Errorf("expected no notice, got %q", msg)
	}
	if got != state {
		t.Errorf("state changed: %+v", got)
	}
}

func TestNotice_NothingWithoutCachedRelease(t *testing.T) {
	msg, _ := update.Notice(update.State{}, update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}, now)

	if msg != "" {
		t.Errorf("expected no notice, got %q", msg)
	}
}

func TestNotice_AtMostOncePerDay(t *testing.T) {
	build := update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}
	state := newerState()
	state.NotifiedAt = now.Add(-23 * time.Hour)

	if msg, _ := update.Notice(state, build, now); msg != "" {
		t.Errorf("expected no notice within a day, got %q", msg)
	}

	state.NotifiedAt = now.Add(-25 * time.Hour)
	if msg, _ := update.Notice(state, build, now); msg == "" {
		t.Error("expected a notice after a day")
	}
}

func TestState_RoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "specdeck", "update-check.json")
	state := update.State{CheckedAt: now, LatestVersion: "v0.3.0", LatestURL: releaseURL, NotifiedAt: now}

	if err := update.SaveState(path, state); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	got := update.LoadState(path)
	if !got.CheckedAt.Equal(now) || !got.NotifiedAt.Equal(now) || got.LatestVersion != "v0.3.0" || got.LatestURL != releaseURL {
		t.Errorf("got %+v, want %+v", got, state)
	}
}

func TestLoadState_MissingOrCorruptIsEmpty(t *testing.T) {
	dir := t.TempDir()
	corrupt := filepath.Join(dir, "corrupt.json")
	if err := os.WriteFile(corrupt, []byte("{not json"), 0644); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{filepath.Join(dir, "missing.json"), corrupt} {
		if got := update.LoadState(path); got != (update.State{}) {
			t.Errorf("%s: got %+v, want empty state", path, got)
		}
	}
}

func TestSaveState_ReportsWriteErrors(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0644); err != nil {
		t.Fatal(err)
	}

	err := update.SaveState(filepath.Join(blocker, "update-check.json"), update.State{})
	if err == nil || !strings.Contains(err.Error(), "update-check") {
		t.Errorf("expected write error, got %v", err)
	}
}

func TestSaveState_ReportsErrorWhenPathIsADirectory(t *testing.T) {
	path := t.TempDir()

	err := update.SaveState(path, update.State{})
	if err == nil || !strings.Contains(err.Error(), "writing update-check cache") {
		t.Errorf("expected write error, got %v", err)
	}
}
