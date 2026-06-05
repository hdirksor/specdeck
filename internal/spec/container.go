package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type SpecValue struct {
	Value       string
	Description string
}

func (s SpecValue) MarshalYAML() (interface{}, error) {
	if s.Description == "" {
		return s.Value, nil
	}
	return struct {
		Value       string `yaml:"value"`
		Description string `yaml:"description"`
	}{s.Value, s.Description}, nil
}

type StateSpec struct {
	Ref    string
	Specs  map[string]SpecValue
	Events map[string]interface{}
}

// Action holds the payload fields for a single action type.
type Action map[string]interface{}

// Event is a top-level interaction on a container: a trigger name, optional
// description, and one or more typed actions.
type Event struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description,omitempty"`
	Actions     map[string]Action `yaml:"actions,omitempty"`
}

type Container struct {
	// Path is relative to the containers/ root directory.
	Path        string
	Title       string
	Description string
	Default     string
	States      []StateSpec
	Containers  []Container
	Events      []Event

	// rawRef and rawOverrides hold unresolved $ref data during parsing.
	// LoadContainerTree resolves them; after that these are zero.
	rawRef       string
	rawOverrides map[string]SpecValue
}

func (c Container) MarshalYAML() (interface{}, error) {
	type stateOut struct {
		Specs map[string]SpecValue `yaml:"specs,omitempty"`
	}
	type marshalForm struct {
		Title       string               `yaml:"title"`
		Description string               `yaml:"description,omitempty"`
		States      map[string]stateOut  `yaml:"states,omitempty"`
		Containers  []Container          `yaml:"containers,omitempty"`
		Events      []Event              `yaml:"events,omitempty"`
	}

	var states map[string]stateOut
	if resolved := resolveStateSpecs(c); len(resolved) > 0 {
		states = make(map[string]stateOut, len(resolved))
		for name, specs := range resolved {
			states[name] = stateOut{Specs: specs}
		}
	}

	return marshalForm{
		Title:       c.Title,
		Description: c.Description,
		States:      states,
		Containers:  c.Containers,
		Events:      c.Events,
	}, nil
}

// containerFile is the raw YAML structure for a container file.
type containerFile struct {
	Title       string  `yaml:"title"`
	Description string  `yaml:"description"`
	Default     string  `yaml:"default"`
	Events      []Event `yaml:"events"`
}

