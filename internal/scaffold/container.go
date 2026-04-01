package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hdickson/specdeck/internal/spec"
)

// IntermediateDirectories returns the relative paths (from projectRoot) of directories
// that do not yet exist and would be created by AddContainer for the given containerPath.
func IntermediateDirectories(projectRoot, containerPath string) []string {
	parts := strings.Split(containerPath, "/")
	// The last part is the filename; intermediate dirs are everything before it.
	dirParts := parts[:len(parts)-1]

	var missing []string
	current := filepath.Join("containers")
	for _, part := range dirParts {
		current = filepath.Join(current, part)
		if _, err := os.Stat(filepath.Join(projectRoot, current)); os.IsNotExist(err) {
			missing = append(missing, current)
		}
	}
	return missing
}

// AddContainer creates a new leaf container file at containerPath within the project,
// creating intermediate directories as needed. All existing states are stubbed in with
// empty specs.
func AddContainer(projectRoot, containerPath string) error {
	parts := strings.Split(containerPath, "/")
	filename := parts[len(parts)-1] + ".yml"
	dirParts := parts[:len(parts)-1]

	containerDir := filepath.Join(append([]string{projectRoot, "containers"}, dirParts...)...)
	filePath := filepath.Join(containerDir, filename)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("container already exists: %s", filePath)
	}

	if err := os.MkdirAll(containerDir, 0755); err != nil {
		return fmt.Errorf("creating container directories: %w", err)
	}

	states, err := spec.LoadStates(filepath.Join(projectRoot, "states"))
	if err != nil {
		return fmt.Errorf("loading states: %w", err)
	}

	stateSpecs := make([]spec.StateSpec, 0, len(states))
	for _, s := range states {
		stateSpecs = append(stateSpecs, spec.StateSpec{
			Ref:    s.Name,
			Specs:  map[string]spec.SpecValue{},
			Events: map[string]interface{}{},
		})
	}

	return spec.WriteContainer(filePath, spec.Container{States: stateSpecs})
}
