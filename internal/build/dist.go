package build

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WriteDist renders each container into distRoot using Hugo's layout:
// index.yml becomes _index.md and every other file becomes <stem>.md.
// Directories holding leaf pages but no index get a generated _index.md stub.
// It returns the written paths relative to distRoot, slash-separated.
func WriteDist(containers []Container, distRoot string) ([]string, error) {
	var written []string
	sectionDirs := make(map[string]bool)
	leafDirs := make(map[string]bool)

	for _, c := range containers {
		stem := strings.TrimSuffix(c.Path, filepath.Ext(c.Path))
		relDir := filepath.Dir(stem)

		var relOut string
		if filepath.Base(stem) == "index" {
			relOut = filepath.Join(relDir, "_index.md")
			sectionDirs[relDir] = true
		} else {
			relOut = stem + ".md"
			if relDir != "." {
				leafDirs[relDir] = true
			}
		}

		if err := Write(c, filepath.Join(distRoot, relOut)); err != nil {
			return written, fmt.Errorf("%s: %w", c.Path, err)
		}
		written = append(written, filepath.ToSlash(relOut))
	}

	var stubDirs []string
	for dir := range leafDirs {
		if !sectionDirs[dir] {
			stubDirs = append(stubDirs, dir)
		}
	}
	sort.Strings(stubDirs)

	for _, dir := range stubDirs {
		relOut := filepath.Join(dir, "_index.md")
		if err := writeSectionStub(filepath.Join(distRoot, relOut), dirTitle(filepath.Base(dir))); err != nil {
			return written, fmt.Errorf("writing section stub for %s: %w", dir, err)
		}
		written = append(written, filepath.ToSlash(relOut))
	}

	return written, nil
}

func writeSectionStub(path, title string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	return os.WriteFile(path, fmt.Appendf(nil, "---\ntitle: %s\n---\n", title), 0644)
}

func dirTitle(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '-' || r == '_' })
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
