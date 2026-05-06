package spec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func TestParseContainer_ImplicitDefault(t *testing.T) {
	f := writeTempFile(t, `
specs:
  background-color: "#FFFFFF"
  title: "Hello"
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(c.States))
	}
	if c.States[0].Ref != "default" {
		t.Errorf("expected ref 'default', got %q", c.States[0].Ref)
	}
	if v := c.States[0].Specs["background-color"].Value; v != "#FFFFFF" {
		t.Errorf("expected background-color '#FFFFFF', got %v", v)
	}
}

func TestParseContainer_VerboseSpec(t *testing.T) {
	f := writeTempFile(t, `
specs:
  title:
    value: "Hello"
    description: "The main heading"
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	titleSpec := c.States[0].Specs["title"]
	if titleSpec.Value != "Hello" {
		t.Errorf("expected value 'Hello', got %v", titleSpec.Value)
	}
	if titleSpec.Description != "The main heading" {
		t.Errorf("expected description 'The main heading', got %q", titleSpec.Description)
	}
}

func TestParseContainer_ExplicitStates(t *testing.T) {
	f := writeTempFile(t, `
default: baseline
states:
  - ref: baseline
    specs:
      background-color: "#FFFFFF"
  - ref: dark-mode
    specs:
      background-color: "#000000"
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Default != "baseline" {
		t.Errorf("expected default 'baseline', got %q", c.Default)
	}
	if len(c.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(c.States))
	}
}

func TestParseContainer_TopLevelSpecsFoldIntoNamedDefault(t *testing.T) {
	f := writeTempFile(t, `
default: baseline
specs:
  background-color: "#FFFFFF"
states:
  - ref: dark-mode
    specs:
      background-color: "#000000"
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// should have baseline (from top-level specs) + dark-mode
	if len(c.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(c.States))
	}
	var baseline *spec.StateSpec
	for i := range c.States {
		if c.States[i].Ref == "baseline" {
			baseline = &c.States[i]
		}
	}
	if baseline == nil {
		t.Fatal("expected a 'baseline' state")
	}
	if v := baseline.Specs["background-color"].Value; v != "#FFFFFF" {
		t.Errorf("expected background-color '#FFFFFF', got %v", v)
	}
}

func TestParseContainer_Events(t *testing.T) {
	f := writeTempFile(t, `
states:
  - ref: default
    specs:
      background-color: "#FFFFFF"
    events:
      on-tap:
        action: navigate
        destination: detail-screen
      on-long-press: show-menu
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	events := c.States[0].Events
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events["on-long-press"] != "show-menu" {
		t.Errorf("expected on-long-press 'show-menu', got %v", events["on-long-press"])
	}
	tap, ok := events["on-tap"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected on-tap to be a map, got %T", events["on-tap"])
	}
	if tap["action"] != "navigate" {
		t.Errorf("expected action 'navigate', got %v", tap["action"])
	}
}

func TestParseContainer_EmptyEvents(t *testing.T) {
	f := writeTempFile(t, `
states:
  - ref: default
    specs:
      background-color: "#FFFFFF"
    events: {}
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States[0].Events) != 0 {
		t.Errorf("expected empty events, got %v", c.States[0].Events)
	}
}

func TestParseContainer_Imports(t *testing.T) {
	f := writeTempFile(t, `
specs:
  title: Hello
containers:
  - $ref: './noteInput.yml'
  - $ref: '../shared/hero.yml'
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(c.Imports))
	}
	if c.Imports[0].Ref != "./noteInput.yml" {
		t.Errorf("expected './noteInput.yml', got %q", c.Imports[0].Ref)
	}
	if c.Imports[1].Ref != "../shared/hero.yml" {
		t.Errorf("expected '../shared/hero.yml', got %q", c.Imports[1].Ref)
	}
}

func TestParseContainer_StateLineNumbers(t *testing.T) {
	f := writeTempFile(t, `states:
  - ref: baseline
    specs:
      color: red
  - ref: dark-mode
    specs:
      color: black
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States) != 2 {
		t.Fatalf("expected 2 states, got %d", len(c.States))
	}
	if c.States[0].Line == 0 {
		t.Error("expected non-zero line for first state")
	}
	if c.States[1].Line <= c.States[0].Line {
		t.Errorf("expected dark-mode line (%d) > baseline line (%d)", c.States[1].Line, c.States[0].Line)
	}
}

func TestParseContainer_ImportLineNumbers(t *testing.T) {
	f := writeTempFile(t, `specs:
  title: Hello
containers:
  - $ref: './a.yml'
  - $ref: './b.yml'
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Imports) != 2 {
		t.Fatalf("expected 2 imports, got %d", len(c.Imports))
	}
	if c.Imports[0].Line == 0 {
		t.Error("expected non-zero line for first import")
	}
	if c.Imports[1].Line <= c.Imports[0].Line {
		t.Errorf("expected second import line (%d) > first (%d)", c.Imports[1].Line, c.Imports[0].Line)
	}
}

func TestParseContainer_NoImports(t *testing.T) {
	f := writeTempFile(t, `
specs:
  title: Hello
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Imports) != 0 {
		t.Errorf("expected no imports, got %d", len(c.Imports))
	}
}

func TestParseContainer_TopLevelEvents(t *testing.T) {
	f := writeTempFile(t, `
events:
  - title: on-press-enter
    description: user presses enter
    actions:
      navigate:
        destination: /jot/tags
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(c.Events))
	}
	ev := c.Events[0]
	if ev.Title != "on-press-enter" {
		t.Errorf("title: want 'on-press-enter', got %q", ev.Title)
	}
	if ev.Description != "user presses enter" {
		t.Errorf("description: want 'user presses enter', got %q", ev.Description)
	}
	if ev.Actions["navigate"]["destination"] != "/jot/tags" {
		t.Errorf("navigate.destination: want '/jot/tags', got %v", ev.Actions["navigate"]["destination"])
	}
}

func TestParseContainer_TopLevelEvents_MultipleActions(t *testing.T) {
	f := writeTempFile(t, `
events:
  - title: on-press-enter
    actions:
      navigate:
        destination: /jot/tags
      track:
        event: note_submitted
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ev := c.Events[0]
	if len(ev.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(ev.Actions))
	}
	if ev.Actions["track"]["event"] != "note_submitted" {
		t.Errorf("track.event: want 'note_submitted', got %v", ev.Actions["track"]["event"])
	}
}

func TestParseContainer_TopLevelEvents_NoPayload(t *testing.T) {
	f := writeTempFile(t, `
events:
  - title: on-dismiss
    actions:
      dismiss:
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(c.Events))
	}
	if _, ok := c.Events[0].Actions["dismiss"]; !ok {
		t.Error("expected 'dismiss' action to be present")
	}
}

func TestLoadContainers(t *testing.T) {
	dir := t.TempDir()

	// leaf container
	leaf := filepath.Join(dir, "app", "home-tab", "feed-screen")
	if err := os.MkdirAll(leaf, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(leaf, "post-card.yml"), `
specs:
  background-color: "#FFFFFF"
`)

	// non-leaf with index.yml
	writeFile(t, filepath.Join(dir, "app", "index.yml"), `
states:
  - ref: default
    specs:
      font-family: "Inter"
`)

	containers, err := spec.LoadContainers(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}
}
