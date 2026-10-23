package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// run executes the root command with args from inside dir and returns stdout.
func run(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	t.Chdir(dir)
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
	})
	err := rootCmd.Execute()
	return out.String(), err
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected %q to contain %q", got, want)
	}
}

func gitDir(t *testing.T, name string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

// -- new --

func TestNew_UsesNameArg(t *testing.T) {
	dir := gitDir(t, "workspace")

	if _, err := run(t, dir, "new", "jot"); err != nil {
		t.Fatalf("new: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "name: jot")
}

func TestNew_DefaultsNameToDirectory(t *testing.T) {
	dir := gitDir(t, "my-specs")

	if _, err := run(t, dir, "new"); err != nil {
		t.Fatalf("new: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "name: my-specs")
}

func TestNew_ErrorsOutsideGitRepo(t *testing.T) {
	_, err := run(t, t.TempDir(), "new", "jot")
	if err == nil || !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("expected git repo error, got %v", err)
	}
}

func TestNew_RejectsExtraArgs(t *testing.T) {
	if _, err := run(t, gitDir(t, "workspace"), "new", "a", "b"); err == nil {
		t.Error("expected error for two args")
	}
}

// -- link --

const agentsHint = "Using an AI coding agent? Run `specdeck agents claude` to install specdeck's commands.\n"

func TestLink_WritesConfigAndSuggestsAgents(t *testing.T) {
	dir := t.TempDir()

	out, err := run(t, dir, "link", "../specs")
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if out != "Linked to ../specs.\n"+agentsHint {
		t.Errorf("unexpected output:\n%s", out)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "specs_repo: ../specs")
	if _, err := os.Stat(filepath.Join(dir, ".claude")); !os.IsNotExist(err) {
		t.Error("expected link not to install agent commands")
	}
}

func TestLink_NoHintWhenAgentsAreSetUp(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "agents", "claude"); err != nil {
		t.Fatalf("agents: %v", err)
	}

	out, err := run(t, dir, "link", "../specs")
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if out != "Linked to ../specs.\n" {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestLink_RequiresExactlyOnePath(t *testing.T) {
	for _, args := range [][]string{{"link"}, {"link", "a", "b"}} {
		if _, err := run(t, t.TempDir(), args...); err == nil {
			t.Errorf("%v: expected an error", args)
		}
	}
}

func TestLink_RelinkUpdatesSpecsRepo(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "link", "../specs"); err != nil {
		t.Fatalf("link: %v", err)
	}

	if _, err := run(t, dir, "link", "../other"); err != nil {
		t.Fatalf("re-link: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "specs_repo: ../other")
}

func TestLink_ErrorsOnMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specdeck.yml"), "specs_repo: [unclosed")

	if _, err := run(t, dir, "link", "../specs"); err == nil || !strings.Contains(err.Error(), "specdeck.yml") {
		t.Errorf("expected specdeck.yml error, got %v", err)
	}
}

// -- agents --

func TestAgents_InstallsClaude(t *testing.T) {
	dir := t.TempDir()

	out, err := run(t, dir, "agents", "claude")
	if err != nil {
		t.Fatalf("agents: %v", err)
	}
	assertContains(t, out, "Updated claude commands in .claude/commands/specdeck/")
	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specdeck", "specify.md")); err != nil {
		t.Errorf("expected specify command: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "- claude")
}

func TestAgents_NoArgsUpdatesRecordedAgents(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "agents", "claude"); err != nil {
		t.Fatalf("agents claude: %v", err)
	}

	out, err := run(t, dir, "agents")
	if err != nil {
		t.Fatalf("agents: %v", err)
	}
	assertContains(t, out, "Updated claude commands")
}

func TestAgents_NoArgsAndNoneRecordedErrors(t *testing.T) {
	_, err := run(t, t.TempDir(), "agents")
	if err == nil || !strings.Contains(err.Error(), "specdeck agents claude") {
		t.Errorf("expected hint to run specdeck agents claude, got %v", err)
	}
}

func TestAgents_UnknownAgentErrors(t *testing.T) {
	_, err := run(t, t.TempDir(), "agents", "cursor")
	if err == nil || !strings.Contains(err.Error(), `unknown agent "cursor"`) {
		t.Errorf("expected unknown agent error, got %v", err)
	}
}

func TestSync_IsRemoved(t *testing.T) {
	_, err := run(t, t.TempDir(), "sync")
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected unknown command error, got %v", err)
	}
}

// -- build --

func TestBuild_WritesDistAndReportsPaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "containers", "home.yml"), "title: Home\n")
	writeFile(t, filepath.Join(dir, "containers", "user-profile", "card.yml"), "title: Card\n")

	out, err := run(t, dir, "build")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "dist", "home.md")), "# Home")
	assertContains(t, readFile(t, filepath.Join(dir, "dist", "user-profile", "_index.md")), "title: User Profile")
	assertContains(t, out, "dist/home.md")
	assertContains(t, out, "dist/user-profile/card.md")
	assertContains(t, out, "dist/user-profile/_index.md")
}

