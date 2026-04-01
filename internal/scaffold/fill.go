package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hdickson/specdeck/internal/spec"
)

// Fill copies the default state's specs into any states with empty specs in the
// container at containerPath. States that already have specs are left untouched.
func Fill(projectRoot, containerPath string) error {
	filePath := resolveContainerPath(projectRoot, containerPath)

	c, err := spec.ParseContainer(filePath)
	if err != nil {
		return fmt.Errorf("parsing container: %w", err)
	}

	defaultRef := c.Default
	if defaultRef == "" {
		defaultRef = "default"
	}

	var defaultSpecs map[string]spec.SpecValue
	for _, s := range c.States {
		if s.Ref == defaultRef {
			defaultSpecs = s.Specs
			break
		}
	}
	if defaultSpecs == nil {
		return fmt.Errorf("container %q has no %q state to fill from", containerPath, defaultRef)
	}

	var defaultEvents map[string]interface{}
	for _, s := range c.States {
		if s.Ref == defaultRef {
			defaultEvents = s.Events
			break
		}
	}

	for i, s := range c.States {
		if s.Ref == defaultRef {
			continue
		}
		if len(s.Specs) == 0 {
			filled := make(map[string]spec.SpecValue, len(defaultSpecs))
			for k, v := range defaultSpecs {
				filled[k] = v
			}
			c.States[i].Specs = filled
		}
		if len(s.Events) == 0 && len(defaultEvents) > 0 {
			filledEvents := make(map[string]interface{}, len(defaultEvents))
			for k, v := range defaultEvents {
				filledEvents[k] = v
			}
			c.States[i].Events = filledEvents
		}
	}

	return spec.WriteContainer(filePath, c)
}

// FillAll runs Fill across every leaf container in the project.
// Returns the number of containers filled and the number skipped (no default state).
func FillAll(projectRoot string) (filled, skipped int, err error) {
	containersRoot := filepath.Join(projectRoot, "containers")
	err = filepath.WalkDir(containersRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(d.Name()) != ".yml" || d.Name() == "index.yml" {
			return nil
		}

		rel, _ := filepath.Rel(filepath.Join(projectRoot, "containers"), path)
		// Strip .yml extension for the logical path.
		logicalPath := rel[:len(rel)-len(".yml")]

		fillErr := Fill(projectRoot, filepath.ToSlash(logicalPath))
		if fillErr != nil {
			// Skip containers that have no default state rather than aborting.
			skipped++
			return nil
		}
		filled++
		return nil
	})
	return
}

// resolveContainerPath returns the filesystem path for a container given its logical path.
// It checks for a leaf file first, then falls back to index.yml for non-leaf containers.
func resolveContainerPath(projectRoot, containerPath string) string {
	leaf := filepath.Join(projectRoot, "containers", filepath.FromSlash(containerPath)+".yml")
	if fileExists(leaf) {
		return leaf
	}
	return filepath.Join(projectRoot, "containers", filepath.FromSlash(containerPath), "index.yml")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
