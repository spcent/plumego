package collection

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

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

func setupService(t *testing.T) (*Service, *document.Service) {
	t.Helper()
	db := openTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	docRepo := document.NewSQLiteRepository(db)
	store := storage.NewLocalStorage(t.TempDir())
	docSvc := document.NewService(docRepo, store)

	return svc, docSvc
}

// createTestDocument inserts a document and returns its ID.
func createTestDocument(t *testing.T, docSvc *document.Service, title string) string {
	t.Helper()
	result, err := docSvc.Create(context.Background(), document.CreateRequest{
		Title:   title,
		Content: "# " + title + "\n\nContent for " + title,
	})
	if err != nil {
		t.Fatalf("create document %q: %v", title, err)
	}
	return result.ID
}

// --- Create ---

func TestService_Create_valid(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, err := svc.Create(ctx, CreateCollectionRequest{Name: "My Collection"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.ID == "" {
		t.Error("ID should not be empty")
	}
	if c.Name != "My Collection" {
		t.Errorf("Name: want 'My Collection', got %q", c.Name)
	}
	if c.Type != "manual" {
		t.Errorf("Type: want 'manual', got %q", c.Type)
	}
	if c.Status != "active" {
		t.Errorf("Status: want 'active', got %q", c.Status)
	}
}

func TestService_Create_emptyName(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.Create(ctx, CreateCollectionRequest{Name: "   "})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestService_Create_customType(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, err := svc.Create(ctx, CreateCollectionRequest{Name: "Search Coll", Type: "search"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if c.Type != "search" {
		t.Errorf("Type: want 'search', got %q", c.Type)
	}
}

// --- Update ---

func TestService_Update_valid(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Old Name"})
	updated, err := svc.Update(ctx, c.ID, UpdateCollectionRequest{Name: "New Name", Description: "Desc"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("Name: want 'New Name', got %q", updated.Name)
	}
	if updated.Description != "Desc" {
		t.Errorf("Description: want 'Desc', got %q", updated.Description)
	}
}

func TestService_Update_emptyName(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Existing"})
	_, err := svc.Update(ctx, c.ID, UpdateCollectionRequest{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestService_Update_notFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.Update(ctx, "no-such-id", UpdateCollectionRequest{Name: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- Delete ---

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Temp"})
	if err := svc.Delete(ctx, c.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := svc.GetByID(ctx, c.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound after delete, got %v", err)
	}
}

func TestService_Delete_notFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	err := svc.Delete(ctx, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- GetByID ---

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Detail Test"})
	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Collection.ID != c.ID {
		t.Errorf("ID: want %q, got %q", c.ID, detail.Collection.ID)
	}
	if detail.Total != 0 {
		t.Errorf("Total: want 0, got %d", detail.Total)
	}
}

func TestService_GetByID_notFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.GetByID(ctx, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- List ---

func TestService_List(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	svc.Create(ctx, CreateCollectionRequest{Name: "A"})
	svc.Create(ctx, CreateCollectionRequest{Name: "B"})
	svc.Create(ctx, CreateCollectionRequest{Name: "C"})

	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 3 || len(result.Items) != 3 {
		t.Errorf("want 3 items, got total=%d items=%d", result.Total, len(result.Items))
	}
}

func TestService_List_empty(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("want 0, got %d", result.Total)
	}
}

// --- AddDocument / RemoveDocument ---

func TestService_AddDocument_emptyDocID(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Coll"})
	err := svc.AddDocument(ctx, c.ID, AddDocumentRequest{DocumentID: ""})
	if err == nil {
		t.Fatal("expected error for empty document_id")
	}
}

func TestService_AddDocument_andRemove(t *testing.T) {
	ctx := context.Background()
	svc, docSvc := setupService(t)

	docID := createTestDocument(t, docSvc, "Test Doc")

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "With Doc"})
	if err := svc.AddDocument(ctx, c.ID, AddDocumentRequest{DocumentID: docID, Note: "my note"}); err != nil {
		t.Fatalf("AddDocument: %v", err)
	}

	// Verify document appears in detail
	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Total != 1 {
		t.Errorf("Total: want 1, got %d", detail.Total)
	}
	if len(detail.Documents) != 1 {
		t.Fatalf("Documents: want 1, got %d", len(detail.Documents))
	}
	if detail.Documents[0].DocumentID != docID {
		t.Errorf("DocumentID: want %q, got %q", docID, detail.Documents[0].DocumentID)
	}

	// Remove
	if err := svc.RemoveDocument(ctx, c.ID, docID); err != nil {
		t.Fatalf("RemoveDocument: %v", err)
	}

	detail2, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID after remove: %v", err)
	}
	if detail2.Total != 0 {
		t.Errorf("Total after remove: want 0, got %d", detail2.Total)
	}
}

// --- Reorder ---

func TestService_Reorder(t *testing.T) {
	ctx := context.Background()
	svc, docSvc := setupService(t)

	doc1 := createTestDocument(t, docSvc, "Doc 1")
	doc2 := createTestDocument(t, docSvc, "Doc 2")
	doc3 := createTestDocument(t, docSvc, "Doc 3")

	c, _ := svc.Create(ctx, CreateCollectionRequest{Name: "Reorder Test"})
	svc.AddDocument(ctx, c.ID, AddDocumentRequest{DocumentID: doc1})
	svc.AddDocument(ctx, c.ID, AddDocumentRequest{DocumentID: doc2})
	svc.AddDocument(ctx, c.ID, AddDocumentRequest{DocumentID: doc3})

	// Reorder: doc3, doc1, doc2
	err := svc.Reorder(ctx, c.ID, ReorderRequest{DocumentIDs: []string{doc3, doc1, doc2}})
	if err != nil {
		t.Fatalf("Reorder: %v", err)
	}

	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(detail.Documents) != 3 {
		t.Fatalf("want 3 docs after reorder, got %d", len(detail.Documents))
	}
	// After reorder by sort_order, doc3 should be first (sort_order=1)
	if detail.Documents[0].DocumentID != doc3 {
		t.Errorf("first doc: want %q, got %q", doc3, detail.Documents[0].DocumentID)
	}
}

// --- CreateFromSearch ---

func TestService_CreateFromSearch_valid(t *testing.T) {
	ctx := context.Background()
	svc, docSvc := setupService(t)

	doc1 := createTestDocument(t, docSvc, "Search Doc 1")
	doc2 := createTestDocument(t, docSvc, "Search Doc 2")

	c, err := svc.CreateFromSearch(ctx, CreateFromSearchRequest{
		Name:        "Search Collection",
		Description: "from search",
		DocumentIDs: []string{doc1, doc2},
	})
	if err != nil {
		t.Fatalf("CreateFromSearch: %v", err)
	}
	if c.Name != "Search Collection" {
		t.Errorf("Name: want 'Search Collection', got %q", c.Name)
	}
	if c.Type != "search" {
		t.Errorf("Type: want 'search', got %q", c.Type)
	}

	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Total != 2 {
		t.Errorf("Total: want 2, got %d", detail.Total)
	}
}

func TestService_CreateFromSearch_emptyName(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.CreateFromSearch(ctx, CreateFromSearchRequest{Name: ""})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestService_CreateFromSearch_noDocuments(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	c, err := svc.CreateFromSearch(ctx, CreateFromSearchRequest{
		Name:        "Empty Search",
		DocumentIDs: []string{},
	})
	if err != nil {
		t.Fatalf("CreateFromSearch: %v", err)
	}

	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Total != 0 {
		t.Errorf("Total: want 0, got %d", detail.Total)
	}
}

// --- timestamps sanity ---

func TestService_Create_timestamps(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	before := time.Now().Add(-time.Second)
	c, err := svc.Create(ctx, CreateCollectionRequest{Name: "TS Test"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	after := time.Now().Add(time.Second)

	detail, err := svc.GetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if detail.Collection.CreatedAt.Before(before) || detail.Collection.CreatedAt.After(after) {
		t.Errorf("CreatedAt out of range: %v", detail.Collection.CreatedAt)
	}
}
