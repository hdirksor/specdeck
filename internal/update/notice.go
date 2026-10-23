package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Interval is how often the release check runs and the notice is shown.
const Interval = 24 * time.Hour

// State is the cached result of the last release check.
type State struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	LatestURL     string    `json:"latest_url"`
	NotifiedAt    time.Time `json:"notified_at"`
}

// LoadState reads the cache at path. A missing or unreadable cache is treated
// as empty, since the update check must never get in the user's way.
func LoadState(path string) State {
	data, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}
	}
	return state
}

// SaveState writes the cache to path, creating its directory.
func SaveState(path string, state State) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("encoding update-check state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating update-check cache directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing update-check cache: %w", err)
	}
	return nil
}

// Notice returns the message to show for the cached state, or "" if there is
// nothing to say. When a message is returned, the state records that it was
// shown so it appears at most once per Interval.
func Notice(state State, current Build, now time.Time) (string, State) {
	if !Newer(state.LatestVersion, current.Version) {
		return "", state
	}
	if now.Sub(state.NotifiedAt) < Interval {
		return "", state
	}

	msg := fmt.Sprintf("specdeck %s is available (you have %s) — %s\n", state.LatestVersion, current.Version, state.LatestURL)
	if current.Method == MethodGoInstall {
		msg += "Upgrade: go install github.com/hdirksor/specdeck@latest\n"
	}
	state.NotifiedAt = now
	return msg, state
}
