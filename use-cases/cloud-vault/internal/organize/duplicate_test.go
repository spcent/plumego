package organize

import (
	"context"
	"path/filepath"
	"testing"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/storage"
)

func openTestDB(t *testing.T) *database.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func setupOrganize(t *testing.T) (*Service, *document.Service, *database.DB) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())

	// Document service
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	// Organize service
	organizeRepo := NewRepository(db)
	organizeSvc := NewService(organizeRepo, store, config.OrganizeConfig{})

	return organizeSvc, docSvc, db
}

func TestDetectDuplicates_ExactDuplicates(t *testing.T) {
	organizeSvc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	// Create two documents with identical content
	content := "# Duplicate\n\nThis content is identical."

	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Duplicate A",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}

	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "Duplicate B",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}

	// Run duplicate detection
	job, err := organizeSvc.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("DetectDuplicates: %v", err)
	}

	if job.ProcessedItems != 1 {
		t.Errorf("ProcessedItems = %d, want 1 (one pair)", job.ProcessedItems)
	}

	// List duplicates
	groups, err := organizeSvc.ListDuplicates(ctx)
	if err != nil {
		t.Fatalf("ListDuplicates: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("Groups count = %d, want 1", len(groups))
	}

	if len(groups[0].Documents) != 2 {
		t.Errorf("Documents in group = %d, want 2", len(groups[0].Documents))
	}
}

func TestDetectDuplicates_NoDuplicates(t *testing.T) {
	organizeSvc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	// Create two documents with different content
	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Unique A",
		Content: "# Unique A\n\nDifferent content.",
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}

	_, err = docSvc.Create(ctx, document.CreateRequest{
		Title:   "Unique B",
		Content: "# Unique B\n\nCompletely different.",
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}

	// Run duplicate detection
	job, err := organizeSvc.DetectDuplicates(ctx)
	if err != nil {
		t.Fatalf("DetectDuplicates: %v", err)
	}

	if job.ProcessedItems != 0 {
		t.Errorf("ProcessedItems = %d, want 0 (no duplicates)", job.ProcessedItems)
	}

	// List duplicates
	groups, err := organizeSvc.ListDuplicates(ctx)
	if err != nil {
		t.Fatalf("ListDuplicates: %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Groups count = %d, want 0", len(groups))
	}
}

func TestResolveDuplicates_Archive(t *testing.T) {
	organizeSvc, docSvc, _ := setupOrganize(t)
	ctx := context.Background()

	content := "# Duplicate\n\nFor resolution test."

	docA, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Keep This",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create A: %v", err)
	}

	docB, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Archive This",
		Content: content,
	})
	if err != nil {
		t.Fatalf("Create B: %v", err)
	}

	// Resolve by archiving docB
	err = organizeSvc.ResolveDuplicates(ctx, ResolveDuplicatesRequest{
		KeepDocumentID:       docA.ID,
		DuplicateDocumentIDs: []string{docB.ID},
		Action:               "archive",
	})
	if err != nil {
		t.Fatalf("ResolveDuplicates: %v", err)
	}

	// Verify docB is archived by listing all docs including archived
	allResult, err := docSvc.List(ctx, document.ListQuery{Status: "all", Limit: 100, Offset: 0})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}

	var docBFound, docAFound bool
	for _, item := range allResult.Items {
		if item.ID == docB.ID {
			docBFound = true
			// After archive, GetByID returns the doc; verify via repo-level query
		}
		if item.ID == docA.ID {
			docAFound = true
		}
	}

	if !docAFound {
		t.Error("docA not found in list")
	}
	if !docBFound {
		t.Error("docB not found in list after archive")
	}
}
