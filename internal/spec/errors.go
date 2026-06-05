package spec

import "fmt"

// ValidationError is a structured validation failure.
type ValidationError struct {
	File    string
	Message string
}

func (e ValidationError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s: %s", e.File, e.Message)
	}
	return e.Message
}
