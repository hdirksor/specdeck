package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/scaffold"
	"github.com/hdickson/specdeck/internal/spec"
)

func setupProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := scaffold.New(dir, "testproject"); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeStateFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	path := filepath.Join(dir, "states", filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestIntermediateDirectories_NoneExist(t *testing.T) {
	dir := setupProject(t)
	intermediates := scaffold.IntermediateDirectories(dir, "app/home-tab/feed-screen/post-card")
	// expects: containers/app, containers/app/home-tab, containers/app/home-tab/feed-screen
	if len(intermediates) != 3 {
		t.Errorf("expected 3 intermediates, got %d: %v", len(intermediates), intermediates)
	}
}

func TestIntermediateDirectories_SomeExist(t *testing.T) {
	dir := setupProject(t)
	if err := os.MkdirAll(filepath.Join(dir, "containers", "app", "home-tab"), 0755); err != nil {
		t.Fatal(err)
	}
	intermediates := scaffold.IntermediateDirectories(dir, "app/home-tab/feed-screen/post-card")
	// containers/app and containers/app/home-tab already exist
	if len(intermediates) != 1 {
		t.Errorf("expected 1 intermediate, got %d: %v", len(intermediates), intermediates)
	}
	if intermediates[0] != filepath.Join("containers", "app", "home-tab", "feed-screen") {
		t.Errorf("unexpected intermediate: %q", intermediates[0])
	}
}

func TestAddContainer_CreatesFile(t *testing.T) {
	dir := setupProject(t)
	// setupProject already provides a 'default' state via core.yml; add one more
	writeStateFile(t, dir, "app.yml", `
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	if err := scaffold.AddContainer(dir, "app/home-tab/post-card"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(dir, "containers", "app", "home-tab", "post-card.yml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file at %s", path)
	}

	c, err := spec.ParseContainer(path)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(c.States) != 2 {
		t.Errorf("expected 2 states (one per existing state), got %d", len(c.States))
	}
}

func TestAddContainer_StubsAllStates(t *testing.T) {
	dir := setupProject(t)
	// setupProject provides 'default'; add two more
	writeStateFile(t, dir, "app.yml", `
- name: spanish
  summary: Spanish user
  facts: {}
- name: dark-mode
  summary: Dark mode
  facts: {}
`)

	if err := scaffold.AddContainer(dir, "screen/hero"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	c, err := spec.ParseContainer(filepath.Join(dir, "containers", "screen", "hero.yml"))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(c.States) != 3 {
		t.Errorf("expected 3 states, got %d", len(c.States))
	}
	for _, s := range c.States {
		if len(s.Specs) != 0 {
			t.Errorf("expected empty specs for state %q, got %v", s.Ref, s.Specs)
		}
		if len(s.Events) != 0 {
			t.Errorf("expected empty events for state %q, got %v", s.Ref, s.Events)
		}
	}
}

func TestAddContainer_ErrorsIfFileAlreadyExists(t *testing.T) {
	dir := setupProject(t)
	if err := scaffold.AddContainer(dir, "hero"); err != nil {
		t.Fatal(err)
	}
	if err := scaffold.AddContainer(dir, "hero"); err == nil {
		t.Fatal("expected error when container file already exists")
	}
}
