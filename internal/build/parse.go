package build

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseFile unmarshals a single YAML file into a Container without resolving $refs.
func ParseFile(path string) (Container, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Container{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var c Container
	if err := yaml.Unmarshal(data, &c); err != nil {
		return Container{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return c, nil
}

// Load walks root, parses every .yml file, and resolves all $ref entries.
func Load(root string) ([]Container, error) {
	var containers []Container
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(d.Name()) != ".yml" {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		c, err := loadFile(path, root)
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		c.Path = rel
		containers = append(containers, c)
		return nil
	})
	return containers, err
}

func loadFile(path, root string) (Container, error) {
	c, err := ParseFile(path)
	if err != nil {
		return Container{}, err
	}
	return resolveRefs(c, filepath.Dir(path), root)
}

func resolveRefs(c Container, baseDir, root string) (Container, error) {
	resolved := make([]Container, 0, len(c.Containers))
	for _, sub := range c.Containers {
		r, err := resolveEntry(sub, baseDir, root)
		if err != nil {
			return Container{}, err
		}
		resolved = append(resolved, r)
	}
	c.Containers = resolved
	c.Ref = ""
	c.Overrides = nil
	return c, nil
}

func resolveEntry(sub Container, baseDir, root string) (Container, error) {
	if sub.Ref == "" {
		return resolveRefs(sub, baseDir, root)
	}

	absRef := filepath.Join(baseDir, sub.Ref)
	r, err := loadFile(absRef, root)
	if err != nil {
		return Container{}, fmt.Errorf("loading %s: %w", sub.Ref, err)
	}
	rel, _ := filepath.Rel(root, absRef)
	r.Path = rel

	for k, v := range sub.Overrides {
		if r.Specs == nil {
			r.Specs = make(Specs)
		}
		r.Specs[k] = v
	}
	return r, nil
}

