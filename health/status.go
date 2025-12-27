package health

import "sync/atomic"

// BuildInfo describes the version metadata of the running binary.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// These variables can be overridden at build time using -ldflags.
var (
	// Version reports the application version.
	Version = version
	// Commit reports the git commit hash used for the build.
	Commit = commit
	// BuildTime reports when the binary was built.
	BuildTime = buildTime
)

// ReadinessStatus describes whether the application is ready to serve traffic.
type ReadinessStatus struct {
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}

var readiness atomic.Value

func init() {
	readiness.Store(ReadinessStatus{Ready: false, Reason: "starting"})
}

// GetBuildInfo returns the current build metadata.
func GetBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, BuildTime: BuildTime}
}

// SetReady marks the application as ready to serve traffic.
func SetReady() {
	readiness.Store(ReadinessStatus{Ready: true})
}

// SetNotReady marks the application as not ready and records the reason.
func SetNotReady(reason string) {
	readiness.Store(ReadinessStatus{Ready: false, Reason: reason})
}

// GetReadiness returns the current readiness status.
func GetReadiness() ReadinessStatus {
	if status, ok := readiness.Load().(ReadinessStatus); ok {
		return status
	}
	return ReadinessStatus{Ready: false, Reason: "unknown"}
}
