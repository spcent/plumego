// Package update provides auto-update checking functionality.
package update

import "time"

// VersionInfo represents the current application version information.
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	Channel   string `json:"channel"`
}

// LatestRelease represents the latest release information from the update server.
type LatestRelease struct {
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	BuildTime   string `json:"build_time"`
	Channel     string `json:"channel"`
	ReleaseDate string `json:"release_date"`
	ReleaseURL  string `json:"release_url"`
	DownloadURL string `json:"download_url"`
	Changelog   string `json:"changelog"`
}

// UpdateStatus represents the current update status.
type UpdateStatus struct {
	CurrentVersion  VersionInfo    `json:"current_version"`
	LatestRelease   *LatestRelease `json:"latest_release,omitempty"`
	UpdateAvailable bool           `json:"update_available"`
	LastCheck       *time.Time     `json:"last_check,omitempty"`
	NextCheck       *time.Time     `json:"next_check,omitempty"`
	CheckEnabled    bool           `json:"check_enabled"`
}
