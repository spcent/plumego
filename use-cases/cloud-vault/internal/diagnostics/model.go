package diagnostics

import "time"

// Bundle represents a diagnostic bundle metadata.
type Bundle struct {
	ID           string    `json:"id"`
	Filename     string    `json:"filename"`
	CreatedAt    time.Time `json:"created_at"`
	Size         int64     `json:"size_bytes"`
	DownloadPath string    `json:"download_path"`
}

// RedactedConfig represents configuration with sensitive fields redacted.
type RedactedConfig struct {
	Server   map[string]interface{} `json:"server"`
	App      map[string]interface{} `json:"app"`
	Database map[string]interface{} `json:"database"`
	Storage  map[string]interface{} `json:"storage"`
	Auth     map[string]interface{} `json:"auth"`
	AI       map[string]interface{} `json:"ai"`
	Update   map[string]interface{} `json:"update"`
}

// BuildInfo contains version and build metadata.
type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
	Channel   string `json:"channel"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// SystemInfo contains runtime system information.
type SystemInfo struct {
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	NumCPU       int    `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
	MemAllocMB   uint64 `json:"mem_alloc_mb"`
	MemSysMB     uint64 `json:"mem_sys_mb"`
	Uptime       string `json:"uptime"`
}

// LogEntry represents a single log line.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}
