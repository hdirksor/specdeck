package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
	"gopkg.in/yaml.v3"
)

// buildFixture writes a minimal container YAML file and returns its path.
func buildFixture(t *testing.T, dir, rel, content string) string {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return full
}

func loadFixture(t *testing.T, root, rel string) spec.Container {
	t.Helper()
	c, err := spec.ParseContainer(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("ParseContainer %s: %v", rel, err)
	}
	c.Path = rel
	return c
}

func TestResolveContainerSections_OwnSpecsOnly(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "hero.yml", `
title: Hero
specs:
  headerText: Hello
`)
	c := loadFixture(t, root, "hero.yml")
	byPath := map[string]spec.Container{"hero.yml": c}

	ownStates, imports := spec.ResolveContainerSections(c, root, byPath)

	if len(imports) != 0 {
		t.Fatalf("expected no imports, got %d", len(imports))
	}
	if len(ownStates) != 1 {
		t.Fatalf("expected 1 state, got %d", len(ownStates))
	}
	if ownStates["default"]["headerText"].Value != "Hello" {
		t.Errorf("unexpected spec value: %v", ownStates["default"]["headerText"])
	}
}

func TestResolveContainerSections_StateInheritance(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "card.yml", `
title: Card
specs:
  bgColor: white
  textColor: black
states:
  - ref: dark
    specs:
      bgColor: "#1A1A1A"
`)
	c := loadFixture(t, root, "card.yml")
	byPath := map[string]spec.Container{"card.yml": c}

	ownStates, _ := spec.ResolveContainerSections(c, root, byPath)

	if len(ownStates) != 2 {
		t.Fatalf("expected 2 states, got %d", len(ownStates))
	}
	// Default state unchanged.
	if got := ownStates["default"]["bgColor"].Value; got != "white" {
		t.Errorf("default bgColor: want 'white', got %q", got)
	}
	// Dark state overrides bgColor, inherits textColor.
	if got := ownStates["dark"]["bgColor"].Value; got != "#1A1A1A" {
		t.Errorf("dark bgColor: want '#1A1A1A', got %q", got)
	}
	if got := ownStates["dark"]["textColor"].Value; got != "black" {
		t.Errorf("dark textColor: want 'black' (inherited), got %q", got)
	}
}

func TestResolveContainerSections_ImportsOnly(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "shared/hero.yml", `
title: Hero
specs:
  headerText: Hello
`)
	buildFixture(t, root, "jot/noteInput.yml", `
title: Note Input
specs:
  formLabelText: Notes
`)
	buildFixture(t, root, "jot/index.yml", `
containers:
- $ref: '../shared/hero.yml'
- $ref: './noteInput.yml'
`)
	hero := loadFixture(t, root, "shared/hero.yml")
	note := loadFixture(t, root, "jot/noteInput.yml")
	index := loadFixture(t, root, "jot/index.yml")
	byPath := map[string]spec.Container{
		"shared/hero.yml":   hero,
		"jot/noteInput.yml": note,
		"jot/index.yml":     index,
	}

	ownStates, imports := spec.ResolveContainerSections(index, root, byPath)

	if len(ownStates) != 0 {
		t.Fatalf("expected no own states, got %d", len(ownStates))
	}
	if len(imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(imports))
	}
	if imports[0].Title != "Hero" {
		t.Errorf("import 0 title: want 'Hero', got %q", imports[0].Title)
	}
	if imports[0].States["default"]["headerText"].Value != "Hello" {
		t.Errorf("import 0 specs: unexpected value %v", imports[0].States["default"]["headerText"])
	}
	if imports[1].Title != "Note Input" {
		t.Errorf("import 1 title: want 'Note Input', got %q", imports[1].Title)
	}
	if imports[1].States["default"]["formLabelText"].Value != "Notes" {
		t.Errorf("import 1 specs: unexpected value %v", imports[1].States["default"]["formLabelText"])
	}
}

func TestResolveContainerSections_OwnSpecsAndImports(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "shared/footer.yml", `
title: Footer
specs:
  footerText: Done
`)
	buildFixture(t, root, "screen.yml", `
title: Screen
specs:
  bgColor: white
containers:
- $ref: './shared/footer.yml'
`)
	footer := loadFixture(t, root, "shared/footer.yml")
	screen := loadFixture(t, root, "screen.yml")
	byPath := map[string]spec.Container{
		"shared/footer.yml": footer,
		"screen.yml":        screen,
	}

	ownStates, imports := spec.ResolveContainerSections(screen, root, byPath)

	if ownStates["default"]["bgColor"].Value != "white" {
		t.Errorf("own bgColor: want 'white', got %v", ownStates["default"]["bgColor"])
	}
	if len(imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(imports))
	}
	if imports[0].Title != "Footer" {
		t.Errorf("import title: want 'Footer', got %q", imports[0].Title)
	}
}

