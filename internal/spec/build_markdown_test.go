package spec_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdickson/specdeck/internal/spec"
)

func readMarkdown(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading markdown output: %v", err)
	}
	return string(data)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q\ngot:\n%s", want, got)
	}
}

func TestWriteBuiltContainerMarkdown_TitleAndDescription(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	c := spec.Container{Title: "Jot", Description: "Note-taking screen"}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	assertContains(t, got, "# Jot")
	assertContains(t, got, "Note-taking screen")
}

func TestWriteBuiltContainerMarkdown_SpecTable(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	c := spec.Container{Title: "Card"}
	ownStates := map[string]map[string]spec.SpecValue{
		"default": {
			"bgColor":   {Value: "white"},
			"textColor": {Value: "black"},
		},
	}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, ownStates, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	assertContains(t, got, "## default")
	assertContains(t, got, "| Spec | Value |")
	assertContains(t, got, "| bgColor | white |")
	assertContains(t, got, "| textColor | black |")
}

func TestWriteBuiltContainerMarkdown_MultiState_DefaultFirst(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	c := spec.Container{Title: "Card"}
	ownStates := map[string]map[string]spec.SpecValue{
		"default": {"bgColor": {Value: "white"}, "textColor": {Value: "black"}},
		"dark":    {"bgColor": {Value: "#1A1A1A"}, "textColor": {Value: "white"}},
	}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, ownStates, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	// ## default section must appear before ## dark section
	defaultIdx := strings.Index(got, "## default")
	darkIdx := strings.Index(got, "## dark")
	if defaultIdx == -1 || darkIdx == -1 {
		t.Fatalf("expected both '## default' and '## dark' sections; got:\n%s", got)
	}
	if defaultIdx > darkIdx {
		t.Errorf("expected '## default' before '## dark'")
	}
	assertContains(t, got, "| bgColor | white |")
	assertContains(t, got, "| bgColor | #1A1A1A |")
}

func TestWriteBuiltContainerMarkdown_ImportedSection(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	c := spec.Container{Title: "Jot"}
	imports := []spec.Section{
		{
			Title: "Note Input",
			States: map[string]map[string]spec.SpecValue{
				"default": {"formLabelText": {Value: "Notes"}},
			},
		},
	}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, nil, imports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	assertContains(t, got, "## Note Input")
	assertContains(t, got, "### default")
	assertContains(t, got, "| formLabelText | Notes |")
}

func TestWriteBuiltContainerMarkdown_OwnEvents(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	c := spec.Container{
		Title: "Jot",
		Events: []spec.Event{
			{
				Title:       "on-press-enter",
				Description: "user presses enter",
				Actions: map[string]spec.Action{
					"navigate": {"destination": "/jot/tags"},
				},
			},
		},
	}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	assertContains(t, got, "## Events")
	assertContains(t, got, "### on-press-enter")
	assertContains(t, got, "user presses enter")
	assertContains(t, got, "**navigate**")
	assertContains(t, got, "destination: /jot/tags")
}

func TestWriteBuiltContainerMarkdown_SectionEvents(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

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
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, nil, imports); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := readMarkdown(t, outPath)
	assertContains(t, got, "## Note Input")
	assertContains(t, got, "### default")
	assertContains(t, got, "### Events")
	assertContains(t, got, "#### on-press-alt-e")
	assertContains(t, got, "**open**")
}

func TestWriteBuiltContainerMarkdown_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nested", "deep", "out.md")

	c := spec.Container{Title: "X"}
	if err := spec.WriteBuiltContainerMarkdown(outPath, c, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}

func TestWriteHugoSectionStub_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	stubPath := filepath.Join(dir, "_index.md")
	if err := spec.WriteHugoSectionStub(stubPath, "Shared"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(stubPath); err != nil {
		t.Errorf("expected stub file to exist: %v", err)
	}
}

func TestWriteHugoSectionStub_Content(t *testing.T) {
	dir := t.TempDir()
	stubPath := filepath.Join(dir, "_index.md")
	if err := spec.WriteHugoSectionStub(stubPath, "My Section"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := readMarkdown(t, stubPath)
	assertContains(t, got, "title: My Section")
}

func TestWriteHugoSectionStub_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	stubPath := filepath.Join(dir, "nested", "_index.md")
	if err := spec.WriteHugoSectionStub(stubPath, "Nested"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(stubPath); err != nil {
		t.Errorf("expected stub file to exist: %v", err)
	}
}
