package build_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/build"
)

// -- helpers --

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// -- ParseFile --

func TestParseFile_Title(t *testing.T) {
	f := writeTempFile(t, `title: My Screen`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Title != "My Screen" {
		t.Errorf("title: want %q, got %q", "My Screen", c.Title)
	}
}

func TestParseFile_AllScalarFields(t *testing.T) {
	f := writeTempFile(t, `
title: Card
description: A simple card.
behavior:
  - scrollable
  - tappable
specs:
  color: blue
  padding: 8dp
`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Description != "A simple card." {
		t.Errorf("description: want %q, got %q", "A simple card.", c.Description)
	}
	if len(c.Behavior) != 2 || c.Behavior[0] != "scrollable" {
		t.Errorf("behavior: want [scrollable tappable], got %v", c.Behavior)
	}
	if c.Specs["color"] != "blue" {
		t.Errorf("specs.color: want %q, got %q", "blue", c.Specs["color"])
	}
	if c.Specs["padding"] != "8dp" {
		t.Errorf("specs.padding: want %q, got %q", "8dp", c.Specs["padding"])
	}
}

func TestParseFile_States(t *testing.T) {
	f := writeTempFile(t, `
title: Button
specs:
  label: Submit
states:
  - title: Loading
    specs:
      label: Loading...
`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States) != 1 {
		t.Fatalf("states: want 1, got %d", len(c.States))
	}
	if c.States[0].Title != "Loading" {
		t.Errorf("state title: want %q, got %q", "Loading", c.States[0].Title)
	}
	if c.States[0].Specs["label"] != "Loading..." {
		t.Errorf("state specs.label: want %q, got %q", "Loading...", c.States[0].Specs["label"])
	}
}

func TestParseFile_Events(t *testing.T) {
	f := writeTempFile(t, `
title: Form
events:
  - title: submit
    description: User taps submit.
    actions:
      - title: save
        description: Persist the form data.
      - title: navigate
`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Events) != 1 {
		t.Fatalf("events: want 1, got %d", len(c.Events))
	}
	ev := c.Events[0]
	if ev.Title != "submit" {
		t.Errorf("event title: want %q, got %q", "submit", ev.Title)
	}
	if ev.Description != "User taps submit." {
		t.Errorf("event description: want %q, got %q", "User taps submit.", ev.Description)
	}
	if len(ev.Actions) != 2 {
		t.Fatalf("actions: want 2, got %d", len(ev.Actions))
	}
	if ev.Actions[0].Title != "save" {
		t.Errorf("action[0] title: want %q, got %q", "save", ev.Actions[0].Title)
	}
	if ev.Actions[0].Description != "Persist the form data." {
		t.Errorf("action[0] description: want %q, got %q", "Persist the form data.", ev.Actions[0].Description)
	}
}

func TestParseFile_RefStubsNotResolved(t *testing.T) {
	f := writeTempFile(t, `
title: Screen
containers:
  - $ref: ./hero.yml
  - title: Inline
`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Containers) != 2 {
		t.Fatalf("containers: want 2, got %d", len(c.Containers))
	}
	if c.Containers[0].Ref != "./hero.yml" {
		t.Errorf("ref: want %q, got %q", "./hero.yml", c.Containers[0].Ref)
	}
	if c.Containers[1].Title != "Inline" {
		t.Errorf("inline title: want %q, got %q", "Inline", c.Containers[1].Title)
	}
}

func TestParseFile_MissingFile(t *testing.T) {
	_, err := build.ParseFile("/does/not/exist.yml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseFile_InvalidYAML(t *testing.T) {
	f := writeTempFile(t, `title: [unclosed`)
	_, err := build.ParseFile(f)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// -- Load --

func TestLoad_SetsPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "card.yml"), `title: Card`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("want 1 container, got %d", len(containers))
	}
	if containers[0].Path != "card.yml" {
		t.Errorf("path: want %q, got %q", "card.yml", containers[0].Path)
	}
}

func TestLoad_WalksSubdirectories(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "jot", "index.yml"), `title: Jot`)
	writeFile(t, filepath.Join(dir, "shared", "nav-bar.yml"), `title: Nav Bar`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Fatalf("want 2 containers, got %d", len(containers))
	}
}

func TestLoad_IgnoresNonYML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "card.yml"), `title: Card`)
	writeFile(t, filepath.Join(dir, "README.md"), `# readme`)
	writeFile(t, filepath.Join(dir, ".gitkeep"), ``)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("want 1 container, got %d", len(containers))
	}
}

func TestLoad_ResolvesRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hero.yml"), `
title: Hero
specs:
  color: red
`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./hero.yml
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var screen build.Container
	for _, c := range containers {
		if c.Title == "Screen" {
			screen = c
		}
	}
	if screen.Title == "" {
		t.Fatal("screen container not found")
	}
	if len(screen.Containers) != 1 {
		t.Fatalf("screen.containers: want 1, got %d", len(screen.Containers))
	}
	hero := screen.Containers[0]
	if hero.Title != "Hero" {
		t.Errorf("resolved title: want %q, got %q", "Hero", hero.Title)
	}
	if hero.Specs["color"] != "red" {
		t.Errorf("resolved specs.color: want %q, got %q", "red", hero.Specs["color"])
	}
	if hero.Ref != "" {
		t.Errorf("ref should be cleared after resolution, got %q", hero.Ref)
	}
}

func TestLoad_RefWithOverrides(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "button.yml"), `
title: Button
specs:
  label: Default
  color: blue
`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./button.yml
    overrides:
      label: Submit
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var screen build.Container
	for _, c := range containers {
		if c.Title == "Screen" {
			screen = c
		}
	}
	btn := screen.Containers[0]
	if btn.Specs["label"] != "Submit" {
		t.Errorf("override label: want %q, got %q", "Submit", btn.Specs["label"])
	}
	if btn.Specs["color"] != "blue" {
		t.Errorf("inherited color: want %q, got %q", "blue", btn.Specs["color"])
	}
}

func TestLoad_NestedRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "icon.yml"), `title: Icon`)
	writeFile(t, filepath.Join(dir, "button.yml"), `
title: Button
containers:
  - $ref: ./icon.yml
`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./button.yml
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var screen build.Container
	for _, c := range containers {
		if c.Title == "Screen" {
			screen = c
		}
	}
	if len(screen.Containers) != 1 {
		t.Fatalf("screen.containers: want 1, got %d", len(screen.Containers))
	}
	btn := screen.Containers[0]
	if len(btn.Containers) != 1 {
		t.Fatalf("button.containers: want 1, got %d", len(btn.Containers))
	}
	if btn.Containers[0].Title != "Icon" {
		t.Errorf("nested ref title: want %q, got %q", "Icon", btn.Containers[0].Title)
	}
}

func TestLoad_RefSetsPath(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "shared", "nav.yml"), `title: Nav`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./shared/nav.yml
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var screen build.Container
	for _, c := range containers {
		if c.Title == "Screen" {
			screen = c
		}
	}
	nav := screen.Containers[0]
	if nav.Path != filepath.Join("shared", "nav.yml") {
		t.Errorf("ref path: want %q, got %q", filepath.Join("shared", "nav.yml"), nav.Path)
	}
}
