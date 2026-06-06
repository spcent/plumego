package importer

import "time"

// CreateJobRequest creates a new import job.
type CreateJobRequest struct {
	Name     string `json:"name"`
	SourceID string `json:"source_id"`
}

// SourceDirectory is a server-discovered import directory.
type SourceDirectory struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	RelPath string `json:"rel_path"`
}

// ListSourcesResponse lists directories that can be imported.
type ListSourcesResponse struct {
	Items []SourceDirectory `json:"items"`
}

// JobResponse is the API representation of an import job.
type JobResponse struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	SourcePath     string     `json:"source_path"`
	Status         string     `json:"status"`
	TotalCount     int        `json:"total_count"`
	ProcessedCount int        `json:"processed_count"`
	SuccessCount   int        `json:"success_count"`
	FailedCount    int        `json:"failed_count"`
	SkippedCount   int        `json:"skipped_count"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	StartedAt      *time.Time `json:"started_at,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ListJobsResponse is the paginated list of import jobs.
type ListJobsResponse struct {
	Items  []JobResponse `json:"items"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// ItemResponse is the API representation of an import job item.
type ItemResponse struct {
	ID           string    `json:"id"`
	JobID        string    `json:"job_id"`
	FilePath     string    `json:"file_path"`
	DocumentID   string    `json:"document_id,omitempty"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListItemsResponse is the paginated list of job items.
type ListItemsResponse struct {
	Items  []ItemResponse `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func toJobResponse(j *ImportJob) JobResponse {
	return JobResponse{
		ID:             j.ID,
		Name:           j.Name,
		SourcePath:     j.SourcePath,
		Status:         j.Status,
		TotalCount:     j.TotalCount,
		ProcessedCount: j.ProcessedCount,
		SuccessCount:   j.SuccessCount,
		FailedCount:    j.FailedCount,
		SkippedCount:   j.SkippedCount,
		ErrorMessage:   j.ErrorMessage,
		StartedAt:      j.StartedAt,
		CompletedAt:    j.CompletedAt,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
}

func toItemResponse(item *ImportJobItem) ItemResponse {
	return ItemResponse{
		ID:           item.ID,
		JobID:        item.JobID,
		FilePath:     item.FilePath,
		DocumentID:   item.DocumentID,
		Status:       item.Status,
		ErrorMessage: item.ErrorMessage,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
