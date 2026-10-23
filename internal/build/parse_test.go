package build_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdirksor/specdeck/internal/build"
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

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err, want)
	}
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
    effects:
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
	if len(ev.Effects) != 2 {
		t.Fatalf("effects: want 2, got %d", len(ev.Effects))
	}
	if ev.Effects[0].Title != "save" {
		t.Errorf("effect[0] title: want %q, got %q", "save", ev.Effects[0].Title)
	}
	if ev.Effects[0].Description != "Persist the form data." {
		t.Errorf("effect[0] description: want %q, got %q", "Persist the form data.", ev.Effects[0].Description)
	}
}

func TestParseFile_EffectShorthand(t *testing.T) {
	f := writeTempFile(t, `
title: Note Card
events:
  - title: delete
    description: User taps delete.
    effects:
      - Permanently deletes this note from storage.
      - title: refresh-list
        description: Updates the note list.
`)
	c, err := build.ParseFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ev := c.Events[0]
	if len(ev.Effects) != 2 {
		t.Fatalf("effects: want 2, got %d", len(ev.Effects))
	}
	if ev.Effects[0].Title != "" {
		t.Errorf("effect[0] title: want empty, got %q", ev.Effects[0].Title)
	}
	if ev.Effects[0].Description != "Permanently deletes this note from storage." {
		t.Errorf("effect[0] description: want %q, got %q", "Permanently deletes this note from storage.", ev.Effects[0].Description)
	}
	if ev.Effects[1].Title != "refresh-list" {
		t.Errorf("effect[1] title: want %q, got %q", "refresh-list", ev.Effects[1].Title)
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

func TestLoad_SetsImage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screen.yml"), `title: Screen`)
	writeFile(t, filepath.Join(dir, "screen.png"), `fake-png-bytes`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 1 {
		t.Fatalf("want 1 container, got %d", len(containers))
	}
	if containers[0].Image != "screen.png" {
		t.Errorf("image: want %q, got %q", "screen.png", containers[0].Image)
	}
}

func TestLoad_NoImageWhenNoSiblingFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screen.yml"), `title: Screen`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if containers[0].Image != "" {
		t.Errorf("image: want empty, got %q", containers[0].Image)
	}
}

func TestLoad_AmbiguousImageErrors(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screen.yml"), `title: Screen`)
	writeFile(t, filepath.Join(dir, "screen.png"), `fake-png-bytes`)
	writeFile(t, filepath.Join(dir, "screen.svg"), `fake-svg-bytes`)

	_, err := build.Load(dir)
	if err == nil {
		t.Fatal("expected error for ambiguous image, got nil")
	}
}

func TestLoad_NestedRefHasNoImage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hero.yml"), `title: Hero`)
	writeFile(t, filepath.Join(dir, "hero.png"), `fake-png-bytes`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./hero.yml
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var hero, screen build.Container
	for _, c := range containers {
		switch c.Title {
		case "Hero":
			hero = c
		case "Screen":
			screen = c
		}
	}
	if hero.Image != "hero.png" {
		t.Errorf("standalone hero image: want %q, got %q", "hero.png", hero.Image)
	}
	if len(screen.Containers) != 1 {
		t.Fatalf("screen.containers: want 1, got %d", len(screen.Containers))
	}
	if screen.Containers[0].Image != "" {
		t.Errorf("nested hero image: want empty, got %q", screen.Containers[0].Image)
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

func TestLoad_InlineContainerResolvesNestedRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "shared", "icon.yml"), `title: Icon`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - title: Toolbar
    containers:
      - $ref: ./shared/icon.yml
`)

	containers, err := build.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	toolbar := containers[0].Containers[0]
	if toolbar.Title != "Toolbar" {
		t.Fatalf("inline title: want %q, got %q", "Toolbar", toolbar.Title)
	}
	if len(toolbar.Containers) != 1 || toolbar.Containers[0].Title != "Icon" {
		t.Errorf("inline container ref not resolved, got %+v", toolbar.Containers)
	}
}

func TestLoad_OverridesOnRefWithoutSpecs(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "divider.yml"), `title: Divider`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./divider.yml
    overrides:
      color: gray
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
	if got := screen.Containers[0].Specs["color"]; got != "gray" {
		t.Errorf("override color: want %q, got %q", "gray", got)
	}
}

func TestLoad_MissingRoot(t *testing.T) {
	_, err := build.Load(filepath.Join(t.TempDir(), "containers"))
	if err == nil {
		t.Fatal("expected error for missing root, got nil")
	}
}

func TestLoad_InvalidYAMLNamesFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screens", "bad.yml"), `title: [unclosed`)

	_, err := build.Load(dir)
	assertErrorContains(t, err, filepath.Join("screens", "bad.yml"))
}

func TestLoad_MissingRefNamesRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./missing.yml
`)

	_, err := build.Load(dir)
	assertErrorContains(t, err, "screen.yml")
	assertErrorContains(t, err, "loading ./missing.yml")
}

func TestLoad_InvalidYAMLInRefNamesRef(t *testing.T) {
	dir := t.TempDir()
	// Load walks files in lexical order; name the referrer so it is visited first.
	writeFile(t, filepath.Join(dir, "a-screen.yml"), `
title: Screen
containers:
  - $ref: ./z-broken.yml
`)
	writeFile(t, filepath.Join(dir, "z-broken.yml"), `title: [unclosed`)

	_, err := build.Load(dir)
	assertErrorContains(t, err, "a-screen.yml")
	assertErrorContains(t, err, "loading ./z-broken.yml")
}

func TestLoad_MissingRefInsideInlineContainer(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - title: Toolbar
    containers:
      - $ref: ./missing.yml
`)

	_, err := build.Load(dir)
	assertErrorContains(t, err, "loading ./missing.yml")
}
