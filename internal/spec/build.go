package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// resolveStateSpecs computes per-state specs for a container with inheritance:
// each state = default specs merged with that state's own overrides (state wins).
func resolveStateSpecs(c Container) map[string]map[string]SpecValue {
	if len(c.States) == 0 {
		return nil
	}

	defaultSpecs := make(map[string]SpecValue)
	for _, ss := range c.States {
		if ss.Ref == c.Default {
			defaultSpecs = ss.Specs
			break
		}
	}

	result := make(map[string]map[string]SpecValue, len(c.States))
	for _, ss := range c.States {
		resolved := make(map[string]SpecValue, len(defaultSpecs))
		for k, v := range defaultSpecs {
			resolved[k] = v
		}
		for k, v := range ss.Specs {
			resolved[k] = v
		}
		result[ss.Ref] = resolved
	}
	return result
}

type builtContainerOutput struct {
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description,omitempty"`
	States      map[string]stateOutput `yaml:"states,omitempty"`
	Containers  []sectionOutput        `yaml:"containers,omitempty"`
	Events      []Event                `yaml:"events,omitempty"`
}

type stateOutput struct {
	Specs map[string]SpecValue `yaml:"specs,omitempty"`
}

type sectionOutput struct {
	Title      string                 `yaml:"title"`
	States     map[string]stateOutput `yaml:"states,omitempty"`
	Containers []sectionOutput        `yaml:"containers,omitempty"`
	Events     []Event                `yaml:"events,omitempty"`
}

func statesToOutput(states map[string]map[string]SpecValue) map[string]stateOutput {
	if len(states) == 0 {
		return nil
	}
	out := make(map[string]stateOutput, len(states))
	for name, specs := range states {
		out[name] = stateOutput{Specs: specs}
	}
	return out
}

func buildSectionOutput(c Container) sectionOutput {
	sec := sectionOutput{
		Title:  c.Title,
		States: statesToOutput(resolveStateSpecs(c)),
		Events: c.Events,
	}
	for _, sub := range c.Containers {
		sec.Containers = append(sec.Containers, buildSectionOutput(sub))
	}
	return sec
}

func WriteBuiltContainer(path string, c Container) error {
	out := builtContainerOutput{
		Title:       c.Title,
		Description: c.Description,
		States:      statesToOutput(resolveStateSpecs(c)),
		Events:      c.Events,
	}
	for _, sub := range c.Containers {
		out.Containers = append(out.Containers, buildSectionOutput(sub))
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(out)
}