// ParseContainer parses a single container file without following $refs.
// $ref entries in containers: become stubs in Container.Containers with
// unexported rawRef set. Use LoadContainerTree to get a fully resolved tree.
func ParseContainer(path string) (Container, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Container{}, fmt.Errorf("reading container file: %w", err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return Container{}, fmt.Errorf("parsing container file: %w", err)
	}

	if root.Kind != yaml.DocumentNode || len(root.Content) == 0 {
		return Container{}, nil
	}

	c, err := parseContainerNode(root.Content[0])
	if err != nil {
		return Container{}, err
	}

	if c.Title == "" {
		c.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return c, nil
}

func parseContainerNode(doc *yaml.Node) (Container, error) {
	var raw containerFile
	if err := doc.Decode(&raw); err != nil {
		return Container{}, fmt.Errorf("parsing container: %w", err)
	}

	defaultRef := "default"
	if raw.Default != "" {
		defaultRef = raw.Default
	}

	c := Container{
		Default:     defaultRef,
		Title:       raw.Title,
		Description: raw.Description,
		Events:      raw.Events,
	}

	if specsNode := findNode(doc, "specs"); specsNode != nil {
		specs, err := parseSpecNode(specsNode)
		if err != nil {
			return Container{}, err
		}
		if len(specs) > 0 {
			c.States = append(c.States, StateSpec{Ref: defaultRef, Specs: specs})
		}
	}

	statesSeq := findSequenceNode(doc, "states")
	for _, item := range statesSeq {
		ss, err := parseStateSpecNode(item)
		if err != nil {
			return Container{}, err
		}
		if ss.Ref != "" {
			c.States = append(c.States, ss)
		}
	}

	containers, err := parseContainersNode(doc)
	if err != nil {
		return Container{}, err
	}
	c.Containers = containers

	return c, nil
}

// parseContainersNode parses the containers: sequence from a YAML mapping node.
// $ref entries produce stubs with rawRef/rawOverrides set.
// Inline entries are fully parsed into Containers.
func parseContainersNode(doc *yaml.Node) ([]Container, error) {
	seq := findSequenceNode(doc, "containers")
	if seq == nil {
		return nil, nil
	}

	var containers []Container
	for _, item := range seq {
		if item.Kind != yaml.MappingNode {
			continue
		}

		var ref string
		var overrideNodes map[string]yaml.Node

		for j := 0; j+1 < len(item.Content); j += 2 {
			k := item.Content[j].Value
			v := item.Content[j+1]
			if k == "$ref" {
				ref = v.Value
			} else {
				if overrideNodes == nil {
					overrideNodes = make(map[string]yaml.Node)
				}
				overrideNodes[k] = *v
			}
		}

		if ref != "" {
			var overrides map[string]SpecValue
			if len(overrideNodes) > 0 {
				var err error
				overrides, err = parseSpecMap(overrideNodes)
				if err != nil {
					return nil, fmt.Errorf("parsing overrides for %s: %w", ref, err)
				}
			}
			containers = append(containers, Container{rawRef: ref, rawOverrides: overrides})
		} else {
			inline, err := parseContainerNode(item)
			if err != nil {
				return nil, fmt.Errorf("parsing inline container: %w", err)
			}
			containers = append(containers, inline)
		}
	}
	return containers, nil
}

// LoadContainerTree loads a container file and recursively resolves all $ref entries.
func LoadContainerTree(path, containersRoot string) (Container, error) {
	c, err := ParseContainer(path)
	if err != nil {
		return Container{}, err
	}
	return resolveContainerRefs(c, filepath.Dir(path), containersRoot)
}

func resolveContainerRefs(c Container, baseDir, containersRoot string) (Container, error) {
	resolved := make([]Container, 0, len(c.Containers))
	for _, sub := range c.Containers {
		if sub.rawRef == "" {
			r, err := resolveContainerRefs(sub, baseDir, containersRoot)
			if err != nil {
				return Container{}, err
			}
			resolved = append(resolved, r)
		} else {
			absPath := filepath.Join(baseDir, sub.rawRef)
			loaded, err := ParseContainer(absPath)
			if err != nil {
				return Container{}, fmt.Errorf("loading %s: %w", sub.rawRef, err)
			}
			rel, err := filepath.Rel(containersRoot, absPath)
			if err != nil {
				return Container{}, fmt.Errorf("resolving path for %s: %w", sub.rawRef, err)
			}
			loaded.Path = rel
			if len(sub.rawOverrides) > 0 {
				loaded = applyOverrides(loaded, sub.rawOverrides)
			}
			r, err := resolveContainerRefs(loaded, filepath.Dir(absPath), containersRoot)
			if err != nil {
				return Container{}, err
			}
			resolved = append(resolved, r)
		}
	}
	c.Containers = resolved
	return c, nil
}

func applyOverrides(c Container, overrides map[string]SpecValue) Container {
	for i, ss := range c.States {
		merged := make(map[string]SpecValue, len(ss.Specs)+len(overrides))
		for k, v := range ss.Specs {
			merged[k] = v
		}
		for k, v := range overrides {
			merged[k] = v
		}
		c.States[i].Specs = merged
	}
	return c
}

// LoadContainers walks the containers root directory and loads every container file
// with all $refs resolved.
func LoadContainers(root string) ([]Container, error) {
	var containers []Container

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".yml" {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		c, err := LoadContainerTree(path, root)
		if err != nil {
			return fmt.Errorf("loading container %s: %w", rel, err)
		}

		c.Path = rel
		containers = append(containers, c)
		return nil
	})

	return containers, err
}

