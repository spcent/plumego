package search

import (
	"context"
	"testing"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/storage"
)

// openTestDB and setupSearch are defined in index_test.go.

// createDocForIndex is a helper that creates a real document and returns its ID.
func createDocForIndex(t *testing.T, db *database.DB, store storage.ObjectStorage, title string) string {
	t.Helper()
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)
	result, err := docSvc.Create(context.Background(), document.CreateRequest{
		Title:   title,
		Content: "# " + title + "\n\nContent for index test.",
	})
	if err != nil {
		t.Fatalf("createDocForIndex %q: %v", title, err)
	}
	return result.ID
}

// --- Repository tests ---

func TestRepository_UpsertIndexStatus_InsertAndUpdate(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docID := createDocForIndex(t, db, store, "Upsert Test")

	// Insert a status record.
	if err := repo.UpsertIndexStatus(ctx, docID, "hash-1", 1, IndexStatusPending, ""); err != nil {
		t.Fatalf("UpsertIndexStatus insert: %v", err)
	}

	// Update same doc to indexed.
	if err := repo.UpsertIndexStatus(ctx, docID, "hash-1", 1, IndexStatusIndexed, ""); err != nil {
		t.Fatalf("UpsertIndexStatus update: %v", err)
	}

	// Verify via summary.
	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Indexed != 1 {
		t.Errorf("Indexed: got %d, want 1", summary.Indexed)
	}
}

func TestRepository_UpsertIndexStatus_Failed(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docID := createDocForIndex(t, db, store, "Failed Index Test")

	if err := repo.UpsertIndexStatus(ctx, docID, "hash-f", 1, IndexStatusFailed, "some error"); err != nil {
		t.Fatalf("UpsertIndexStatus failed: %v", err)
	}

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed: got %d, want 1", summary.Failed)
	}
}

