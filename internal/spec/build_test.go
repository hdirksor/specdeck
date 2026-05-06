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

	sections := spec.ResolveContainerSections(c, root, byPath)

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Title != "Hero" {
		t.Errorf("expected title 'Hero', got %q", sections[0].Title)
	}
	if sections[0].Specs["headerText"].Value != "Hello" {
		t.Errorf("unexpected spec value: %v", sections[0].Specs["headerText"])
	}
}

func TestResolveContainerSections_TitleFallsBackToFilename(t *testing.T) {
	root := t.TempDir()
	buildFixture(t, root, "noteInput.yml", `
specs:
  formLabelText: Notes
`)
	c := loadFixture(t, root, "noteInput.yml")
	byPath := map[string]spec.Container{"noteInput.yml": c}

	sections := spec.ResolveContainerSections(c, root, byPath)

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if sections[0].Title != "noteInput" {
		t.Errorf("expected title 'noteInput', got %q", sections[0].Title)
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

	sections := spec.ResolveContainerSections(index, root, byPath)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].Title != "Hero" {
		t.Errorf("section 0 title: want 'Hero', got %q", sections[0].Title)
	}
	if sections[0].Specs["headerText"].Value != "Hello" {
		t.Errorf("section 0 specs: unexpected value %v", sections[0].Specs["headerText"])
	}
	if sections[1].Title != "Note Input" {
		t.Errorf("section 1 title: want 'Note Input', got %q", sections[1].Title)
	}
	if sections[1].Specs["formLabelText"].Value != "Notes" {
		t.Errorf("section 1 specs: unexpected value %v", sections[1].Specs["formLabelText"])
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

	sections := spec.ResolveContainerSections(screen, root, byPath)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].Title != "Screen" {
		t.Errorf("section 0 (own) title: want 'Screen', got %q", sections[0].Title)
	}
	if sections[0].Specs["bgColor"].Value != "white" {
		t.Errorf("section 0 specs: unexpected value %v", sections[0].Specs["bgColor"])
	}
	if sections[1].Title != "Footer" {
		t.Errorf("section 1 (import) title: want 'Footer', got %q", sections[1].Title)
	}
}

func TestWriteBuiltContainer_Sections(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{Title: "Jot", Description: "Note-taking screen"}
	sections := []spec.Section{
		{
			Title: "Hero",
			Specs: map[string]spec.SpecValue{"headerText": {Value: "░ jot"}},
		},
		{
			Title: "Note Input",
			Specs: map[string]spec.SpecValue{
				"formLabelText": {Value: "Notes"},
				"footerHelpText": {Value: "enter to submit", Description: "shown in footer"},
			},
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c, sections); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
		Sections    []struct {
			Title string                 `yaml:"title"`
			Specs map[string]interface{} `yaml:"specs"`
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
	if len(out.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(out.Sections))
	}
	if out.Sections[0].Title != "Hero" {
		t.Errorf("section 0 title: want 'Hero', got %q", out.Sections[0].Title)
	}
	if out.Sections[1].Title != "Note Input" {
		t.Errorf("section 1 title: want 'Note Input', got %q", out.Sections[1].Title)
	}
	if out.Sections[1].Specs["footerHelpText"] == nil {
		t.Error("expected footerHelpText in section 1 specs")
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

	sections := spec.ResolveContainerSections(index, root, byPath)

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if got := sections[0].Specs["background-color"].Value; got != "blue" {
		t.Errorf("background-color: want 'blue', got %q", got)
	}
	if got := sections[0].Specs["headerText"].Value; got != "Hello" {
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

	sections := spec.ResolveContainerSections(index, root, byPath)

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if len(sections[0].Events) != 2 {
		t.Fatalf("expected 2 events in section, got %d", len(sections[0].Events))
	}
	if sections[0].Events[0].Title != "on-press-alt-e" {
		t.Errorf("event 0 title: want 'on-press-alt-e', got %q", sections[0].Events[0].Title)
	}
	if sections[0].Events[1].Title != "on-type-space-hash" {
		t.Errorf("event 1 title: want 'on-type-space-hash', got %q", sections[0].Events[1].Title)
	}
}

func TestWriteBuiltContainer_SectionEvents(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{Title: "Jot"}
	sections := []spec.Section{
		{
			Title: "Note Input",
			Specs: map[string]spec.SpecValue{"formLabelText": {Value: "Notes"}},
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

	if err := spec.WriteBuiltContainer(outPath, c, sections); err != nil {
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

	if err := spec.WriteBuiltContainer(outPath, c, nil); err != nil {
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
