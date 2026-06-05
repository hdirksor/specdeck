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

func TestParseContainer_ContainerRefs(t *testing.T) {
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
	if len(c.Containers) != 2 {
		t.Fatalf("expected 2 container stubs, got %d", len(c.Containers))
	}
}

func TestParseContainer_NoContainers(t *testing.T) {
	f := writeTempFile(t, `
specs:
  title: Hello
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Containers) != 0 {
		t.Errorf("expected no containers, got %d", len(c.Containers))
	}
}

func TestParseContainer_InlineContainer(t *testing.T) {
	f := writeTempFile(t, `
title: Screen
containers:
  - title: Header
    specs:
      label: My App
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Containers) != 1 {
		t.Fatalf("expected 1 inline container, got %d", len(c.Containers))
	}
	if c.Containers[0].Title != "Header" {
		t.Errorf("expected title 'Header', got %q", c.Containers[0].Title)
	}
	if v := c.Containers[0].States[0].Specs["label"].Value; v != "My App" {
		t.Errorf("expected label 'My App', got %q", v)
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

func TestParseContainer_SequenceSpecs(t *testing.T) {
	f := writeTempFile(t, `
specs:
  - border: 2dp white
  - content-padding: 16dp on all sides
  - scroll-direction: vertical
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(c.States))
	}
	specs := c.States[0].Specs
	if v := specs["border"].Value; v != "2dp white" {
		t.Errorf("border: want '2dp white', got %v", v)
	}
	if v := specs["content-padding"].Value; v != "16dp on all sides" {
		t.Errorf("content-padding: want '16dp on all sides', got %v", v)
	}
	if v := specs["scroll-direction"].Value; v != "vertical" {
		t.Errorf("scroll-direction: want 'vertical', got %v", v)
	}
}

func TestParseContainer_ExplicitStates_SequenceSpecs(t *testing.T) {
	f := writeTempFile(t, `
states:
  - ref: default
    specs:
      - color: red
      - font-size: 14sp
`)
	c, err := spec.ParseContainer(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.States) != 1 {
		t.Fatalf("expected 1 state, got %d", len(c.States))
	}
	specs := c.States[0].Specs
	if v := specs["color"].Value; v != "red" {
		t.Errorf("color: want 'red', got %v", v)
	}
	if v := specs["font-size"].Value; v != "14sp" {
		t.Errorf("font-size: want '14sp', got %v", v)
	}
}

func TestLoadContainers(t *testing.T) {
	dir := t.TempDir()

	leaf := filepath.Join(dir, "app", "home-tab", "feed-screen")
	if err := os.MkdirAll(leaf, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(leaf, "post-card.yml"), `
specs:
  background-color: "#FFFFFF"
`)

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

func TestLoadContainerTree_ResolvesRef(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "hero.yml"), `
title: Hero
specs:
  headerText: Hello
`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./hero.yml
`)

	c, err := spec.LoadContainerTree(filepath.Join(dir, "screen.yml"), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Containers) != 1 {
		t.Fatalf("expected 1 resolved container, got %d", len(c.Containers))
	}
	if c.Containers[0].Title != "Hero" {
		t.Errorf("expected title 'Hero', got %q", c.Containers[0].Title)
	}
	if v := c.Containers[0].States[0].Specs["headerText"].Value; v != "Hello" {
		t.Errorf("expected headerText 'Hello', got %q", v)
	}
}

func TestLoadContainerTree_AppliesOverrides(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "button.yml"), `
title: Button
specs:
  label: Click
  color: blue
`)
	writeFile(t, filepath.Join(dir, "screen.yml"), `
title: Screen
containers:
  - $ref: ./button.yml
    label: OK
`)

	c, err := spec.LoadContainerTree(filepath.Join(dir, "screen.yml"), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	btn := c.Containers[0]
	if v := btn.States[0].Specs["label"].Value; v != "OK" {
		t.Errorf("override label: want 'OK', got %q", v)
	}
	if v := btn.States[0].Specs["color"].Value; v != "blue" {
		t.Errorf("inherited color: want 'blue', got %q", v)
	}
}