func TestRepository_DeleteIndexStatus(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docID := createDocForIndex(t, db, store, "Delete Index Test")

	if err := repo.UpsertIndexStatus(ctx, docID, "hash-d", 1, IndexStatusIndexed, ""); err != nil {
		t.Fatalf("UpsertIndexStatus: %v", err)
	}

	if err := repo.DeleteIndexStatus(ctx, docID); err != nil {
		t.Fatalf("DeleteIndexStatus: %v", err)
	}

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Indexed != 0 {
		t.Errorf("After delete: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestRepository_MarkStatusByScope_All(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	// Insert 3 indexed docs.
	for i := 1; i <= 3; i++ {
		docID := createDocForIndex(t, db, store, "Scope Doc "+string(rune('0'+i)))
		hash := "hash-s" + string(rune('0'+i))
		if err := repo.UpsertIndexStatus(ctx, docID, hash, 1, IndexStatusIndexed, ""); err != nil {
			t.Fatalf("UpsertIndexStatus: %v", err)
		}
	}

	// Mark all as pending.
	if err := repo.MarkStatusByScope(ctx, "all", ""); err != nil {
		t.Fatalf("MarkStatusByScope all: %v", err)
	}

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	// After marking all pending, Indexed should be 0.
	if summary.Indexed != 0 {
		t.Errorf("After mark-all: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestRepository_MarkStatusByScope_Failed(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docIDFail := createDocForIndex(t, db, store, "Mark Failed Doc")
	docIDOK := createDocForIndex(t, db, store, "Mark Indexed Doc")

	if err := repo.UpsertIndexStatus(ctx, docIDFail, "hash-mf1", 1, IndexStatusFailed, "oops"); err != nil {
		t.Fatalf("UpsertIndexStatus: %v", err)
	}
	if err := repo.UpsertIndexStatus(ctx, docIDOK, "hash-mi1", 1, IndexStatusIndexed, ""); err != nil {
		t.Fatalf("UpsertIndexStatus: %v", err)
	}

	if err := repo.MarkStatusByScope(ctx, "failed", ""); err != nil {
		t.Fatalf("MarkStatusByScope failed: %v", err)
	}

	// Failed should have moved to pending, so failed count is 0.
	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Failed != 0 {
		t.Errorf("After mark-failed: Failed = %d, want 0", summary.Failed)
	}
}

func TestRepository_MarkStatusByScope_Document(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docID := createDocForIndex(t, db, store, "Mark Document Doc")

	if err := repo.UpsertIndexStatus(ctx, docID, "hash-md1", 1, IndexStatusIndexed, ""); err != nil {
		t.Fatalf("UpsertIndexStatus: %v", err)
	}

	if err := repo.MarkStatusByScope(ctx, "document", docID); err != nil {
		t.Fatalf("MarkStatusByScope document: %v", err)
	}
}

func TestRepository_MarkStatusByScope_Stale(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewRepository(db)
	ctx := context.Background()

	docID := createDocForIndex(t, db, store, "Mark Stale Doc")

	if err := repo.UpsertIndexStatus(ctx, docID, "hash-stale", 1, IndexStatusStale, ""); err != nil {
		t.Fatalf("UpsertIndexStatus: %v", err)
	}

	if err := repo.MarkStatusByScope(ctx, "stale", ""); err != nil {
		t.Fatalf("MarkStatusByScope stale: %v", err)
	}
}

func TestRepository_MarkStatusByScope_Unknown(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	err := repo.MarkStatusByScope(ctx, "bogus-scope", "")
	if err == nil {
		t.Error("MarkStatusByScope unknown scope: expected error, got nil")
	}
}

func TestRepository_GetIndexStatusSummary_Empty(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary empty: %v", err)
	}
	if summary.TotalDocuments != 0 {
		t.Errorf("TotalDocuments: got %d, want 0", summary.TotalDocuments)
	}
	if summary.Indexed != 0 {
		t.Errorf("Indexed: got %d, want 0", summary.Indexed)
	}
	if summary.LastIndexedAt != nil {
		t.Errorf("LastIndexedAt: got %v, want nil", summary.LastIndexedAt)
	}
}

func TestRepository_GetPendingDocs_Empty(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	docs, err := repo.GetPendingDocs(ctx, 100)
	if err != nil {
		t.Fatalf("GetPendingDocs: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("GetPendingDocs empty: got %d, want 0", len(docs))
	}
}

func TestRepository_GetPendingDocs_WithDocuments(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	store := storage.NewLocalStorage(t.TempDir())
	ctx := context.Background()

	// Create a document via document service so it exists in documents table.
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)
	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Pending Doc",
		Content: "# Pending\n\nThis doc needs indexing.",
	})
	if err != nil {
		t.Fatalf("Create doc: %v", err)
	}

	// Should appear in pending (never indexed).
	pending, err := repo.GetPendingDocs(ctx, 10)
	if err != nil {
		t.Fatalf("GetPendingDocs: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("GetPendingDocs: got %d, want 1", len(pending))
	}
}

func TestRepository_SaveHistory_GetHistory_ClearHistory(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Save 3 history records.
	for i := 1; i <= 3; i++ {
		rec := SearchHistoryRecord{
			ID:          "hist-" + string(rune('0'+i)),
			Query:       "query " + string(rune('0'+i)),
			ResultCount: i * 5,
			CreatedAt:   time.Now().UTC(),
		}
		if err := repo.SaveHistory(ctx, rec); err != nil {
			t.Fatalf("SaveHistory %d: %v", i, err)
		}
	}

	// Get history.
	records, err := repo.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("GetHistory count: got %d, want 3", len(records))
	}

	// Clear history.
	if err := repo.ClearHistory(ctx); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}
	records, err = repo.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory after clear: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("GetHistory after clear: got %d, want 0", len(records))
	}
}

func TestRepository_GetHistory_WithFiltersJSON(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	rec := SearchHistoryRecord{
		ID:          "hist-f1",
		Query:       "filtered query",
		FiltersJSON: `{"status":"active"}`,
		ResultCount: 10,
		CreatedAt:   time.Now().UTC(),
	}
	if err := repo.SaveHistory(ctx, rec); err != nil {
		t.Fatalf("SaveHistory: %v", err)
	}

	records, err := repo.GetHistory(ctx, 10)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("GetHistory count: got %d, want 1", len(records))
	}
	if records[0].FiltersJSON != `{"status":"active"}` {
		t.Errorf("FiltersJSON: got %q, want %q", records[0].FiltersJSON, `{"status":"active"}`)
	}
}