func TestBuild_ErrorsWithoutContainersDir(t *testing.T) {
	_, err := run(t, t.TempDir(), "build")
	if err == nil || !strings.Contains(err.Error(), "containers") {
		t.Errorf("expected error naming containers dir, got %v", err)
	}
}

func TestBuild_ErrorsOnInvalidContainer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "containers", "bad.yml"), "title: [unclosed\n")

	_, err := run(t, dir, "build")
	if err == nil || !strings.Contains(err.Error(), "bad.yml") {
		t.Errorf("expected error naming bad.yml, got %v", err)
	}
}

func TestBuild_RejectsArgs(t *testing.T) {
	if _, err := run(t, t.TempDir(), "build", "extra"); err == nil {
		t.Error("expected error for positional arg")
	}
}

// -- version --

func TestVersionFlag_PrintsVersion(t *testing.T) {
	out, err := run(t, t.TempDir(), "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	assertContains(t, out, "specdeck version ")
}

// -- update check --

// executeC runs the root command and returns the command cobra executed. The
// help and version flags are reset afterwards because cobra reuses them.
func executeC(t *testing.T, dir string, args ...string) *cobra.Command {
	t.Helper()
	t.Chdir(dir)
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetArgs(args)
	executed, _ := rootCmd.ExecuteC()
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetArgs(nil)
		for _, name := range []string{"help", "version"} {
			if flag := executed.Flags().Lookup(name); flag != nil {
				flag.Changed = false
				flag.Value.Set("false")
			}
		}
	})
	return executed
}

func TestSkipUpdateCheck_SkipsHelpVersionAndCompletion(t *testing.T) {
	for _, args := range [][]string{
		{"--version"},
		{"help"},
		{"build", "--help"},
		{"completion", "bash"},
		{"__complete", "bu"},
	} {
		if executed := executeC(t, t.TempDir(), args...); !skipUpdateCheck(executed) {
			t.Errorf("%v: expected update check to be skipped", args)
		}
	}
}

func TestSkipUpdateCheck_SkipsBackgroundRefresh(t *testing.T) {
	found, _, err := rootCmd.Find([]string{"__update-check"})
	if err != nil {
		t.Fatalf("finding __update-check: %v", err)
	}
	if !skipUpdateCheck(found) {
		t.Error("expected __update-check to skip the update check")
	}
}

func TestSkipUpdateCheck_RunsForRegularCommands(t *testing.T) {
	if executed := executeC(t, t.TempDir(), "link", "../specs"); skipUpdateCheck(executed) {
		t.Error("expected update check for link")
	}
}

func TestUpdateCheckCommand_IsHidden(t *testing.T) {
	found, _, err := rootCmd.Find([]string{"__update-check"})
	if err != nil || found.Name() != "__update-check" {
		t.Fatalf("expected __update-check command, got %v (%v)", found, err)
	}
	if !found.Hidden {
		t.Error("expected __update-check to be hidden")
	}
}

func TestUpdateCachePath_IsUnderUserCacheDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))

	path, err := updateCachePath()
	if err != nil {
		t.Fatalf("updateCachePath: %v", err)
	}
	if !strings.HasPrefix(path, home) || !strings.HasSuffix(path, filepath.Join("specdeck", "update-check.json")) {
		t.Errorf("unexpected cache path %q", path)
	}
}

func TestIsTerminal_RegularFileIsNot(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if isTerminal(f) {
		t.Error("expected a regular file not to be a terminal")
	}
}

// -- stale commands --

func TestStaleCommandsNotice(t *testing.T) {
	noAgents := t.TempDir()
	if _, err := run(t, noAgents, "link", "../specs"); err != nil {
		t.Fatalf("link: %v", err)
	}
	current := t.TempDir()
	if _, err := run(t, current, "agents", "claude"); err != nil {
		t.Fatalf("agents: %v", err)
	}
	stale := t.TempDir()
	if _, err := run(t, stale, "agents", "claude"); err != nil {
		t.Fatalf("agents: %v", err)
	}
	writeFile(t, filepath.Join(stale, ".claude", "commands", "specdeck", "specify.md"), "old")

	cases := map[string]struct {
		dir  string
		want string
	}{
		"no config":   {t.TempDir(), ""},
		"no agents":   {noAgents, ""},
		"up to date":  {current, ""},
		"out of date": {stale, "Agent commands are out of date — run `specdeck agents` to update.\n"},
	}
	for name, c := range cases {
		if got := staleCommandsNotice(c.dir); got != c.want {
			t.Errorf("%s: got %q, want %q", name, got, c.want)
		}
	}
}

func TestSkipStaleCheck(t *testing.T) {
	if !skipStaleCheck(agentsCmd) {
		t.Error("expected agents to skip the stale check")
	}
	if !skipStaleCheck(nil) {
		t.Error("expected no command to skip the stale check")
	}
	for _, c := range []*cobra.Command{buildCmd, linkCmd} {
		if skipStaleCheck(c) {
			t.Errorf("expected %s to run the stale check", c.Name())
		}
	}
}
