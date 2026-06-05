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

func TestResolveStateSpecs_Inheritance(t *testing.T) {
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
	c, err := spec.LoadContainerTree(filepath.Join(root, "card.yml"), root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := spec.WriteBuiltContainer(filepath.Join(root, "out.yml"), c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, "out.yml"))
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		States map[string]struct {
			Specs map[string]string `yaml:"specs"`
		} `yaml:"states"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	if got := out.States["default"].Specs["bgColor"]; got != "white" {
		t.Errorf("default bgColor: want 'white', got %q", got)
	}
	if got := out.States["dark"].Specs["bgColor"]; got != "#1A1A1A" {
		t.Errorf("dark bgColor: want '#1A1A1A', got %q", got)
	}
	if got := out.States["dark"].Specs["textColor"]; got != "black" {
		t.Errorf("dark textColor (inherited): want 'black', got %q", got)
	}
}

func TestWriteBuiltContainer_Sections(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.yml")

	c := spec.Container{
		Title:       "Jot",
		Description: "Note-taking screen",
		States: []spec.StateSpec{
			{Ref: "default", Specs: map[string]spec.SpecValue{
				"bgColor": {Value: "white"},
			}},
		},
		Containers: []spec.Container{
			{
				Title: "Note Input",
				States: []spec.StateSpec{
					{Ref: "default", Specs: map[string]spec.SpecValue{
						"formLabelText":  {Value: "Notes"},
						"footerHelpText": {Value: "enter to submit", Description: "shown in footer"},
					}},
				},
			},
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c); err != nil {
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

	c := spec.Container{
		Title: "Jot",
		Containers: []spec.Container{
			{
				Title: "Note Input",
				States: []spec.StateSpec{
					{Ref: "default", Specs: map[string]spec.SpecValue{
						"formLabelText": {Value: "Notes"},
					}},
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
		},
	}

	if err := spec.WriteBuiltContainer(outPath, c); err != nil {
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

	if err := spec.WriteBuiltContainer(outPath, c); err != nil {
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

func TestWriteBuiltContainer_RefResolutionAndOverrides(t *testing.T) {
	dir := t.TempDir()
	buildFixture(t, dir, "shared/hero.yml", `
title: Hero
specs:
  background-color: red
  headerText: Hello
`)
	buildFixture(t, dir, "jot/index.yml", `
containers:
- $ref: '../shared/hero.yml'
  background-color: blue
`)

	c, err := spec.LoadContainerTree(filepath.Join(dir, "jot/index.yml"), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outPath := filepath.Join(dir, "out.yml")
	if err := spec.WriteBuiltContainer(outPath, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	var out struct {
		Containers []struct {
			Title  string `yaml:"title"`
			States map[string]struct {
				Specs map[string]string `yaml:"specs"`
			} `yaml:"states"`
		} `yaml:"containers"`
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	if len(out.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(out.Containers))
	}
	if out.Containers[0].Title != "Hero" {
		t.Errorf("title: want 'Hero', got %q", out.Containers[0].Title)
	}
	if got := out.Containers[0].States["default"].Specs["background-color"]; got != "blue" {
		t.Errorf("background-color override: want 'blue', got %q", got)
	}
	if got := out.Containers[0].States["default"].Specs["headerText"]; got != "Hello" {
		t.Errorf("headerText (inherited): want 'Hello', got %q", got)
	}
}