func TestRepository_GetTagsForDocuments_Empty(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	result, err := repo.GetTagsForDocuments(ctx, []string{})
	if err != nil {
		t.Fatalf("GetTagsForDocuments empty: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("GetTagsForDocuments empty: got %d, want 0", len(result))
	}
}

func TestRepository_GetTagsForDocuments_WithDocs(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	// Should return empty map for docs with no tags.
	result, err := repo.GetTagsForDocuments(ctx, []string{"doc-no-tags"})
	if err != nil {
		t.Fatalf("GetTagsForDocuments no tags: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("GetTagsForDocuments no tags: got %d, want 0 entries", len(result))
	}
}

// --- FTSEngine tests ---

func TestFTSEngine_IndexAndSearch(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	// Need documents in the documents table for JOIN to work.
	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Golang Testing",
		Content: "# Golang Testing\n\nThis document is about golang unit tests.",
	})
	if err != nil {
		t.Fatalf("Create doc: %v", err)
	}

	// Index the document.
	sdoc := SearchDocument{
		DocumentID: created.ID,
		Title:      "Golang Testing",
		Content:    "This document is about golang unit tests.",
	}
	if err := engine.IndexDocument(ctx, sdoc); err != nil {
		t.Fatalf("IndexDocument: %v", err)
	}

	// Search for it.
	result, err := engine.Search(ctx, SearchQuery{Q: "golang", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Search results: got %d, want 1", len(result.Items))
	}
	if result.Total != 1 {
		t.Errorf("Search total: got %d, want 1", result.Total)
	}
	if len(result.Items) > 0 && result.Items[0].ID != created.ID {
		t.Errorf("Search result ID: got %q, want %q", result.Items[0].ID, created.ID)
	}
}

func TestFTSEngine_IndexDocument_Update(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Updateable Doc",
		Content: "# Updateable\n\nInitial content.",
	})
	if err != nil {
		t.Fatalf("Create doc: %v", err)
	}

	// Index initial.
	if err := engine.IndexDocument(ctx, SearchDocument{
		DocumentID: created.ID,
		Title:      "Updateable Doc",
		Content:    "Initial content.",
	}); err != nil {
		t.Fatalf("Initial IndexDocument: %v", err)
	}

	// Re-index with updated content.
	if err := engine.IndexDocument(ctx, SearchDocument{
		DocumentID: created.ID,
		Title:      "Updateable Doc",
		Content:    "Completely different keywords here.",
	}); err != nil {
		t.Fatalf("Re-index: %v", err)
	}

	// Old content should not match.
	r1, err := engine.Search(ctx, SearchQuery{Q: "Initial", Limit: 10})
	if err != nil {
		t.Fatalf("Search old: %v", err)
	}
	if len(r1.Items) != 0 {
		t.Errorf("Old content search: got %d items, want 0", len(r1.Items))
	}

	// New content should match.
	r2, err := engine.Search(ctx, SearchQuery{Q: "keywords", Limit: 10})
	if err != nil {
		t.Fatalf("Search new: %v", err)
	}
	if len(r2.Items) != 1 {
		t.Errorf("New content search: got %d items, want 1", len(r2.Items))
	}
}

func TestFTSEngine_DeleteDocument(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "To Be Deleted",
		Content: "# Delete Me\n\nSearchable content.",
	})
	if err != nil {
		t.Fatalf("Create doc: %v", err)
	}

	if err := engine.IndexDocument(ctx, SearchDocument{
		DocumentID: created.ID,
		Title:      "To Be Deleted",
		Content:    "Searchable content.",
	}); err != nil {
		t.Fatalf("IndexDocument: %v", err)
	}

	// Delete from index.
	if err := engine.DeleteDocument(ctx, created.ID); err != nil {
		t.Fatalf("DeleteDocument: %v", err)
	}

	// Should no longer be findable.
	result, err := engine.Search(ctx, SearchQuery{Q: "Searchable", Limit: 10})
	if err != nil {
		t.Fatalf("Search after delete: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("After delete search: got %d items, want 0", len(result.Items))
	}
}

func TestFTSEngine_Search_EmptyQuery_Fallback(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	for i := 0; i < 3; i++ {
		_, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   "Fallback Doc " + string(rune('A'+i)),
			Content: "Content " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("Create doc %d: %v", i, err)
		}
	}

	// Empty query should use fallback list.
	result, err := engine.Search(ctx, SearchQuery{Q: "", Limit: 10})
	if err != nil {
		t.Fatalf("Search empty: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Fallback search total: got %d, want 3", result.Total)
	}
}

