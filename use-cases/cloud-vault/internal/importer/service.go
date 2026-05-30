package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud-vault/internal/document"
)

const maxScanFiles = 10000
const importerSafeRootEnv = "IMPORTER_SAFE_ROOT"

func safeImportRoot() string {
	if v := strings.TrimSpace(os.Getenv(importerSafeRootEnv)); v != "" {
		return v
	}
	return "/data/imports"
}

func resolveAndValidateSourcePath(input string) (string, error) {
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return "", fmt.Errorf("source_path must not be empty")
	}

	baseAbs, err := filepath.Abs(filepath.Clean(safeImportRoot()))
	if err != nil {
		return "", fmt.Errorf("resolve safe root: %w", err)
	}

	resolvedAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanInput))
	if err != nil {
		return "", fmt.Errorf("resolve source_path: %w", err)
	}

	basePrefix := baseAbs + string(os.PathSeparator)
	if resolvedAbs != baseAbs && !strings.HasPrefix(resolvedAbs, basePrefix) {
		return "", fmt.Errorf("source_path must be within %q", baseAbs)
	}

	return resolvedAbs, nil
}

// Config holds importer configuration.
type Config struct {
	MaxFileSizeMB int64 // unused in scanner but reserved
}

// Service manages import job lifecycle and background workers.
type Service struct {
	repo    *Repository
	docSvc  *document.Service
	cfg     Config
	workers map[string]context.CancelFunc
	mu      sync.Mutex
}

func NewService(repo *Repository, docSvc *document.Service, cfg Config) *Service {
	return &Service{
		repo:    repo,
		docSvc:  docSvc,
		cfg:     cfg,
		workers: make(map[string]context.CancelFunc),
	}
}

// CreateJob scans the source directory and registers a new import job.
func (s *Service) CreateJob(ctx context.Context, req CreateJobRequest) (*JobResponse, error) {
	sourcePath, err := resolveAndValidateSourcePath(req.SourcePath)
	if err != nil {
		return nil, err
	}
	if req.Name == "" {
		req.Name = req.SourcePath
	}

	files, err := ScanDirectory(sourcePath, maxScanFiles)
	if err != nil {
		return nil, fmt.Errorf("scan directory: %w", err)
	}

	now := time.Now().UTC()
	job := &ImportJob{
		ID:         newID(),
		Name:       req.Name,
		SourcePath: sourcePath,
		Status:     JobStatusPending,
		TotalCount: len(files),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.repo.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Create items for all scanned files.
	items := make([]*ImportJobItem, 0, len(files))
	for _, f := range files {
		items = append(items, &ImportJobItem{
			ID:       newID(),
			JobID:    job.ID,
			FilePath: f.AbsPath,
			Status:   ItemStatusPending,
		})
	}
	if err := s.repo.BulkCreateItems(ctx, items); err != nil {
		return nil, fmt.Errorf("create job items: %w", err)
	}

	resp := toJobResponse(job)
	return &resp, nil
}

// GetJob returns a job by ID.
func (s *Service) GetJob(ctx context.Context, id string) (*JobResponse, error) {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toJobResponse(job)
	return &resp, nil
}

// ListJobs returns paginated import jobs.
func (s *Service) ListJobs(ctx context.Context, limit, offset int) (*ListJobsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	jobs, total, err := s.repo.ListJobs(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	items := make([]JobResponse, 0, len(jobs))
	for _, j := range jobs {
		items = append(items, toJobResponse(j))
	}
	return &ListJobsResponse{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

// StartJob launches the background worker for a job.
func (s *Service) StartJob(ctx context.Context, id string) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job.Status == JobStatusRunning {
		return fmt.Errorf("job is already running")
	}
	if job.Status == JobStatusDone || job.Status == JobStatusCancelled {
		return fmt.Errorf("job has already completed or been cancelled")
	}

	workerCtx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.workers[id] = cancel
	s.mu.Unlock()

	w := newWorker(s.repo, s.docSvc)
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.workers, id)
			s.mu.Unlock()
		}()
		w.run(workerCtx, id)
	}()

	return nil
}

// PauseJob marks the job as paused; the worker stops after the current item.
func (s *Service) PauseJob(ctx context.Context, id string) error {
	job, err := s.repo.GetJob(ctx, id)
	if err != nil {
		return err
	}
	if job.Status != JobStatusRunning && job.Status != JobStatusPending {
		return fmt.Errorf("job cannot be paused (status: %s)", job.Status)
	}
	return s.repo.UpdateJobStatus(ctx, id, JobStatusPaused, nil)
}

// CancelJob stops the worker and marks the job cancelled.
func (s *Service) CancelJob(ctx context.Context, id string) error {
	s.mu.Lock()
	cancel, running := s.workers[id]
	s.mu.Unlock()

	if running {
		cancel()
	}
	return s.repo.UpdateJobStatus(ctx, id, JobStatusCancelled, nil)
}

// RetryJob resets failed items and restarts the worker.
func (s *Service) RetryJob(ctx context.Context, id string) error {
	n, err := s.repo.ResetFailedItems(ctx, id)
	if err != nil {
		return fmt.Errorf("reset failed items: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("no failed items to retry")
	}
	// Update job total to account for newly-pending items.
	_ = s.repo.UpdateJobStatus(ctx, id, JobStatusPending, nil)
	return s.StartJob(ctx, id)
}

// ListItems returns paginated items for a job.
func (s *Service) ListItems(ctx context.Context, jobID, statusFilter string, limit, offset int) (*ListItemsResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	items, total, err := s.repo.ListItems(ctx, jobID, statusFilter, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := make([]ItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, toItemResponse(item))
	}
	return &ListItemsResponse{Items: resp, Total: total, Limit: limit, Offset: offset}, nil
}
