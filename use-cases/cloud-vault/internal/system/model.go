package system

const (
	StatusOK       = "ok"
	StatusWarning  = "warning"
	StatusError    = "error"
	StatusDisabled = "disabled"
)

// HealthResult is the response for GET /api/v1/system/health.
type HealthResult struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Storage  string `json:"storage"`
	Search   string `json:"search"`
	AI       string `json:"ai"`
}

// StatsResult is the response for GET /api/v1/system/stats.
type StatsResult struct {
	Documents        int64 `json:"documents"`
	Versions         int64 `json:"versions"`
	Collections      int64 `json:"collections"`
	Tags             int64 `json:"tags"`
	ImportJobs       int64 `json:"import_jobs"`
	IndexedDocuments int64 `json:"indexed_documents"`
	PendingIndexes   int64 `json:"pending_indexes"`
	FailedIndexes    int64 `json:"failed_indexes"`
	AITasks          int64 `json:"ai_tasks"`
	Prompts          int64 `json:"prompts"`
}

// DoctorRequest is the request body for POST /api/v1/system/doctor.
// If Checks is empty, all checks are run.
type DoctorRequest struct {
	Checks     []string `json:"checks"`
	SampleSize int      `json:"sample_size"` // for hash check; 0 = use default
}

// DoctorResult is the response for POST /api/v1/system/doctor.
type DoctorResult struct {
	Status string        `json:"status"`
	Checks []CheckResult `json:"checks"`
}

// CheckResult holds the outcome of one doctor check.
type CheckResult struct {
	Name   string      `json:"name"`
	Status string      `json:"status"`
	Total  int         `json:"total"`
	Failed int         `json:"failed"`
	Items  []IssueItem `json:"items,omitempty"`
}

// IssueItem describes a single data consistency problem found.
type IssueItem struct {
	DocumentID string `json:"document_id,omitempty"`
	EntityID   string `json:"entity_id,omitempty"`
	Issue      string `json:"issue"`
	Detail     string `json:"detail,omitempty"`
}

// All valid check names.
var AllChecks = []string{
	"storage_objects",
	"document_versions",
	"document_hash",
	"fts_index",
	"tags",
	"collections",
	"sources",
	"imports",
	"ai",
}