func TestFTSEngine_Search_Pagination(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	for i := 0; i < 5; i++ {
		_, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   "Paginate " + string(rune('A'+i)),
			Content: "paginate test content",
		})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// Empty query fallback with limit.
	r1, err := engine.Search(ctx, SearchQuery{Q: "", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("Search page 1: %v", err)
	}
	if len(r1.Items) != 2 {
		t.Errorf("Page 1 count: got %d, want 2", len(r1.Items))
	}
	if r1.Total != 5 {
		t.Errorf("Page 1 total: got %d, want 5", r1.Total)
	}

	r2, err := engine.Search(ctx, SearchQuery{Q: "", Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("Search page 3: %v", err)
	}
	if len(r2.Items) != 1 {
		t.Errorf("Page 3 count: got %d, want 1", len(r2.Items))
	}
}

func TestFTSEngine_Search_LimitCap(t *testing.T) {
	db := openTestDB(t)
	engine := NewFTSEngine(db, 20)
	ctx := context.Background()

	// Limit of 0 should default to 20, >100 should cap at 100.
	r1, err := engine.Search(ctx, SearchQuery{Q: "", Limit: 0})
	if err != nil {
		t.Fatalf("Search limit 0: %v", err)
	}
	if r1.Limit != 20 {
		t.Errorf("Default limit: got %d, want 20", r1.Limit)
	}

	r2, err := engine.Search(ctx, SearchQuery{Q: "", Limit: 200})
	if err != nil {
		t.Fatalf("Search limit 200: %v", err)
	}
	if r2.Limit != 100 {
		t.Errorf("Capped limit: got %d, want 100", r2.Limit)
	}
}

// --- Service tests ---

func TestService_HandleIndexEvent_Disabled(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: false})
	ctx := context.Background()

	// Should be a no-op when disabled.
	svc.HandleIndexEvent(ctx, IndexEvent{DocID: "doc-1", Content: "content", Version: 1, Hash: "hash"})

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	// Nothing should have been indexed.
	if summary.Indexed != 0 {
		t.Errorf("Indexed when disabled: got %d, want 0", summary.Indexed)
	}
}

func TestService_HandleIndexEvent_Delete(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true, IndexOnSave: true})
	ctx := context.Background()

	// First index a document.
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)
	created, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Delete Event Test",
		Content: "content to delete",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	svc.HandleIndexEvent(ctx, IndexEvent{
		DocID: created.ID, Content: "content to delete", Version: 1, Hash: "hash-del",
	})

	// Now fire a delete event.
	svc.HandleIndexEvent(ctx, IndexEvent{DocID: created.ID, Deleted: true})

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Indexed != 0 {
		t.Errorf("After delete event: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestService_HandleIndexEvent_Deferred(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	// IndexOnSave=false → deferred to background indexer.
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true, IndexOnSave: false})
	ctx := context.Background()

	svc.HandleIndexEvent(ctx, IndexEvent{
		DocID: "doc-deferred", Content: "some content", Version: 1, Hash: "hash-deferred",
	})

	// Should be recorded as pending, not indexed.
	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	if summary.Indexed != 0 {
		t.Errorf("Deferred: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestService_HandleIndexEvent_Import_Deferred(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	// IndexOnSave=true but import goes to pending unless IndexOnImport=inline.
	svc := NewService(engine, repo, store, config.SearchConfig{
		Enabled:       true,
		IndexOnSave:   true,
		IndexOnImport: "deferred",
	})
	ctx := context.Background()

	svc.HandleIndexEvent(ctx, IndexEvent{
		DocID: "doc-import", Content: "import content", Version: 1, Hash: "hash-imp", IsImport: true,
	})

	summary, err := repo.GetIndexStatusSummary(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatusSummary: %v", err)
	}
	// Should be pending, not indexed.
	if summary.Indexed != 0 {
		t.Errorf("Import deferred: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestService_Search_Disabled(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: false})
	ctx := context.Background()

	result, err := svc.Search(ctx, SearchQuery{Q: "anything"})
	if err != nil {
		t.Fatalf("Search disabled: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("Search disabled: got %d items, want 0", len(result.Items))
	}
}

func TestService_Reindex_All(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true, IndexOnSave: true})
	ctx := context.Background()

	// Insert some status records.
	for i := 1; i <= 3; i++ {
		_ = repo.UpsertIndexStatus(ctx, "doc-ri"+string(rune('0'+i)), "h"+string(rune('0'+i)), 1, IndexStatusIndexed, "")
	}

	if err := svc.Reindex(ctx, ReindexRequest{Scope: "all"}); err != nil {
		t.Fatalf("Reindex all: %v", err)
	}

	summary, err := svc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if summary.Indexed != 0 {
		t.Errorf("After reindex all: Indexed = %d, want 0", summary.Indexed)
	}
}

func TestService_Reindex_EmptyScopeDefaultsToAll(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true})
	ctx := context.Background()

	// Empty scope should default to all.
	if err := svc.Reindex(ctx, ReindexRequest{Scope: ""}); err != nil {
		t.Fatalf("Reindex empty scope: %v", err)
	}
}

