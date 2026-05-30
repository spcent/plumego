package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultCheckInterval is the default interval between update checks (in minutes).
	DefaultCheckInterval = 60 * 24 // 24 hours
	// DefaultUpdateURL is the default URL for checking updates.
	DefaultUpdateURL = "https://releases.example.com/cloud-vault/latest.json"
)

// Checker handles checking for application updates.
type Checker struct {
	currentVersion VersionInfo
	updateURL      string
	httpClient     *http.Client
}

// NewChecker creates a new update checker.
func NewChecker(currentVersion VersionInfo, updateURL string) *Checker {
	if updateURL == "" {
		updateURL = DefaultUpdateURL
	}

	return &Checker{
		currentVersion: currentVersion,
		updateURL:      updateURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckForUpdate checks if a newer version is available.
func (c *Checker) CheckForUpdate(ctx context.Context) (*LatestRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.updateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch update info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("update server returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	var release LatestRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("parse update info: %w", err)
	}

	return &release, nil
}

// CompareVersions compares two semantic version strings.
// Returns true if latest is newer than current.
func CompareVersions(current, latest string) bool {
	// Simple version comparison - can be enhanced with proper semver parsing
	// For now, just do string comparison which works for most cases
	return latest > current
}
