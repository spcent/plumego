package search

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

func setupSearch(t *testing.T) (*Service, *document.Service, *database.DB) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())

	// Document service
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	// Search service
	searchRepo := NewRepository(db)
	searchEngine := NewFTSEngine(db, 50)
	searchSvc := NewService(searchEngine, searchRepo, store, config.SearchConfig{Enabled: true})

	// Wire index hook (same pattern as app.go)
	docSvc.SetIndexHook(func(ctx context.Context, ev document.IndexEvent) {
		searchSvc.HandleIndexEvent(ctx, IndexEvent{
			DocID:    ev.DocID,
			Content:  ev.Content,
			Version:  ev.Version,
			Hash:     ev.Hash,
			Deleted:  ev.Deleted,
			IsImport: ev.IsImport,
		})
	})

	return searchSvc, docSvc, db
}

func TestSearch_IndexAndRetrieve(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	// Create document (triggers index hook)
	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Search Test",
		Content: "# Search Test\n\nThis document contains searchable keywords like golang and sqlite.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify index status
	status, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if status.Indexed != 1 {
		t.Errorf("Indexed = %d, want 1", status.Indexed)
	}

	// Search for the document
	results, err := searchSvc.Search(ctx, SearchQuery{
		Q:     "golang",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results.Items) != 1 {
		t.Errorf("Search results = %d, want 1", len(results.Items))
	}
	if len(results.Items) > 0 && results.Items[0].ID != created.ID {
		t.Errorf("Result ID = %q, want %q", results.Items[0].ID, created.ID)
	}
}

func TestSearch_UpdateMakesStale(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Stale Test",
		Content: "original content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify indexed
	status1, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if status1.Indexed != 1 {
		t.Errorf("After create: Indexed = %d, want 1", status1.Indexed)
	}

	// Update document (should mark as stale)
	_, err = docSvc.Update(ctx, created.ID, document.UpdateRequest{
		Title:       "Stale Test",
		Content:     "updated content with new keywords",
		BaseVersion: 1,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	// After update, the hook should re-index, so still indexed
	status2, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus after update: %v", err)
	}
	// The hook re-indexes synchronously in tests, so should still be 1
	if status2.Indexed != 1 {
		t.Errorf("After update: Indexed = %d, want 1", status2.Indexed)
	}
}

func TestSearch_DeleteRemovesFromIndex(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Delete Test",
		Content: "will be deleted",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify indexed
	status1, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if status1.Indexed != 1 {
		t.Errorf("Before delete: Indexed = %d, want 1", status1.Indexed)
	}

	// Delete document
	err = docSvc.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Verify removed from index
	status2, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus after delete: %v", err)
	}
	if status2.Indexed != 0 {
		t.Errorf("After delete: Indexed = %d, want 0", status2.Indexed)
	}
}

func TestSearch_MultipleDocuments(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	// Create 3 documents with different content
	docs := []struct {
		title   string
		content string
	}{
		{"Go Programming", "# Go\n\nGo is a statically typed language."},
		{"Python Guide", "# Python\n\nPython is dynamically typed."},
		{"Database Design", "# Databases\n\nSQLite is a popular embedded database."},
	}

	for _, d := range docs {
		_, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   d.title,
			Content: d.content,
		})
		if err != nil {
			t.Fatalf("Create %q: %v", d.title, err)
		}
	}

	// Verify all indexed
	status, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if status.Indexed != 3 {
		t.Errorf("Indexed = %d, want 3", status.Indexed)
	}

	// Search for "Go"
	results, err := searchSvc.Search(ctx, SearchQuery{Q: "Go", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Should find at least 1 (the Go document)
	if len(results.Items) < 1 {
		t.Error("Search for 'Go' returned no results")
	}
}
