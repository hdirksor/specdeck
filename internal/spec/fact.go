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

type FactScope string

const (
	FactScopeCross    FactScope = "cross"
	FactScopeIsolated FactScope = "isolated"
	FactScopeManual   FactScope = "manual"
)

type Fact struct {
	Name   string
	Type   FactType
	Values []string
	Scope  FactScope
}

type factFile struct {
	Name   string    `yaml:"name"`
	Type   FactType  `yaml:"type"`
	Values []string  `yaml:"values"`
	Scope  FactScope `yaml:"scope"`
}

func ParseFact(path string) (Fact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Fact{}, fmt.Errorf("reading fact file: %w", err)
	}

	var f factFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Fact{}, fmt.Errorf("parsing fact file: %w", err)
	}

	switch f.Type {
	case FactTypeBoolean:
	case FactTypeEnum:
		if len(f.Values) == 0 {
			return Fact{}, fmt.Errorf("fact %q: enum type requires at least one value", f.Name)
		}
	default:
		return Fact{}, fmt.Errorf("fact %q: unknown type %q (must be boolean or enum)", f.Name, f.Type)
	}

	switch f.Scope {
	case FactScopeCross, FactScopeIsolated, FactScopeManual:
	case "":
		f.Scope = FactScopeManual
	default:
		return Fact{}, fmt.Errorf("fact %q: unknown scope %q (must be cross, isolated, or manual)", f.Name, f.Scope)
	}

	return Fact{Name: f.Name, Type: f.Type, Values: f.Values, Scope: f.Scope}, nil
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
		fact, err := ParseFact(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		facts = append(facts, fact)
	}
	return facts, nil
}
