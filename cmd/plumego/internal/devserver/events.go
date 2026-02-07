package devserver

import "time"

// Event types for development workflow
const (
	EventFileChange   = "file.change"
	EventBuildStart   = "build.start"
	EventBuildSuccess = "build.success"
	EventBuildFail    = "build.fail"
	EventAppStart     = "app.start"
	EventAppStop      = "app.stop"
	EventAppRestart   = "app.restart"
	EventAppLog       = "app.log"
	EventAppError     = "app.error"
	EventAppHealth    = "app.health"
	EventDashboard    = "dashboard.info"
)

// Event is a generic event structure
type Event struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Data      any       `json:"data,omitempty"`
}

// NewEvent creates a new event
func NewEvent(eventType string, data any) Event {
	return Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// FileChangeEvent represents a file change
type FileChangeEvent struct {
	Path   string `json:"path"`
	Action string `json:"action"` // "modify", "create", "delete"
}

// BuildEvent represents a build event
type BuildEvent struct {
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error,omitempty"`
	Output   string        `json:"output,omitempty"`
}

// AppLifecycleEvent represents app start/stop/restart
type AppLifecycleEvent struct {
	State string `json:"state"` // "starting", "running", "stopped", "crashed"
	PID   int    `json:"pid,omitempty"`
	Error string `json:"error,omitempty"`
}

// LogEvent represents a log message from the application
type LogEvent struct {
	Level   string `json:"level"` // "info", "warn", "error", "debug"
	Message string `json:"message"`
	Source  string `json:"source"` // "stdout", "stderr"
}

// HealthEvent represents application health status
type HealthEvent struct {
	Healthy bool              `json:"healthy"`
	Checks  map[string]string `json:"checks,omitempty"`
}

// DashboardInfo provides dashboard metadata
type DashboardInfo struct {
	Version      string `json:"version"`
	DashboardURL string `json:"dashboard_url"`
	AppURL       string `json:"app_url"`
	Uptime       string `json:"uptime"`
	UptimeMS     int64  `json:"uptime_ms"`
	StartTime    string `json:"start_time"`
	ProjectDir   string `json:"project_dir"`
	GoVersion    string `json:"go_version"`
	AppRunning   bool   `json:"app_running"`
	AppPID       int    `json:"app_pid,omitempty"`
}
