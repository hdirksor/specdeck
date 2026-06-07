package build_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hdickson/specdeck/internal/build"
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	return string(data)
}

func assertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("expected output to contain %q\ngot:\n%s", want, got)
	}
}

func assertNotContains(t *testing.T, got, want string) {
	t.Helper()
	if strings.Contains(got, want) {
		t.Errorf("expected output NOT to contain %q\ngot:\n%s", want, got)
	}
}

func writeOut(t *testing.T, c build.Container) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "out.md")
	if err := build.Write(c, path); err != nil {
		t.Fatalf("Write: %v", err)
	}
	return readFile(t, path)
}

func TestWrite_Title(t *testing.T) {
	got := writeOut(t, build.Container{Title: "Jot"})
	assertContains(t, got, "# Jot")
}

func TestWrite_Description(t *testing.T) {
	got := writeOut(t, build.Container{Title: "Jot", Description: "Note-taking screen."})
	assertContains(t, got, "Note-taking screen.")
}

func TestWrite_Behavior(t *testing.T) {
	got := writeOut(t, build.Container{
		Title:    "List",
		Behavior: []string{"scrollable", "updates in real time"},
	})
	assertContains(t, got, "- scrollable")
	assertContains(t, got, "- updates in real time")
}

func TestWrite_SpecTable(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Card",
		Specs: build.Specs{"color": "blue", "padding": "8dp"},
	})
	assertContains(t, got, "| Spec | Value |")
	assertContains(t, got, "| color | blue |")
	assertContains(t, got, "| padding | 8dp |")
}

func TestWrite_SpecsAreSorted(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Card",
		Specs: build.Specs{"zebra": "z", "alpha": "a", "middle": "m"},
	})
	alphaIdx := strings.Index(got, "alpha")
	middleIdx := strings.Index(got, "middle")
	zebraIdx := strings.Index(got, "zebra")
	if !(alphaIdx < middleIdx && middleIdx < zebraIdx) {
		t.Errorf("specs not sorted alphabetically:\n%s", got)
	}
}

func TestWrite_NoSpecTableWhenEmpty(t *testing.T) {
	got := writeOut(t, build.Container{Title: "Card"})
	assertNotContains(t, got, "| Spec | Value |")
}

func TestWrite_StatesWithMergedSpecs(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Button",
		Specs: build.Specs{"color": "blue", "label": "Submit"},
		States: []build.State{
			{Title: "Loading", Specs: build.Specs{"label": "Loading..."}},
		},
	})
	assertContains(t, got, "## Loading")
	assertContains(t, got, "| color | blue |")    // inherited from base
	assertContains(t, got, "| label | Loading... |") // overridden by state
}

func TestWrite_StateOverrideWins(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Button",
		Specs: build.Specs{"label": "Submit"},
		States: []build.State{
			{Title: "Loading", Specs: build.Specs{"label": "Loading..."}},
		},
	})
	assertContains(t, got, "| label | Loading... |")
	assertNotContains(t, got, "| label | Submit |")
}

func TestWrite_NoSpecTableWithoutStatesOrSpecs(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Screen",
		States: []build.State{
			{Title: "Empty"},
		},
	})
	assertNotContains(t, got, "| Spec | Value |")
}

func TestWrite_Events(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Form",
		Events: []build.Event{
			{
				Title:       "submit",
				Description: "User taps submit.",
				Actions: []build.Action{
					{Title: "save", Description: "Persist data."},
					{Title: "navigate"},
				},
			},
		},
	})
	assertContains(t, got, "## Events")
	assertContains(t, got, "**submit**")
	assertContains(t, got, "User taps submit.")
	assertContains(t, got, "- save: Persist data.")
	assertContains(t, got, "- navigate")
}

func TestWrite_SubContainerHeadingLevel(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Screen",
		Containers: []build.Container{
			{
				Title: "Header",
				Containers: []build.Container{
					{Title: "Logo"},
				},
			},
		},
	})
	assertContains(t, got, "# Screen")
	assertContains(t, got, "## Header")
	assertContains(t, got, "### Logo")
}

func TestWrite_SubContainerSpecs(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Screen",
		Containers: []build.Container{
			{Title: "Card", Specs: build.Specs{"border": "1dp"}},
		},
	})
	assertContains(t, got, "## Card")
	assertContains(t, got, "| border | 1dp |")
}

func TestWrite_SubContainerEvents(t *testing.T) {
	got := writeOut(t, build.Container{
		Title: "Screen",
		Containers: []build.Container{
			{
				Title:  "Button",
				Events: []build.Event{{Title: "on-press"}},
			},
		},
	})
	assertContains(t, got, "### Events")
	assertContains(t, got, "**on-press**")
}

func TestWrite_CreatesNestedDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "c", "out.md")
	if err := build.Write(build.Container{Title: "X"}, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected output file to exist: %v", err)
	}
}
