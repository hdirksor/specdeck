package spec

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func WriteHugoSectionStub(path, title string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	fmt.Fprintf(f, "---\ntitle: %s\n---\n", title)
	return nil
}

func WriteBuiltContainerMarkdown(path string, c Container) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	return renderMarkdown(f, c)
}

func renderMarkdown(w io.Writer, c Container) error {
	fmt.Fprintf(w, "# %s\n", c.Title)
	if c.Description != "" {
		fmt.Fprintf(w, "\n%s\n", c.Description)
	}

	if states := resolveStateSpecs(c); len(states) > 0 {
		writeStateSections(w, states, "##")
	}

	if len(c.Events) > 0 {
		fmt.Fprintf(w, "\n## Events\n\n")
		writeEventList(w, c.Events, "###")
	}

	for _, sub := range c.Containers {
		renderSubContainer(w, sub, 2)
	}

	return nil
}

func renderSubContainer(w io.Writer, c Container, level int) {
	h := strings.Repeat("#", level)
	fmt.Fprintf(w, "\n%s %s\n", h, c.Title)

	if states := resolveStateSpecs(c); len(states) > 0 {
		writeStateSections(w, states, strings.Repeat("#", level+1))
	}

	if len(c.Events) > 0 {
		fmt.Fprintf(w, "\n%s Events\n\n", strings.Repeat("#", level+1))
		writeEventList(w, c.Events, strings.Repeat("#", level+2))
	}

	for _, sub := range c.Containers {
		renderSubContainer(w, sub, level+1)
	}
}

func writeStateSections(w io.Writer, states map[string]map[string]SpecValue, heading string) {
	names := sortedStateNames(states)
	for _, name := range names {
		fmt.Fprintf(w, "\n%s %s\n\n", heading, name)
		specs := states[name]
		keys := sortedKeys(specs)
		fmt.Fprintln(w, "| Spec | Value |")
		fmt.Fprintln(w, "| --- | --- |")
		for _, key := range keys {
			fmt.Fprintf(w, "| %s | %s |\n", key, specs[key].Value)
		}
	}
}

func sortedStateNames(states map[string]map[string]SpecValue) []string {
	names := make([]string, 0, len(states))
	for name := range states {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if names[i] == "default" {
			return true
		}
		if names[j] == "default" {
			return false
		}
		return names[i] < names[j]
	})
	return names
}

func sortedKeys(specs map[string]SpecValue) []string {
	keys := make([]string, 0, len(specs))
	for k := range specs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func writeEventList(w io.Writer, events []Event, heading string) {
	for _, ev := range events {
		fmt.Fprintf(w, "%s %s\n", heading, ev.Title)
		if ev.Description != "" {
			fmt.Fprintf(w, "\n%s\n", ev.Description)
		}
		if len(ev.Actions) > 0 {
			fmt.Fprintf(w, "\n**Actions**\n\n")
			actionNames := make([]string, 0, len(ev.Actions))
			for k := range ev.Actions {
				actionNames = append(actionNames, k)
			}
			sort.Strings(actionNames)
			for _, name := range actionNames {
				action := ev.Actions[name]
				fmt.Fprintf(w, "- **%s**", name)
				if len(action) > 0 {
					fields := make([]string, 0, len(action))
					for k := range action {
						fields = append(fields, k)
					}
					sort.Strings(fields)
					parts := make([]string, 0, len(fields))
					for _, k := range fields {
						parts = append(parts, fmt.Sprintf("%s: %v", k, action[k]))
					}
					fmt.Fprintf(w, ": %s", strings.Join(parts, ", "))
				}
				fmt.Fprintln(w)
			}
		}
		fmt.Fprintln(w)
	}
}
