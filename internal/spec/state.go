package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type State struct {
	Name    string
	Summary string
	Facts   map[string]interface{}
}

type stateFile struct {
	Name    string                 `yaml:"name"`
	Summary string                 `yaml:"summary"`
	Facts   map[string]interface{} `yaml:"facts"`
}

func ParseStates(path string) ([]State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading states file: %w", err)
	}

	var raw []stateFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing states file: %w", err)
	}

	seen := make(map[string]bool)
	states := make([]State, 0, len(raw))
	for _, r := range raw {
		if seen[r.Name] {
			return nil, fmt.Errorf("duplicate state name %q in %s", r.Name, path)
		}
		seen[r.Name] = true
		states = append(states, State{Name: r.Name, Summary: r.Summary, Facts: r.Facts})
	}
	return states, nil
}

func LoadStates(dir string) ([]State, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading states directory: %w", err)
	}

	// Load manual files first; states.lock is loaded last and only fills gaps.
	seen := make(map[string]bool)
	var states []State

	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yml" || e.Name() == "states.lock" {
			continue
		}
		parsed, err := ParseStates(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		for _, s := range parsed {
			if seen[s.Name] {
				return nil, fmt.Errorf("duplicate state name %q across state files", s.Name)
			}
			seen[s.Name] = true
			states = append(states, s)
		}
	}

	generatedPath := filepath.Join(dir, "states.lock")
	if _, err := os.Stat(generatedPath); err == nil {
		generated, err := ParseStates(generatedPath)
		if err != nil {
			return nil, err
		}
		for _, s := range generated {
			if !seen[s.Name] {
				seen[s.Name] = true
				states = append(states, s)
			}
		}
	}

	return states, nil
}
