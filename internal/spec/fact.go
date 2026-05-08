package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FactType string

const (
	FactTypeBoolean FactType = "boolean"
	FactTypeEnum    FactType = "enum"
)

type Fact struct {
	Name   string
	Type   FactType
	Values []string
}

type factFile struct {
	Name   string   `yaml:"name"`
	Type   FactType `yaml:"type"`
	Values []string `yaml:"values"`
}

func validateFactFile(f factFile) (Fact, error) {
	switch f.Type {
	case FactTypeBoolean:
	case FactTypeEnum:
		if len(f.Values) == 0 {
			return Fact{}, fmt.Errorf("fact %q: enum type requires at least one value", f.Name)
		}
	default:
		return Fact{}, fmt.Errorf("fact %q: unknown type %q (must be boolean or enum)", f.Name, f.Type)
	}

	return Fact{Name: f.Name, Type: f.Type, Values: f.Values}, nil
}

// ParseFacts parses a fact file containing either a single fact (mapping) or
// multiple facts (sequence) and returns all facts defined in the file.
func ParseFacts(path string) ([]Fact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading fact file: %w", err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("parsing fact file: %w", err)
	}

	if node.Kind == 0 {
		return nil, nil
	}

	doc := node.Content[0]

	switch doc.Kind {
	case yaml.MappingNode:
		var f factFile
		if err := doc.Decode(&f); err != nil {
			return nil, fmt.Errorf("parsing fact file: %w", err)
		}
		fact, err := validateFactFile(f)
		if err != nil {
			return nil, err
		}
		return []Fact{fact}, nil

	case yaml.SequenceNode:
		var raw []factFile
		if err := doc.Decode(&raw); err != nil {
			return nil, fmt.Errorf("parsing fact file: %w", err)
		}
		facts := make([]Fact, 0, len(raw))
		for _, f := range raw {
			fact, err := validateFactFile(f)
			if err != nil {
				return nil, err
			}
			facts = append(facts, fact)
		}
		return facts, nil

	default:
		return nil, fmt.Errorf("parsing fact file: unexpected YAML structure")
	}
}

// ParseFact parses a single-fact file. Use ParseFacts for files with multiple facts.
func ParseFact(path string) (Fact, error) {
	facts, err := ParseFacts(path)
	if err != nil {
		return Fact{}, err
	}
	if len(facts) != 1 {
		return Fact{}, fmt.Errorf("expected exactly one fact, got %d", len(facts))
	}
	return facts[0], nil
}

func LoadFacts(dir string) ([]Fact, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading facts directory: %w", err)
	}

	var facts []Fact
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".yml" {
			continue
		}
		parsed, err := ParseFacts(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		facts = append(facts, parsed...)
	}
	return facts, nil
}
