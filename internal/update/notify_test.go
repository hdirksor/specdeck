package update_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/hdirksor/specdeck/internal/update"
)

var goInstall = update.Build{Version: "v0.2.1", Method: update.MethodGoInstall}

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestEnabled(t *testing.T) {
	cases := []struct {
		name     string
		build    update.Build
		vars     map[string]string
		terminal bool
		want     bool
	}{
		{"interactive release", goInstall, nil, true, true},
		{"not a terminal", goInstall, nil, false, false},
		{"CI", goInstall, map[string]string{"CI": "true"}, true, false},
		{"opted out", goInstall, map[string]string{"SPECDECK_NO_UPDATE_CHECK": "1"}, true, false},
		{"dev build", update.Build{Version: "dev", Method: update.MethodDev}, nil, true, false},
	}
	for _, c := range cases {
		if got := update.Enabled(c.build, env(c.vars), c.terminal); got != c.want {
			t.Errorf("%s: Enabled = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestInteractive(t *testing.T) {
	cases := []struct {
		name     string
		vars     map[string]string
		terminal bool
		want     bool
	}{
		{"terminal", nil, true, true},
		{"not a terminal", nil, false, false},
		{"CI", map[string]string{"CI": "true"}, true, false},
		{"opted out", map[string]string{"SPECDECK_NO_UPDATE_CHECK": "1"}, true, false},
	}
	for _, c := range cases {
		if got := update.Interactive(env(c.vars), c.terminal); got != c.want {
			t.Errorf("%s: Interactive = %v, want %v", c.name, got, c.want)
		}
	}
}

func noRefresh(t *testing.T) func() error {
	return func() error {
		t.Error("refresh should not start")
		return nil
	}
}

func TestNotify_PrintsCachedNoticeAndRecordsIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update-check.json")
	state := newerState()
	state.CheckedAt = now
	if err := update.SaveState(path, state); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer

	if err := update.Notify(&out, goInstall, path, now, noRefresh(t)); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	assertContainsText(t, out.String(), "specdeck v0.3.0 is available (you have v0.2.1)")
	if got := update.LoadState(path); !got.NotifiedAt.Equal(now) {
		t.Errorf("NotifiedAt = %v, want %v", got.NotifiedAt, now)
	}
}

func TestNotify_StartsRefreshAfterRecordingAttempt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update-check.json")
	started := false
	var out bytes.Buffer

	err := update.Notify(&out, goInstall, path, now, func() error {
		started = true
		if got := update.LoadState(path); !got.CheckedAt.Equal(now) {
			t.Errorf("CheckedAt should be saved before refresh starts, got %v", got.CheckedAt)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if !started {
		t.Error("expected refresh to start for an empty cache")
	}
	if out.Len() != 0 {
		t.Errorf("expected no output, got %q", out.String())
	}
}

func TestNotify_SkipsRefreshWhenCheckedRecently(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update-check.json")
	if err := update.SaveState(path, update.State{CheckedAt: now.Add(-time.Hour)}); err != nil {
		t.Fatal(err)
	}

	if err := update.Notify(&bytes.Buffer{}, goInstall, path, now, noRefresh(t)); err != nil {
		t.Fatalf("Notify: %v", err)
	}
}

func TestNotify_ReportsRefreshStartError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "update-check.json")
	failure := errors.New("no executable")

	err := update.Notify(&bytes.Buffer{}, goInstall, path, now, func() error { return failure })
	if !errors.Is(err, failure) {
		t.Errorf("expected refresh start error, got %v", err)
	}
}

func TestNotify_ReportsCacheWriteError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-parent-is-a-file")
	if err := update.SaveState(path, update.State{}); err != nil {
		t.Fatal(err)
	}

	err := update.Notify(&bytes.Buffer{}, goInstall, filepath.Join(path, "update-check.json"), now, noRefresh(t))
	if err == nil {
		t.Error("expected cache write error")
	}
}

func assertContainsText(t *testing.T, got, want string) {
	t.Helper()
	if !bytes.Contains([]byte(got), []byte(want)) {
		t.Errorf("expected %q to contain %q", got, want)
	}
}
