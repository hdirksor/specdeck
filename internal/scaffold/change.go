package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// NewChange creates a dated change record in the changes/ directory and returns
// its relative path. The changes/ directory must already exist (created by New).
func NewChange(dir, title string) (string, error) {
	changesDir := filepath.Join(dir, "changes")
	if _, err := os.Stat(changesDir); os.IsNotExist(err) {
		return "", fmt.Errorf("changes/ directory not found — is this a specdeck project?")
	}

	slug := nonAlphanumeric.ReplaceAllString(strings.ToLower(title), "-")
	slug = strings.Trim(slug, "-")
	filename := time.Now().Format("2006-01-02") + "-" + slug + ".md"
	relPath := filepath.Join("changes", filename)

	content := "# " + title + "\n\n" +
		"## What changed\n\n" +
		"<!-- Describe what spec changes this PR introduces. -->\n\n" +
		"## Why\n\n" +
		"<!-- Motivation: user research, product decision, bug fix, etc. -->\n\n" +
		"## References\n\n" +
		"<!-- Ticket URL, Slack thread, stakeholder name, etc. -->\n"

	if err := os.WriteFile(filepath.Join(dir, relPath), []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing change record: %w", err)
	}

	return relPath, nil
}
