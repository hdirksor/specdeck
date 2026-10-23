package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestLink_WritesConfigAndSkills(t *testing.T) {
	dir := t.TempDir()

	if _, err := run(t, dir, "link", "../specs"); err != nil {
		t.Fatalf("link: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "specs_repo: ../specs")
	if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", "specify.md")); err != nil {
		t.Errorf("expected specify skill: %v", err)
	}
}

func TestLink_WithoutPathWritesEmptySpecsRepo(t *testing.T) {
	dir := t.TempDir()

	if _, err := run(t, dir, "link"); err != nil {
		t.Fatalf("link: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), `specs_repo: ""`)
}

func TestLink_RejectsExtraArgs(t *testing.T) {
	if _, err := run(t, t.TempDir(), "link", "a", "b"); err == nil {
		t.Error("expected error for two args")
	}
}

func TestLink_KeepsExistingConfig(t *testing.T) {
	dir := t.TempDir()
	if _, err := run(t, dir, "link", "../specs"); err != nil {
		t.Fatalf("link: %v", err)
	}

	if _, err := run(t, dir, "link", "../other"); err != nil {
		t.Fatalf("re-link: %v", err)
	}
	assertContains(t, readFile(t, filepath.Join(dir, "specdeck.yml")), "specs_repo: ../specs")
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
