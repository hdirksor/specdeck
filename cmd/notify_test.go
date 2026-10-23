package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hdirksor/specdeck/internal/update"
)

// TestMain stands in for the background refresh that startUpdateCheck
// launches, since os.Executable is the test binary.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == updateCheckCmd.Name() {
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fakeTerminal sends notices to a buffer that counts as an interactive
// terminal and returns the buffer.
func fakeTerminal(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	oldStderr, oldIsTerminal := stderr, stderrIsTerminal
	stderr = &buf
	stderrIsTerminal = func() bool { return true }
	t.Setenv("CI", "")
	t.Setenv("SPECDECK_NO_UPDATE_CHECK", "")
	t.Cleanup(func() { stderr, stderrIsTerminal = oldStderr, oldIsTerminal })
	return &buf
}

// useBuild makes specdeck behave as the given build for the test.
func useBuild(t *testing.T, build update.Build) {
	t.Helper()
	old := currentBuild
	currentBuild = build
	t.Cleanup(func() { currentBuild = old })
}

// useCacheHome points the user cache directory at a temp dir and returns the
// update-check cache path inside it.
func useCacheHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	path, err := updateCachePath()
	if err != nil {
		t.Fatalf("updateCachePath: %v", err)
	}
	return path
}

// noCacheHome makes the user cache directory unavailable.
func noCacheHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
}

// countRefreshes replaces the background refresh with a counter.
func countRefreshes(t *testing.T) *int {
	t.Helper()
	var count int
	old := startRefresh
	startRefresh = func() error { count++; return nil }
	t.Cleanup(func() { startRefresh = old })
	return &count
}

// executeArgs runs execute with args from inside dir and returns its exit code.
func executeArgs(t *testing.T, dir string, args ...string) int {
	t.Helper()
	t.Chdir(dir)
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})
	return execute()
}

var goInstallBuild = update.Build{Version: "v0.1.0", Method: update.MethodGoInstall}

// -- execute --

func TestExecute_ReturnsZeroOnSuccess(t *testing.T) {
	out := fakeTerminal(t)

	if code := executeArgs(t, t.TempDir(), "link", "../specs"); code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if out.Len() != 0 {
		t.Errorf("expected nothing on stderr, got %q", out)
	}
}

func TestExecute_PrintsErrorAndReturnsOne(t *testing.T) {
	out := fakeTerminal(t)

	if code := executeArgs(t, t.TempDir(), "link"); code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	assertContains(t, out.String(), "accepts 1 arg")
}

func TestExecute_PrintsStaleCommandsNotice(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "agents", "claude"); err != nil {
		t.Fatalf("agents: %v", err)
	}
	writeFile(t, filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"), "old")
	out := fakeTerminal(t)

	executeArgs(t, dir, "link", "../specs")

	assertContains(t, out.String(), "Agent commands are out of date")
}

func TestExecute_NoNoticesWhenStderrIsNotATerminal(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "agents", "claude"); err != nil {
		t.Fatalf("agents: %v", err)
	}
	writeFile(t, filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md"), "old")
	out := fakeTerminal(t)
	stderrIsTerminal = func() bool { return false }

	executeArgs(t, dir, "link", "../specs")

	if out.Len() != 0 {
		t.Errorf("expected nothing on stderr, got %q", out)
	}
}

// -- update notice --

func TestNotifyUpdates_PrintsCachedReleaseAndStartsRefresh(t *testing.T) {
	out := fakeTerminal(t)
	useBuild(t, goInstallBuild)
	path := useCacheHome(t)
	refreshes := countRefreshes(t)
	state := update.State{LatestVersion: "v0.2.0", LatestURL: "https://example.com/v0.2.0"}
	if err := update.SaveState(path, state); err != nil {
		t.Fatal(err)
	}

	notifyUpdates(linkCmd)

	assertContains(t, out.String(), "specdeck v0.2.0 is available (you have v0.1.0)")
	if *refreshes != 1 {
		t.Errorf("refreshes = %d, want 1", *refreshes)
	}
}

func TestNotifyUpdates_SilentForDevBuilds(t *testing.T) {
	out := fakeTerminal(t)
	useBuild(t, update.Build{Version: "dev", Method: update.MethodDev})
	useCacheHome(t)
	refreshes := countRefreshes(t)

	notifyUpdates(linkCmd)

	if out.Len() != 0 || *refreshes != 0 {
		t.Errorf("expected no notice or refresh, got %q and %d refreshes", out, *refreshes)
	}
}

func TestNotifyUpdates_SilentWithoutCacheDir(t *testing.T) {
	out := fakeTerminal(t)
	useBuild(t, goInstallBuild)
	noCacheHome(t)
	refreshes := countRefreshes(t)

	notifyUpdates(linkCmd)

	if out.Len() != 0 || *refreshes != 0 {
		t.Errorf("expected no notice or refresh, got %q and %d refreshes", out, *refreshes)
	}
}

func TestUpdateCachePath_ErrorsWithoutCacheDir(t *testing.T) {
	noCacheHome(t)

	if _, err := updateCachePath(); err == nil {
		t.Error("expected an error without a cache directory")
	}
}

// -- background refresh --

func TestUpdateCheckCommand_StoresLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"tag_name":"v0.2.0","html_url":"https://example.com/v0.2.0"}`))
	}))
	t.Cleanup(server.Close)
	old := latestReleaseURL
	latestReleaseURL = server.URL
	t.Cleanup(func() { latestReleaseURL = old })
	path := useCacheHome(t)

	if _, err := run(t, t.TempDir(), "__update-check"); err != nil {
		t.Fatalf("__update-check: %v", err)
	}

	if got := update.LoadState(path).LatestVersion; got != "v0.2.0" {
		t.Errorf("cached version = %q, want v0.2.0", got)
	}
}

func TestUpdateCheckCommand_ErrorsWithoutCacheDir(t *testing.T) {
	noCacheHome(t)

	if _, err := run(t, t.TempDir(), "__update-check"); err == nil {
		t.Error("expected an error without a cache directory")
	}
}

func TestStartUpdateCheck_LaunchesBackgroundProcess(t *testing.T) {
	start := time.Now()

	if err := startUpdateCheck(); err != nil {
		t.Fatalf("startUpdateCheck: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("expected startUpdateCheck not to wait for the process")
	}
}
