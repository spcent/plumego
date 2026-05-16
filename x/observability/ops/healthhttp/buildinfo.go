package healthhttp

// BuildInfo describes the version metadata of the running binary.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

// These variables can be overridden at build time using -ldflags.
var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

// GetBuildInfo returns the current build metadata.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
	}
}
