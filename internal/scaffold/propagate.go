package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/spec"
)

// PropagateState adds the named state to all leaf containers that don't already have it,
// copying the default specs of each container as a starting point.
func PropagateState(projectRoot, stateName string) error {
	if err := requireStateExists(projectRoot, stateName); err != nil {
		return err
	}

	containersRoot := filepath.Join(projectRoot, "containers")
	return filepath.WalkDir(containersRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".yml" || d.Name() == "index.yml" {
			return nil
		}

		return addStateToContainer(path, stateName)
	})
}

func requireStateExists(projectRoot, stateName string) error {
	states, err := spec.LoadStates(filepath.Join(projectRoot, "states"))
	if err != nil {
		return err
	}
	for _, s := range states {
		if s.Name == stateName {
			return nil
		}
	}
	return fmt.Errorf("state %q is not defined in states/", stateName)
}

func addStateToContainer(path, stateName string) error {
	c, err := spec.ParseContainer(path)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	for _, s := range c.States {
		if s.Ref == stateName {
			return nil // already present, skip
		}
	}

	defaultRef := c.Default
	if defaultRef == "" {
		defaultRef = "default"
	}

	var defaultSpecs map[string]spec.SpecValue
	var defaultEvents map[string]interface{}
	for _, s := range c.States {
		if s.Ref == defaultRef {
			defaultSpecs = s.Specs
			defaultEvents = s.Events
			break
		}
	}

	copiedSpecs := make(map[string]spec.SpecValue, len(defaultSpecs))
	for k, v := range defaultSpecs {
		copiedSpecs[k] = v
	}
	copiedEvents := make(map[string]interface{}, len(defaultEvents))
	for k, v := range defaultEvents {
		copiedEvents[k] = v
	}

	c.States = append(c.States, spec.StateSpec{Ref: stateName, Specs: copiedSpecs, Events: copiedEvents})
	return spec.WriteContainer(path, c)
}
