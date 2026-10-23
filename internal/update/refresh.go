package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LatestReleaseURL is GitHub's latest stable release endpoint for specdeck.
const LatestReleaseURL = "https://api.github.com/repos/hdirksor/specdeck/releases/latest"

// Timeout bounds the background release request.
const Timeout = 2 * time.Second

// RefreshDue reports whether the last release check is more than Interval old.
func RefreshDue(state State, now time.Time) bool {
	return now.Sub(state.CheckedAt) >= Interval
}

// Refresh fetches the latest release from url and stores it in the cache at
// path. On failure the cache is left as it was.
func Refresh(ctx context.Context, client *http.Client, url, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building release request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetching latest release: unexpected status %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("decoding latest release: %w", err)
	}
	if release.TagName == "" {
		return fmt.Errorf("decoding latest release: missing tag_name")
	}

	state := LoadState(path)
	state.LatestVersion = release.TagName
	state.LatestURL = release.HTMLURL
	return SaveState(path, state)
}