func TestService_Reindex_DocumentMissingID(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true})
	ctx := context.Background()

	err := svc.Reindex(ctx, ReindexRequest{Scope: "document", DocumentID: ""})
	if err == nil {
		t.Error("Reindex document without ID: expected error, got nil")
	}
}

func TestService_GetHistory_ClearHistory(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true, IndexOnSave: true, HistoryLimit: 50})
	ctx := context.Background()

	// Create a document and search for it to generate history.
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	docSvc.SetIndexHook(func(ctx context.Context, ev document.IndexEvent) {
		svc.HandleIndexEvent(ctx, IndexEvent{
			DocID: ev.DocID, Content: ev.Content, Version: ev.Version, Hash: ev.Hash, Deleted: ev.Deleted,
		})
	})

	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "History Test",
		Content: "# History\n\nContent for history testing.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = svc.Search(ctx, SearchQuery{Q: "history testing", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	histResp, err := svc.GetHistory(ctx)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(histResp.Items) < 1 {
		t.Error("GetHistory: expected at least 1 item")
	}

	if err := svc.ClearHistory(ctx); err != nil {
		t.Fatalf("ClearHistory: %v", err)
	}

	histResp2, err := svc.GetHistory(ctx)
	if err != nil {
		t.Fatalf("GetHistory after clear: %v", err)
	}
	if len(histResp2.Items) != 0 {
		t.Errorf("After clear: got %d items, want 0", len(histResp2.Items))
	}
}

func TestService_GetHistory_EmptyQueryNotSaved(t *testing.T) {
	db := openTestDB(t)
	repo := NewRepository(db)
	engine := NewFTSEngine(db, 20)
	store := storage.NewLocalStorage(t.TempDir())
	svc := NewService(engine, repo, store, config.SearchConfig{Enabled: true, IndexOnSave: true})
	ctx := context.Background()

	// Empty query should not be saved to history.
	_, err := svc.Search(ctx, SearchQuery{Q: "", Limit: 10})
	if err != nil {
		t.Fatalf("Search empty: %v", err)
	}

	histResp, err := svc.GetHistory(ctx)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(histResp.Items) != 0 {
		t.Errorf("Empty query should not be saved: got %d items", len(histResp.Items))
	}
}

func TestService_GetIndexStatus_WithRealDocs(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	// Create docs and verify index status.
	for i := 0; i < 3; i++ {
		_, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   "Status Doc " + string(rune('A'+i)),
			Content: "Content for status test",
		})
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	status, err := searchSvc.GetIndexStatus(ctx)
	if err != nil {
		t.Fatalf("GetIndexStatus: %v", err)
	}
	if status.TotalDocuments != 3 {
		t.Errorf("TotalDocuments: got %d, want 3", status.TotalDocuments)
	}
	if status.Indexed != 3 {
		t.Errorf("Indexed: got %d, want 3", status.Indexed)
	}
}

func TestService_Search_SavesHistory(t *testing.T) {
	searchSvc, docSvc, _ := setupSearch(t)
	ctx := context.Background()

	_, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Unique Keyword Doc",
		Content: "# Unique\n\nContent with xyzunique keyword.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = searchSvc.Search(ctx, SearchQuery{Q: "xyzunique", Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	hist, err := searchSvc.GetHistory(ctx)
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(hist.Items) < 1 {
		t.Error("Search history: expected at least 1 item")
	}
	found := false
	for _, item := range hist.Items {
		if item.Query == "xyzunique" {
			found = true
		}
	}
	if !found {
		t.Error("Search history: 'xyzunique' query not found in history")
	}
}
