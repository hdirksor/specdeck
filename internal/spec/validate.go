package spec

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateImports checks that every $ref in each container resolves to an existing file.
// Paths are resolved relative to the container file's directory within containersRoot.
func ValidateImports(containersRoot string, containers []Container) []ValidationError {
	var errs []ValidationError
	for _, c := range containers {
		containerDir := filepath.Dir(filepath.Join(containersRoot, c.Path))
		for _, imp := range c.Imports {
			resolved := filepath.Join(containerDir, imp.Ref)
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				errs = append(errs, ValidationError{
					File:    c.Path,
					Line:    imp.Line,
					Message: fmt.Sprintf("$ref %q not found", imp.Ref),
				})
			}
		}
	}
	return errs
}
