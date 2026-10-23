package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/scaffold"
	"gopkg.in/yaml.v3"
)

// gitDir returns an empty temp directory containing only a .git directory.
func gitDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func newProject(t *testing.T, name string) string {
	t.Helper()
	dir := gitDir(t)
	if err := scaffold.New(dir, name); err != nil {
		t.Fatalf("New: %v", err)
	}
	return dir
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected %s not to exist", path)
	}
}

func TestNew_RequiresGitRepo(t *testing.T) {
	dir := t.TempDir()

	err := scaffold.New(dir, "jot")

	assertErrorContains(t, err, "not a git repository")
	assertNotExists(t, filepath.Join(dir, "specdeck.yml"))
}

func TestNew_AcceptsGitFile(t *testing.T) {
	// Worktrees and submodules use a .git file instead of a directory.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: ../repo/.git\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.New(dir, "jot"); err != nil {
		t.Fatalf("New: %v", err)
	}
}

func TestNew_RequiresEmptyDir(t *testing.T) {
	dir := gitDir(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}

	err := scaffold.New(dir, "jot")

	assertErrorContains(t, err, "not empty")
	assertNotExists(t, filepath.Join(dir, "specdeck.yml"))
}

func TestNew_FailsOnSecondRun(t *testing.T) {
	dir := newProject(t, "jot")

	assertErrorContains(t, scaffold.New(dir, "jot"), "not empty")
}

func TestNew_WritesConfigWithName(t *testing.T) {
	dir := newProject(t, "jot")

	data, err := os.ReadFile(filepath.Join(dir, "specdeck.yml"))
	if err != nil {
		t.Fatalf("reading specdeck.yml: %v", err)
	}
	var cfg struct {
		Name string `yaml:"name"`
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("decoding specdeck.yml: %v", err)
	}
	if cfg.Name != "jot" {
		t.Errorf("name = %q, want %q", cfg.Name, "jot")
	}
}

func TestNew_CreatesStarterFiles(t *testing.T) {
	dir := newProject(t, "jot")

	for _, rel := range []string{
		"states/facts/is-logged-in.yml",
		"states/core.yml",
		"containers/.gitkeep",
		"changes/TEMPLATE.md",
		".claude/commands/change.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("expected %s to exist: %v", rel, err)
		}
	}
}

func TestNew_StarterStateUsesStarterFact(t *testing.T) {
	dir := newProject(t, "jot")

	fact := readFile(t, filepath.Join(dir, "states", "facts", "is-logged-in.yml"))
	if !strings.Contains(fact, "name: is-logged-in") {
		t.Errorf("starter fact missing name, got:\n%s", fact)
	}
	states := readFile(t, filepath.Join(dir, "states", "core.yml"))
	if !strings.Contains(states, "is-logged-in: false") {
		t.Errorf("starter state does not reference starter fact, got:\n%s", states)
	}
}

func TestNew_ChangeTemplateHasSections(t *testing.T) {
	dir := newProject(t, "jot")

	tmpl := readFile(t, filepath.Join(dir, "changes", "TEMPLATE.md"))
	for _, heading := range []string{"## What changed", "## Why", "## References"} {
		if !strings.Contains(tmpl, heading) {
			t.Errorf("TEMPLATE.md missing %q", heading)
		}
	}
}

func TestNew_ChangeSkillHasFrontmatter(t *testing.T) {
	dir := newProject(t, "jot")

	skill := readFile(t, filepath.Join(dir, ".claude", "commands", "change.md"))
	if !strings.HasPrefix(skill, "---\ndescription: ") {
		t.Errorf("change.md missing description frontmatter, got:\n%s", skill)
	}
}

func TestNew_DoesNotWriteTOML(t *testing.T) {
	dir := newProject(t, "jot")

	assertNotExists(t, filepath.Join(dir, "specdeck.toml"))
}

func TestNew_ErrorsWhenConfigCannotBeCreated(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	dir := gitDir(t)
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0755) })

	assertErrorContains(t, scaffold.New(dir, "jot"), "creating specdeck.yml")
}
