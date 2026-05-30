// Package version provides build-time version information.
//
// Version, Commit, BuildTime, and Channel are set via ldflags during build:
//
//	go build -ldflags "-X cloud-vault/internal/version.Version=1.0.0 \
//	  -X cloud-vault/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X cloud-vault/internal/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
//	  -X cloud-vault/internal/version.Channel=stable"
package version

// Version is the semantic version (e.g., "1.0.0").
var Version = "1.0.0"

// Commit is the git commit hash (short form) at build time.
var Commit = "dev"

// BuildTime is the UTC timestamp when the binary was built.
var BuildTime = "unknown"

// Channel is the release channel (e.g., "stable", "beta", "dev").
var Channel = "stable"

// BuildInfo contains all version information.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	Channel   string `json:"channel"`
}

// GetBuildInfo returns the current build information.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
		Channel:   Channel,
	}
}