func TestResolveContainerSections_InlineOverrides(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "shared/hero.yml", `
title: Hero
specs:
  background-color: red
  headerText: Hello
`)
	buildFixture(t, root, "jot/index.yml", `
containers:
- $ref: '../shared/hero.yml'
  background-color: blue
`)
	hero := loadFixture(t, root, "shared/hero.yml")
	index := loadFixture(t, root, "jot/index.yml")
	byPath := map[string]spec.Container{
		"shared/hero.yml": hero,
		"jot/index.yml":   index,
	}

	_, imports := spec.ResolveContainerSections(index, root, byPath)

	if len(imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(imports))
	}
	if got := imports[0].States["default"]["background-color"].Value; got != "blue" {
		t.Errorf("background-color: want 'blue', got %q", got)
	}
	if got := imports[0].States["default"]["headerText"].Value; got != "Hello" {
		t.Errorf("headerText: want 'Hello', got %q", got)
	}
}

func TestResolveContainerSections_ImportedEvents(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "jot/noteInput.yml", `
title: Note Input
specs:
  formLabelText: Notes
events:
  - title: on-press-alt-e
    actions:
      open:
        description: open editor
  - title: on-type-space-hash
    description: triggers autocomplete
    actions:
      dispatch:
        job: activate-autocomplete-tags
`)
	buildFixture(t, root, "jot/index.yml", `
containers:
- $ref: './noteInput.yml'
`)
	note := loadFixture(t, root, "jot/noteInput.yml")
	index := loadFixture(t, root, "jot/index.yml")
	byPath := map[string]spec.Container{
		"jot/noteInput.yml": note,
		"jot/index.yml":     index,
	}

	_, imports := spec.ResolveContainerSections(index, root, byPath)

	if len(imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(imports))
	}
	if len(imports[0].Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(imports[0].Events))
	}
	if imports[0].Events[0].Title != "on-press-alt-e" {
		t.Errorf("event 0 title: want 'on-press-alt-e', got %q", imports[0].Events[0].Title)
	}
	if imports[0].Events[1].Title != "on-type-space-hash" {
		t.Errorf("event 1 title: want 'on-type-space-hash', got %q", imports[0].Events[1].Title)
	}
}

func TestWriteBuiltContainer_Sections(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{Title: "Jot", Description: "Note-taking screen"}
	ownStates := map[string]map[string]spec.SpecValue{
		"default": {"bgColor": {Value: "white"}},
	}
	imports := []spec.Section{
		{
			Title: "Note Input",
			States: map[string]map[string]spec.SpecValue{
				"default": {
					"formLabelText":  {Value: "Notes"},
					"footerHelpText": {Value: "enter to submit", Description: "shown in footer"},
				},
			},
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c, ownStates, imports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		States      map[string]struct {
			Specs map[string]interface{} `yaml:"specs"`
		} `yaml:"states"`
		Sections []struct {
			Title  string `yaml:"title"`
			States map[string]struct {
				Specs map[string]interface{} `yaml:"specs"`
			} `yaml:"states"`
		} `yaml:"containers"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	if out.Title != "Jot" {
		t.Errorf("title: want 'Jot', got %q", out.Title)
	}
	if out.Description != "Note-taking screen" {
		t.Errorf("description: want 'Note-taking screen', got %q", out.Description)
	}
	if out.States["default"].Specs["bgColor"] != "white" {
		t.Errorf("own default bgColor: want 'white', got %v", out.States["default"].Specs["bgColor"])
	}
	if len(out.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(out.Sections))
	}
	if out.Sections[0].Title != "Note Input" {
		t.Errorf("section title: want 'Note Input', got %q", out.Sections[0].Title)
	}
	if out.Sections[0].States["default"].Specs["footerHelpText"] == nil {
		t.Error("expected footerHelpText in section default state")
	}
}

func TestWriteBuiltContainer_SectionEvents(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{Title: "Jot"}
	imports := []spec.Section{
		{
			Title: "Note Input",
			States: map[string]map[string]spec.SpecValue{
				"default": {"formLabelText": {Value: "Notes"}},
			},
			Events: []spec.Event{
				{
					Title: "on-press-alt-e",
					Actions: map[string]spec.Action{
						"open": {"description": "open editor"},
					},
				},
			},
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c, nil, imports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Sections []struct {
			Title  string `yaml:"title"`
			Events []struct {
				Title string `yaml:"title"`
			} `yaml:"events"`
		} `yaml:"containers"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	if len(out.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(out.Sections))
	}
	if len(out.Sections[0].Events) != 1 {
		t.Fatalf("expected 1 event in section, got %d", len(out.Sections[0].Events))
	}
	if out.Sections[0].Events[0].Title != "on-press-alt-e" {
		t.Errorf("event title: want 'on-press-alt-e', got %q", out.Sections[0].Events[0].Title)
	}
}

func TestWriteBuiltContainer_Events(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{
		Title: "Jot",
		Events: []spec.Event{
			{
				Title:       "on-press-enter",
				Description: "user presses enter",
				Actions: map[string]spec.Action{
					"navigate": {"destination": "/jot/tags"},
					"track":    {"event": "note_submitted"},
				},
			},
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Events []struct {
			Title       string                 `yaml:"title"`
			Description string                 `yaml:"description"`
			Actions     map[string]interface{} `yaml:"actions"`
		} `yaml:"events"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	if len(out.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(out.Events))
	}
	ev := out.Events[0]
	if ev.Title != "on-press-enter" {
		t.Errorf("title: want 'on-press-enter', got %q", ev.Title)
	}
	if ev.Description != "user presses enter" {
		t.Errorf("description: want 'user presses enter', got %q", ev.Description)
	}
	if len(ev.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(ev.Actions))
	}
}
