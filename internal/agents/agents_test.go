package agents_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/agents"
	"github.com/hdirksor/specdeck/internal/config"
)

func claudeFile(dir, name string) string {
	return filepath.Join(dir, ".claude", "commands", "specdeck", name)
}

func update(t *testing.T, dir string, names ...string) []agents.Installed {
	t.Helper()
	installed, err := agents.Update(dir, names)
	if err != nil {
		t.Fatalf("Update(%v): %v", names, err)
	}
	return installed
}

func TestUpdate_InstallsClaudeCommands(t *testing.T) {
	dir := t.TempDir()

	installed := update(t, dir, "claude")

	want := []agents.Installed{{Name: "claude", Dir: ".claude/commands/specdeck"}}
	if !reflect.DeepEqual(installed, want) {
		t.Errorf("installed = %+v, want %+v", installed, want)
	}
	data, err := os.ReadFile(claudeFile(dir, "specify.md"))
	if err != nil {
		t.Fatalf("expected specify.md: %v", err)
	}
	if !strings.Contains(string(data), "specdeck.yml") {
		t.Errorf("specify.md should reference specdeck.yml, got:\n%s", data)
	}
}

func TestUpdate_InstallsOnlyUnderSpecdeckNamespace(t *testing.T) {
	dir := t.TempDir()

	update(t, dir, "claude")

	entries, err := os.ReadDir(filepath.Join(dir, ".claude", "commands"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "specdeck" {
			t.Errorf("unexpected .claude/commands/%s; commands belong in .claude/commands/specdeck/", entry.Name())
		}
	}
}

func TestUpdate_RecordsAgentAndKeepsSpecsRepo(t *testing.T) {
	dir := t.TempDir()
	if err := config.Save(dir, config.Config{SpecsRepo: "../specs"}); err != nil {
		t.Fatal(err)
	}

	update(t, dir, "claude")
	update(t, dir, "claude")

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := config.Config{SpecsRepo: "../specs", Agents: []string{"claude"}}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("config = %+v, want %+v", cfg, want)
	}
}

func TestUpdate_NoNamesUpdatesRecordedAgents(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	if err := os.WriteFile(claudeFile(dir, "specify.md"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	installed := update(t, dir)

	if len(installed) != 1 || installed[0].Name != "claude" {
		t.Errorf("installed = %+v, want claude", installed)
	}
	if data, _ := os.ReadFile(claudeFile(dir, "specify.md")); string(data) == "old" {
		t.Error("expected specify.md to be rewritten")
	}
}

func TestUpdate_NoNamesAndNoneRecordedErrors(t *testing.T) {
	_, err := agents.Update(t.TempDir(), nil)

	if !errors.Is(err, agents.ErrNoAgents) {
		t.Errorf("expected ErrNoAgents, got %v", err)
	}
	if err == nil || !strings.Contains(err.Error(), "specdeck agents claude") {
		t.Errorf("expected hint to run specdeck agents claude, got %v", err)
	}
}

func TestUpdate_UnknownAgentErrorsWithSupportedList(t *testing.T) {
	dir := t.TempDir()

	_, err := agents.Update(dir, []string{"cursor"})

	if err == nil || !strings.Contains(err.Error(), `unknown agent "cursor" (supported: claude)`) {
		t.Errorf("expected unknown agent error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "specdeck.yml")); !os.IsNotExist(statErr) {
		t.Error("expected nothing to be written for an unknown agent")
	}
}

func TestUpdate_RemovesFilesNoLongerShipped(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	if err := os.WriteFile(claudeFile(dir, "change.md"), []byte("retired"), 0644); err != nil {
		t.Fatal(err)
	}

	update(t, dir)

	if _, err := os.Stat(claudeFile(dir, "change.md")); !os.IsNotExist(err) {
		t.Error("expected change.md to be removed")
	}
}

func TestUpdate_ErrorsWhenCommandsDirCannotBeCreated(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".claude"), nil, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := agents.Update(dir, []string{"claude"})
	if err == nil || !strings.Contains(err.Error(), ".claude/commands/specdeck") {
		t.Errorf("expected commands dir error, got %v", err)
	}
}

func TestUpdate_ErrorsOnMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "specdeck.yml"), []byte("agents: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := agents.Update(dir, []string{"claude"}); err == nil || !strings.Contains(err.Error(), "specdeck.yml") {
		t.Errorf("expected specdeck.yml error, got %v", err)
	}
}

func TestStale_NoneWhenNothingRecorded(t *testing.T) {
	stale, err := agents.Stale(t.TempDir())
	if err != nil {
		t.Fatalf("Stale: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected nothing stale, got %v", stale)
	}
}

func TestStale_NoneAfterUpdate(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")

	stale, err := agents.Stale(dir)
	if err != nil {
		t.Fatalf("Stale: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected nothing stale, got %v", stale)
	}
}

func TestStale_ReportsChangedMissingAndExtraFiles(t *testing.T) {
	cases := map[string]func(dir string) error{
		"changed": func(dir string) error { return os.WriteFile(claudeFile(dir, "specify.md"), []byte("old"), 0644) },
		"missing": func(dir string) error { return os.Remove(claudeFile(dir, "specify.md")) },
		"extra":   func(dir string) error { return os.WriteFile(claudeFile(dir, "change.md"), []byte("retired"), 0644) },
	}
	want := map[string]string{
		"changed": ".claude/commands/specdeck/specify.md",
		"missing": ".claude/commands/specdeck/specify.md",
		"extra":   ".claude/commands/specdeck/change.md",
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			update(t, dir, "claude")
			if err := breakIt(dir); err != nil {
				t.Fatal(err)
			}

			stale, err := agents.Stale(dir)
			if err != nil {
				t.Fatalf("Stale: %v", err)
			}
			if len(stale) != 1 || stale[0] != want[name] {
				t.Errorf("stale = %v, want [%s]", stale, want[name])
			}
		})
	}
}

