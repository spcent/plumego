package importer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cloud-vault/internal/document"
	"cloud-vault/internal/markdown"
)

// worker processes pending items for a single import job.
type worker struct {
	repo   *Repository
	docSvc *document.Service
}

func newWorker(repo *Repository, docSvc *document.Service) *worker {
	return &worker{repo: repo, docSvc: docSvc}
}

// run processes all pending items for jobID until completion, pause, or cancellation.
func (w *worker) run(ctx context.Context, jobID string) {
	if err := w.repo.UpdateJobStatus(ctx, jobID, JobStatusRunning, nil); err != nil {
		return
	}

	items, err := w.repo.GetPendingItems(ctx, jobID)
	if err != nil {
		_ = w.repo.UpdateJobStatus(ctx, jobID, JobStatusFailed,
			map[string]any{"error_message": err.Error()})
		return
	}

	for _, item := range items {
		// Check context cancellation.
		if ctx.Err() != nil {
			_ = w.repo.UpdateJobStatus(ctx, jobID, JobStatusCancelled, nil)
			return
		}

		// Check if the job has been paused/cancelled in the DB.
		job, err := w.repo.GetJob(ctx, jobID)
		if err != nil || job.Status == JobStatusPaused || job.Status == JobStatusCancelled {
			return
		}

		w.processItem(ctx, jobID, item)
	}

	// Check final outcome.
	finalJob, err := w.repo.GetJob(ctx, jobID)
	if err != nil {
		return
	}
	if finalJob.Status == JobStatusPaused || finalJob.Status == JobStatusCancelled {
		return
	}
	_ = w.repo.UpdateJobStatus(ctx, jobID, JobStatusDone, nil)
}

func (w *worker) processItem(ctx context.Context, jobID string, item *ImportJobItem) {
	content, err := readFile(item.FilePath)
	if err != nil {
		_ = w.repo.UpdateItemStatus(ctx, item.ID, ItemStatusFailed, "", err.Error())
		_ = w.repo.IncrementJobCounts(ctx, jobID, 0, 1, 0)
		return
	}

	// SHA-256 duplicate check.
	hash := document.HashContent(content)
	exists, err := w.docSvc.ExistsByHash(ctx, hash)
	if err != nil {
		_ = w.repo.UpdateItemStatus(ctx, item.ID, ItemStatusFailed, "", err.Error())
		_ = w.repo.IncrementJobCounts(ctx, jobID, 0, 1, 0)
		return
	}
	if exists {
		_ = w.repo.UpdateItemStatus(ctx, item.ID, ItemStatusSkipped, "", "duplicate content")
		_ = w.repo.IncrementJobCounts(ctx, jobID, 0, 0, 1)
		return
	}

	// Extract title from filename / content.
	title := markdown.ExtractTitle(content)
	if title == "" {
		base := filepath.Base(item.FilePath)
		title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Create document.
	result, err := w.docSvc.CreateFromImport(ctx, document.CreateFromImportRequest{
		Title:        title,
		Content:      content,
		OriginalPath: item.FilePath,
		ImportJobID:  jobID,
	})
	if err != nil {
		_ = w.repo.UpdateItemStatus(ctx, item.ID, ItemStatusFailed, "", err.Error())
		_ = w.repo.IncrementJobCounts(ctx, jobID, 0, 1, 0)
		return
	}

	// Save rich metadata.
	meta := markdown.Parse(content)
	dm := &DocumentMetadata{
		ID:             newID(),
		DocumentID:     result.ID,
		Headings:       meta.HeadingsJSON(),
		CodeLanguages:  meta.CodeLanguagesJSON(),
		CodeBlockCount: meta.CodeBlockCount,
		LinkCount:      meta.LinkCount,
		ImageCount:     meta.ImageCount,
		ExtractedAt:    result.UpdatedAt,
	}
	_ = w.repo.UpsertMetadata(ctx, dm)

	_ = w.repo.UpdateItemStatus(ctx, item.ID, ItemStatusSuccess, result.ID, "")
	_ = w.repo.IncrementJobCounts(ctx, jobID, 1, 0, 0)
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}
	return string(b), nil
}
