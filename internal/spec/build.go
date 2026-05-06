package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Section is one named group of specs within a built container output.
type Section struct {
	Title  string
	Specs  map[string]SpecValue
	Events []Event
}

// ResolveContainerSections returns one Section per logical unit:
// the container's own specs (if any) followed by one section per direct import.
// Each section title comes from the referenced container's title field,
// falling back to its filename stem.
func ResolveContainerSections(c Container, containersRoot string, byPath map[string]Container) []Section {
	var sections []Section

	ownSpecs := make(map[string]SpecValue)
	for _, ss := range c.States {
		for k, v := range ss.Specs {
			if _, exists := ownSpecs[k]; !exists {
				ownSpecs[k] = v
			}
		}
	}
	if len(ownSpecs) > 0 {
		sections = append(sections, Section{Title: c.Title, Specs: ownSpecs})
	}

	for _, imp := range c.Imports {
		absDir := filepath.Dir(filepath.Join(containersRoot, c.Path))
		absImport := filepath.Join(absDir, imp.Ref)
		rel, err := filepath.Rel(containersRoot, absImport)
		if err != nil {
			continue
		}
		imported, ok := byPath[rel]
		if !ok {
			continue
		}
		resolved := resolveContainer(imported, containersRoot, byPath, make(map[string]bool))
		for k, v := range imp.Overrides {
			resolved[k] = v
		}
		if len(resolved) > 0 || len(imported.Events) > 0 {
			sections = append(sections, Section{Title: imported.Title, Specs: resolved, Events: imported.Events})
		}
	}

	return sections
}

// resolveContainer merges specs from a container and all its imports recursively.
func resolveContainer(c Container, containersRoot string, byPath map[string]Container, visited map[string]bool) map[string]SpecValue {
	if visited[c.Path] {
		return nil
	}
	visited[c.Path] = true

	specs := make(map[string]SpecValue)

	for _, ss := range c.States {
		for k, v := range ss.Specs {
			if _, exists := specs[k]; !exists {
				specs[k] = v
			}
		}
	}

	for _, imp := range c.Imports {
		absDir := filepath.Dir(filepath.Join(containersRoot, c.Path))
		absImport := filepath.Join(absDir, imp.Ref)
		rel, err := filepath.Rel(containersRoot, absImport)
		if err != nil {
			continue
		}
		imported, ok := byPath[rel]
		if !ok {
			continue
		}
		for k, v := range resolveContainer(imported, containersRoot, byPath, visited) {
			if _, exists := specs[k]; !exists {
				specs[k] = v
			}
		}
	}

	return specs
}

type builtContainerOutput struct {
	Title       string          `yaml:"title"`
	Description string          `yaml:"description,omitempty"`
	Sections    []sectionOutput `yaml:"containers,omitempty"`
	Events      []eventOutput   `yaml:"events,omitempty"`
}

type eventOutput struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description,omitempty"`
	Actions     map[string]Action `yaml:"actions,omitempty"`
}

type sectionOutput struct {
	Title  string                     `yaml:"title"`
	Specs  map[string]specValueOutput `yaml:"specs,omitempty"`
	Events []eventOutput              `yaml:"events,omitempty"`
}

func WriteBuiltContainer(path string, c Container, sections []Section) error {
	out := builtContainerOutput{
		Title:       c.Title,
		Description: c.Description,
	}
	for _, s := range sections {
		specs := make(map[string]specValueOutput, len(s.Specs))
		for k, v := range s.Specs {
			specs[k] = specValueOutput{Value: v.Value, Description: v.Description}
		}
		sec := sectionOutput{Title: s.Title, Specs: specs}
		for _, ev := range s.Events {
			sec.Events = append(sec.Events, eventOutput{
				Title:       ev.Title,
				Description: ev.Description,
				Actions:     ev.Actions,
			})
		}
		out.Sections = append(out.Sections, sec)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	for _, ev := range c.Events {
		out.Events = append(out.Events, eventOutput{
			Title:       ev.Title,
			Description: ev.Description,
			Actions:     ev.Actions,
		})
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	return enc.Encode(out)
}