func findNode(doc *yaml.Node, key string) *yaml.Node {
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key {
			return doc.Content[i+1]
		}
	}
	return nil
}

func parseSpecNode(node *yaml.Node) (map[string]SpecValue, error) {
	if node == nil {
		return nil, nil
	}
	switch node.Kind {
	case yaml.MappingNode:
		specs := make(map[string]SpecValue, len(node.Content)/2)
		for i := 0; i+1 < len(node.Content); i += 2 {
			sv, err := parseSpecValue(*node.Content[i+1])
			if err != nil {
				return nil, fmt.Errorf("spec %q: %w", node.Content[i].Value, err)
			}
			specs[node.Content[i].Value] = sv
		}
		return specs, nil
	case yaml.SequenceNode:
		specs := make(map[string]SpecValue, len(node.Content))
		for _, item := range node.Content {
			if item.Kind != yaml.MappingNode || len(item.Content) < 2 {
				continue
			}
			key := item.Content[0].Value
			sv, err := parseSpecValue(*item.Content[1])
			if err != nil {
				return nil, fmt.Errorf("spec %q: %w", key, err)
			}
			specs[key] = sv
		}
		return specs, nil
	default:
		return nil, fmt.Errorf("unexpected YAML node kind %v for specs", node.Kind)
	}
}

func findSequenceNode(doc *yaml.Node, key string) []*yaml.Node {
	if doc.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(doc.Content); i += 2 {
		if doc.Content[i].Value == key && doc.Content[i+1].Kind == yaml.SequenceNode {
			return doc.Content[i+1].Content
		}
	}
	return nil
}

func parseStateSpecNode(node *yaml.Node) (StateSpec, error) {
	if node.Kind != yaml.MappingNode {
		return StateSpec{}, fmt.Errorf("expected state entry to be a mapping")
	}

	var ref string
	var specs map[string]SpecValue
	var events map[string]interface{}

	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "ref" {
			ref = node.Content[i+1].Value
			break
		}
	}

	if ref == "" && len(node.Content) >= 2 {
		known := map[string]bool{"specs": true, "events": true, "behavior": true}
		for i := 0; i+1 < len(node.Content); i += 2 {
			if !known[node.Content[i].Value] {
				ref = node.Content[i].Value
				break
			}
		}
	}

	for i := 0; i+1 < len(node.Content); i += 2 {
		switch node.Content[i].Value {
		case "specs":
			var err error
			specs, err = parseSpecNode(node.Content[i+1])
			if err != nil {
				return StateSpec{}, fmt.Errorf("parsing specs: %w", err)
			}
		case "events":
			if err := node.Content[i+1].Decode(&events); err != nil {
				return StateSpec{}, fmt.Errorf("parsing events: %w", err)
			}
		}
	}

	return StateSpec{Ref: ref, Specs: specs, Events: events}, nil
}

func parseSpecMap(raw map[string]yaml.Node) (map[string]SpecValue, error) {
	specs := make(map[string]SpecValue, len(raw))
	for key, node := range raw {
		sv, err := parseSpecValue(node)
		if err != nil {
			return nil, fmt.Errorf("spec %q: %w", key, err)
		}
		specs[key] = sv
	}
	return specs, nil
}

func parseSpecValue(node yaml.Node) (SpecValue, error) {
	n := &node
	if n.Kind == yaml.AliasNode {
		n = n.Alias
	}

	switch n.Kind {
	case yaml.ScalarNode:
		return SpecValue{Value: n.Value}, nil
	case yaml.MappingNode:
		var verbose struct {
			Value       string `yaml:"value"`
			Description string `yaml:"description"`
		}
		if err := n.Decode(&verbose); err != nil {
			return SpecValue{}, fmt.Errorf("decoding verbose spec: %w", err)
		}
		return SpecValue{Value: verbose.Value, Description: verbose.Description}, nil
	default:
		return SpecValue{}, fmt.Errorf("unexpected YAML node kind %v", n.Kind)
	}
}