func TestStale_ReportsMissingFolderForRecordedAgent(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	if err := os.RemoveAll(filepath.Join(dir, ".claude")); err != nil {
		t.Fatal(err)
	}

	stale, err := agents.Stale(dir)
	if err != nil {
		t.Fatalf("Stale: %v", err)
	}
	if len(stale) != 1 || stale[0] != ".claude/commands/specdeck/specify.md" {
		t.Errorf("stale = %v, want specify.md", stale)
	}
}

func TestStale_ReportsReadErrors(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	// A directory where a command file should be cannot be read as a file.
	if err := os.Remove(claudeFile(dir, "specify.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(claudeFile(dir, "specify.md"), 0755); err != nil {
		t.Fatal(err)
	}

	if _, err := agents.Stale(dir); err == nil || !strings.Contains(err.Error(), "specify.md") {
		t.Errorf("expected read error naming specify.md, got %v", err)
	}
}

// skipIfRoot skips tests that rely on permissions, which root ignores.
func skipIfRoot(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
}

// chmod changes path's mode and restores it so the temp dir can be removed.
func chmod(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(path, 0755) })
}

func TestUpdate_ErrorsWhenCommandCannotBeWritten(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(claudeFile(dir, "specify.md"), 0755); err != nil {
		t.Fatal(err)
	}

	_, err := agents.Update(dir, []string{"claude"})
	if err == nil || !strings.Contains(err.Error(), "writing .claude/commands/specdeck/specify.md") {
		t.Errorf("expected write error, got %v", err)
	}
}

func TestUpdate_ErrorsWhenCommandsDirCannotBeListed(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	update(t, dir, "claude")
	chmod(t, filepath.Dir(claudeFile(dir, "specify.md")), 0300)

	_, err := agents.Update(dir, []string{"claude"})
	if err == nil || !strings.Contains(err.Error(), "reading .claude/commands/specdeck") {
		t.Errorf("expected listing error, got %v", err)
	}
}

func TestUpdate_ErrorsWhenRetiredFileCannotBeRemoved(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	update(t, dir, "claude")
	writeTestFile(t, claudeFile(dir, "change.md"), "retired")
	chmod(t, filepath.Dir(claudeFile(dir, "specify.md")), 0555)

	_, err := agents.Update(dir, []string{"claude"})
	if err == nil || !strings.Contains(err.Error(), "removing .claude/commands/specdeck/change.md") {
		t.Errorf("expected remove error, got %v", err)
	}
}

func TestUpdate_ErrorsWhenConfigCannotBeSaved(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	update(t, dir, "claude")
	if err := os.Remove(filepath.Join(dir, "specdeck.yml")); err != nil {
		t.Fatal(err)
	}
	chmod(t, dir, 0555)

	_, err := agents.Update(dir, []string{"claude"})
	if err == nil || !strings.Contains(err.Error(), "writing specdeck.yml") {
		t.Errorf("expected config write error, got %v", err)
	}
}

func TestUpdate_KeepsSubfolders(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	writeTestFile(t, claudeFile(dir, filepath.Join("mine", "notes.md")), "mine")

	update(t, dir)

	if _, err := os.Stat(claudeFile(dir, filepath.Join("mine", "notes.md"))); err != nil {
		t.Errorf("expected subfolder to be kept: %v", err)
	}
}

func TestStale_IgnoresSubfolders(t *testing.T) {
	dir := t.TempDir()
	update(t, dir, "claude")
	writeTestFile(t, claudeFile(dir, filepath.Join("mine", "notes.md")), "mine")

	stale, err := agents.Stale(dir)
	if err != nil {
		t.Fatalf("Stale: %v", err)
	}
	if len(stale) != 0 {
		t.Errorf("expected nothing stale, got %v", stale)
	}
}

func TestStale_ErrorsWhenCommandsDirCannotBeListed(t *testing.T) {
	skipIfRoot(t)
	dir := t.TempDir()
	update(t, dir, "claude")
	chmod(t, filepath.Dir(claudeFile(dir, "specify.md")), 0300)

	if _, err := agents.Stale(dir); err == nil || !strings.Contains(err.Error(), "reading .claude/commands/specdeck") {
		t.Errorf("expected listing error, got %v", err)
	}
}

func TestStale_ErrorsOnMalformedConfig(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "specdeck.yml"), "agents: [unclosed")

	if _, err := agents.Stale(dir); err == nil || !strings.Contains(err.Error(), "specdeck.yml") {
		t.Errorf("expected specdeck.yml error, got %v", err)
	}
}

func TestStale_ErrorsOnUnknownRecordedAgent(t *testing.T) {
	dir := t.TempDir()
	if err := config.Save(dir, config.Config{Agents: []string{"cursor"}}); err != nil {
		t.Fatal(err)
	}

	if _, err := agents.Stale(dir); err == nil || !strings.Contains(err.Error(), `unknown agent "cursor"`) {
		t.Errorf("expected unknown agent error, got %v", err)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
