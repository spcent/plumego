package document

import (
	"context"
	"testing"
	"time"
)

// openTestDB is already defined in service_test.go in this package.
// We reuse it here since we're in the same package.

func newTestDoc(id, title, hash string) *Document {
	now := time.Now().UTC()
	return &Document{
		ID:           id,
		Title:        title,
		Slug:         title,
		StorageKey:   "docs/" + id + "/current.md",
		ContentHash:  hash,
		Status:       StatusActive,
		SyncStatus:   SyncStatusSynced,
		SourceType:   SourceTypeManual,
		ReviewStatus: ReviewStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func newTestVersion(id, docID string, version int) *DocumentVersion {
	return &DocumentVersion{
		ID:          id,
		DocumentID:  docID,
		Version:     version,
		StorageKey:  VersionKey(docID, version),
		ContentHash: "hash-v" + string(rune('0'+version)),
		SizeBytes:   100,
		CreatedAt:   time.Now().UTC(),
		Note:        "",
	}
}

// TestSQLiteRepository_CreateAndGetByID verifies Create stores a document and GetByID retrieves it.
func TestSQLiteRepository_CreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-001", "Hello World", "hash-abc")
	doc.SizeBytes = 512
	doc.WordCount = 50
	doc.LineCount = 5
	doc.Summary = "A test summary"
	doc.HeadingText = "Hello World"

	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != doc.ID {
		t.Errorf("ID mismatch: got %q, want %q", got.ID, doc.ID)
	}
	if got.Title != doc.Title {
		t.Errorf("Title mismatch: got %q, want %q", got.Title, doc.Title)
	}
	if got.ContentHash != doc.ContentHash {
		t.Errorf("ContentHash mismatch: got %q, want %q", got.ContentHash, doc.ContentHash)
	}
	if got.Status != StatusActive {
		t.Errorf("Status mismatch: got %q, want %q", got.Status, StatusActive)
	}
	if got.Summary != doc.Summary {
		t.Errorf("Summary mismatch: got %q, want %q", got.Summary, doc.Summary)
	}
	if got.HeadingText != doc.HeadingText {
		t.Errorf("HeadingText mismatch: got %q, want %q", got.HeadingText, doc.HeadingText)
	}
}

// TestSQLiteRepository_GetByID_NotFound verifies ErrNotFound for missing document.
func TestSQLiteRepository_GetByID_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("GetByID nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_GetByID_DeletedDoc verifies deleted doc is not found.
func TestSQLiteRepository_GetByID_DeletedDoc(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-del", "To Delete", "hash-del")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.SoftDelete(ctx, doc.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	_, err := repo.GetByID(ctx, doc.ID)
	if err != ErrNotFound {
		t.Errorf("GetByID deleted: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_Update verifies that Update modifies doc fields.
func TestSQLiteRepository_Update(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-upd", "Original Title", "hash-orig")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Modify and update.
	doc.Title = "Updated Title"
	doc.StorageKey = "docs/doc-upd/current-v2.md"
	doc.ContentHash = "hash-updated"
	doc.SizeBytes = 1024
	doc.WordCount = 200
	doc.LineCount = 20
	doc.CurrentVersion = 2
	doc.Summary = "Updated summary"
	doc.HeadingText = "Updated Title"
	now := time.Now().UTC()
	doc.UpdatedAt = now
	uploaded := now
	doc.UploadedAt = &uploaded

	if err := repo.Update(ctx, doc); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Title != "Updated Title" {
		t.Errorf("Title: got %q, want %q", got.Title, "Updated Title")
	}
	if got.ContentHash != "hash-updated" {
		t.Errorf("ContentHash: got %q, want %q", got.ContentHash, "hash-updated")
	}
	if got.CurrentVersion != 2 {
		t.Errorf("CurrentVersion: got %d, want 2", got.CurrentVersion)
	}
	if got.Summary != "Updated summary" {
		t.Errorf("Summary: got %q, want %q", got.Summary, "Updated summary")
	}
}

// TestSQLiteRepository_SoftDelete verifies SoftDelete hides document from normal queries.
func TestSQLiteRepository_SoftDelete(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-soft", "Soft Delete Me", "hash-soft")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Verify exists.
	if _, err := repo.GetByID(ctx, doc.ID); err != nil {
		t.Fatalf("GetByID before delete: %v", err)
	}

	// Delete.
	if err := repo.SoftDelete(ctx, doc.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// Verify no longer found.
	if _, err := repo.GetByID(ctx, doc.ID); err != ErrNotFound {
		t.Errorf("After SoftDelete: got %v, want ErrNotFound", err)
	}

	// Verify no longer in List.
	docs, _, err := repo.List(ctx, ListQuery{Status: StatusActive})
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	for _, d := range docs {
		if d.ID == doc.ID {
			t.Errorf("Deleted document still appears in List")
		}
	}
}

// TestSQLiteRepository_SoftDelete_AlreadyDeleted verifies double-delete returns ErrNotFound.
func TestSQLiteRepository_SoftDelete_AlreadyDeleted(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-dd", "Double Delete", "hash-dd")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := repo.SoftDelete(ctx, doc.ID); err != nil {
		t.Fatalf("First SoftDelete: %v", err)
	}
	err := repo.SoftDelete(ctx, doc.ID)
	if err != ErrNotFound {
		t.Errorf("Second SoftDelete: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_SoftDelete_NonExistent verifies ErrNotFound for missing document.
func TestSQLiteRepository_SoftDelete_NonExistent(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.SoftDelete(ctx, "nope")
	if err != ErrNotFound {
		t.Errorf("SoftDelete nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_ExistsByHash verifies hash lookup.
func TestSQLiteRepository_ExistsByHash(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Not yet exists.
	exists, err := repo.ExistsByHash(ctx, "unique-hash-xyz")
	if err != nil {
		t.Fatalf("ExistsByHash (absent): %v", err)
	}
	if exists {
		t.Error("ExistsByHash = true for absent hash")
	}

	// Create doc.
	doc := newTestDoc("doc-hash", "Hash Test", "unique-hash-xyz")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	exists, err = repo.ExistsByHash(ctx, "unique-hash-xyz")
	if err != nil {
		t.Fatalf("ExistsByHash (present): %v", err)
	}
	if !exists {
		t.Error("ExistsByHash = false for present hash")
	}

	// After soft delete should return false.
	if err := repo.SoftDelete(ctx, doc.ID); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}
	exists, err = repo.ExistsByHash(ctx, "unique-hash-xyz")
	if err != nil {
		t.Fatalf("ExistsByHash (deleted): %v", err)
	}
	if exists {
		t.Error("ExistsByHash = true after soft delete")
	}
}

// TestSQLiteRepository_UpdateFavorite verifies toggling favorite.
func TestSQLiteRepository_UpdateFavorite(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-fav", "Fav Test", "hash-fav")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Default should be false.
	got, _ := repo.GetByID(ctx, doc.ID)
	if got.IsFavorite {
		t.Error("Initial IsFavorite should be false")
	}

	// Set to true.
	if err := repo.UpdateFavorite(ctx, doc.ID, true); err != nil {
		t.Fatalf("UpdateFavorite(true): %v", err)
	}
	got, _ = repo.GetByID(ctx, doc.ID)
	if !got.IsFavorite {
		t.Error("IsFavorite should be true after UpdateFavorite(true)")
	}

	// Set back to false.
	if err := repo.UpdateFavorite(ctx, doc.ID, false); err != nil {
		t.Fatalf("UpdateFavorite(false): %v", err)
	}
	got, _ = repo.GetByID(ctx, doc.ID)
	if got.IsFavorite {
		t.Error("IsFavorite should be false after UpdateFavorite(false)")
	}
}

// TestSQLiteRepository_UpdateFavorite_NotFound verifies ErrNotFound for missing document.
func TestSQLiteRepository_UpdateFavorite_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.UpdateFavorite(ctx, "nope", true)
	if err != ErrNotFound {
		t.Errorf("UpdateFavorite nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_UpdateStatus verifies status change.
func TestSQLiteRepository_UpdateStatus(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-status", "Status Test", "hash-st")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, doc.ID, StatusArchived); err != nil {
		t.Fatalf("UpdateStatus(archived): %v", err)
	}

	// GetByID uses status != 'deleted' filter — archived doc is still retrievable by ID.
	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID archived doc: %v", err)
	}
	if got.Status != StatusArchived {
		t.Errorf("Status after archive: got %q, want %q", got.Status, StatusArchived)
	}

	// Should NOT appear in active List (default).
	activeDocs, activeTotal, err := repo.List(ctx, ListQuery{Status: StatusActive})
	if err != nil {
		t.Fatalf("List active after archive: %v", err)
	}
	for _, d := range activeDocs {
		if d.ID == doc.ID {
			t.Errorf("Archived doc should not appear in active list")
		}
	}
	_ = activeTotal

	// Should be findable when filtering archived.
	docs, total, err := repo.List(ctx, ListQuery{Status: StatusArchived})
	if err != nil {
		t.Fatalf("List archived: %v", err)
	}
	if total != 1 {
		t.Errorf("List archived total: got %d, want 1", total)
	}
	if len(docs) == 0 || docs[0].ID != doc.ID {
		t.Error("Archived doc not found in List with archived status")
	}

	// Restore back to active.
	if err := repo.UpdateStatus(ctx, doc.ID, StatusActive); err != nil {
		t.Fatalf("UpdateStatus(active): %v", err)
	}
	got, err = repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID after restore: %v", err)
	}
	if got.Status != StatusActive {
		t.Errorf("Status after restore: got %q, want %q", got.Status, StatusActive)
	}
}

// TestSQLiteRepository_UpdateStatus_NotFound verifies ErrNotFound behavior.
func TestSQLiteRepository_UpdateStatus_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.UpdateStatus(ctx, "nope", StatusArchived)
	if err != ErrNotFound {
		t.Errorf("UpdateStatus nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_BatchUpdateStatus verifies batch status changes.
func TestSQLiteRepository_BatchUpdateStatus(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	ids := []string{"doc-b1", "doc-b2", "doc-b3"}
	for i, id := range ids {
		d := newTestDoc(id, "Batch "+string(rune('A'+i)), "hash-b"+string(rune('1'+i)))
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create %s: %v", id, err)
		}
	}

	// Batch archive first two.
	if err := repo.BatchUpdateStatus(ctx, ids[:2], StatusArchived); err != nil {
		t.Fatalf("BatchUpdateStatus: %v", err)
	}

	archived, _, err := repo.List(ctx, ListQuery{Status: StatusArchived})
	if err != nil {
		t.Fatalf("List archived: %v", err)
	}
	if len(archived) != 2 {
		t.Errorf("Archived count: got %d, want 2", len(archived))
	}

	// Third should still be active.
	active, _, err := repo.List(ctx, ListQuery{Status: StatusActive})
	if err != nil {
		t.Fatalf("List active: %v", err)
	}
	if len(active) != 1 || active[0].ID != "doc-b3" {
		t.Errorf("Active after batch: expected doc-b3 only, got %v", active)
	}
}

// TestSQLiteRepository_BatchUpdateStatus_Empty verifies no error for empty batch.
func TestSQLiteRepository_BatchUpdateStatus_Empty(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	if err := repo.BatchUpdateStatus(ctx, []string{}, StatusArchived); err != nil {
		t.Errorf("BatchUpdateStatus empty: %v", err)
	}
}

// TestSQLiteRepository_UpdateReviewStatus verifies review status update.
func TestSQLiteRepository_UpdateReviewStatus(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-rev", "Review Test", "hash-rev")
	doc.ReviewStatus = ReviewStatusPending
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateReviewStatus(ctx, doc.ID, ReviewStatusReviewed); err != nil {
		t.Fatalf("UpdateReviewStatus: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ReviewStatus != ReviewStatusReviewed {
		t.Errorf("ReviewStatus: got %q, want %q", got.ReviewStatus, ReviewStatusReviewed)
	}
}

// TestSQLiteRepository_UpdateReviewStatus_NotFound verifies ErrNotFound behavior.
func TestSQLiteRepository_UpdateReviewStatus_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	err := repo.UpdateReviewStatus(ctx, "nope", ReviewStatusReviewed)
	if err != ErrNotFound {
		t.Errorf("UpdateReviewStatus nonexistent: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_CreateVersion_GetVersions_GetVersion verifies version CRUD.
func TestSQLiteRepository_CreateVersion_GetVersions_GetVersion(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-ver", "Version Test", "hash-v1")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create doc: %v", err)
	}

	// Create versions 1, 2, 3.
	for v := 1; v <= 3; v++ {
		ver := newTestVersion("ver-id-"+string(rune('0'+v)), doc.ID, v)
		if err := repo.CreateVersion(ctx, ver); err != nil {
			t.Fatalf("CreateVersion %d: %v", v, err)
		}
	}

	// GetVersions returns all in descending order.
	versions, err := repo.GetVersions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("GetVersions count: got %d, want 3", len(versions))
	}
	if versions[0].Version != 3 {
		t.Errorf("First version: got %d, want 3 (descending)", versions[0].Version)
	}
	if versions[2].Version != 1 {
		t.Errorf("Last version: got %d, want 1", versions[2].Version)
	}

	// GetVersion for a specific version.
	ver2, err := repo.GetVersion(ctx, doc.ID, 2)
	if err != nil {
		t.Fatalf("GetVersion(2): %v", err)
	}
	if ver2.Version != 2 {
		t.Errorf("GetVersion(2): got %d, want 2", ver2.Version)
	}
	if ver2.DocumentID != doc.ID {
		t.Errorf("GetVersion DocumentID: got %q, want %q", ver2.DocumentID, doc.ID)
	}
}

// TestSQLiteRepository_GetVersion_NotFound verifies ErrNotFound for missing version.
func TestSQLiteRepository_GetVersion_NotFound(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-vnf", "No Versions", "hash-vnf")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err := repo.GetVersion(ctx, doc.ID, 99)
	if err != ErrNotFound {
		t.Errorf("GetVersion missing: got %v, want ErrNotFound", err)
	}
}

// TestSQLiteRepository_GetVersions_Empty verifies empty list for doc with no versions.
func TestSQLiteRepository_GetVersions_Empty(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	doc := newTestDoc("doc-ve", "Empty Versions", "hash-ve")
	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create: %v", err)
	}

	versions, err := repo.GetVersions(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetVersions empty: %v", err)
	}
	if len(versions) != 0 {
		t.Errorf("GetVersions empty count: got %d, want 0", len(versions))
	}
}

// TestSQLiteRepository_List_StatusFilter verifies status filtering in List.
func TestSQLiteRepository_List_StatusFilter(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	active := newTestDoc("doc-la", "Active Doc", "hash-la")
	active.Status = StatusActive
	archived := newTestDoc("doc-arch", "Archived Doc", "hash-arch")
	archived.Status = StatusArchived

	for _, d := range []*Document{active, archived} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create %s: %v", d.ID, err)
		}
	}

	// Default (active only).
	docs, total, err := repo.List(ctx, ListQuery{})
	if err != nil {
		t.Fatalf("List default: %v", err)
	}
	if total != 1 || len(docs) != 1 || docs[0].ID != "doc-la" {
		t.Errorf("List default: got %d docs (total=%d), want 1 active", len(docs), total)
	}

	// Archived.
	docs, total, err = repo.List(ctx, ListQuery{Status: StatusArchived})
	if err != nil {
		t.Fatalf("List archived: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-arch" {
		t.Errorf("List archived: got %d docs, want 1", total)
	}

	// All (active + archived, not deleted).
	docs, total, err = repo.List(ctx, ListQuery{Status: "all"})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if total != 2 {
		t.Errorf("List all total: got %d, want 2", total)
	}
	_ = docs
}

// TestSQLiteRepository_List_SourceTypeFilter verifies source_type filtering.
func TestSQLiteRepository_List_SourceTypeFilter(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	manual := newTestDoc("doc-man", "Manual", "hash-man")
	manual.SourceType = SourceTypeManual
	imported := newTestDoc("doc-imp", "Imported", "hash-imp")
	imported.SourceType = SourceTypeImported

	for _, d := range []*Document{manual, imported} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create %s: %v", d.ID, err)
		}
	}

	docs, total, err := repo.List(ctx, ListQuery{SourceType: SourceTypeImported})
	if err != nil {
		t.Fatalf("List imported: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-imp" {
		t.Errorf("List source_type=imported: got %d, want 1", total)
	}

	docs, total, err = repo.List(ctx, ListQuery{SourceType: SourceTypeManual})
	if err != nil {
		t.Fatalf("List manual: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-man" {
		t.Errorf("List source_type=manual: got %d, want 1", total)
	}
}

// TestSQLiteRepository_List_IsFavoriteFilter verifies IsFavorite filtering.
func TestSQLiteRepository_List_IsFavoriteFilter(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	favDoc := newTestDoc("doc-fav1", "Fav One", "hash-fav1")
	favDoc.IsFavorite = true
	normalDoc := newTestDoc("doc-norm", "Normal One", "hash-norm1")
	normalDoc.IsFavorite = false

	for _, d := range []*Document{favDoc, normalDoc} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	favTrue := true
	docs, total, err := repo.List(ctx, ListQuery{IsFavorite: &favTrue})
	if err != nil {
		t.Fatalf("List is_favorite=true: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-fav1" {
		t.Errorf("List is_favorite=true: got %d docs, want 1", total)
	}

	favFalse := false
	docs, total, err = repo.List(ctx, ListQuery{IsFavorite: &favFalse})
	if err != nil {
		t.Fatalf("List is_favorite=false: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-norm" {
		t.Errorf("List is_favorite=false: got %d docs, want 1", total)
	}
}

// TestSQLiteRepository_List_SearchQuery verifies title search (q param).
func TestSQLiteRepository_List_SearchQuery(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	docs := []*Document{
		newTestDoc("doc-q1", "Go Programming Guide", "hash-q1"),
		newTestDoc("doc-q2", "Python Tutorial", "hash-q2"),
		newTestDoc("doc-q3", "Go Web Development", "hash-q3"),
	}
	for _, d := range docs {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// Search for "Go" should return 2 results.
	results, total, err := repo.List(ctx, ListQuery{Q: "Go"})
	if err != nil {
		t.Fatalf("List Q=Go: %v", err)
	}
	if total != 2 {
		t.Errorf("List Q=Go total: got %d, want 2", total)
	}

	// Search for "Python" should return 1 result.
	results, total, err = repo.List(ctx, ListQuery{Q: "Python"})
	if err != nil {
		t.Fatalf("List Q=Python: %v", err)
	}
	if total != 1 {
		t.Errorf("List Q=Python total: got %d, want 1", total)
	}
	_ = results
}

// TestSQLiteRepository_List_ReviewStatusFilter verifies review_status filtering.
func TestSQLiteRepository_List_ReviewStatusFilter(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	pending := newTestDoc("doc-rp", "Review Pending", "hash-rp")
	pending.ReviewStatus = ReviewStatusPending
	reviewed := newTestDoc("doc-rr", "Review Reviewed", "hash-rr")
	reviewed.ReviewStatus = ReviewStatusReviewed

	for _, d := range []*Document{pending, reviewed} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	docs, total, err := repo.List(ctx, ListQuery{ReviewStatus: ReviewStatusPending})
	if err != nil {
		t.Fatalf("List review_status=pending: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-rp" {
		t.Errorf("List review_status=pending: got %d docs", total)
	}

	docs, total, err = repo.List(ctx, ListQuery{ReviewStatus: ReviewStatusReviewed})
	if err != nil {
		t.Fatalf("List review_status=reviewed: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-rr" {
		t.Errorf("List review_status=reviewed: got %d docs", total)
	}
}

// TestSQLiteRepository_List_SortOrders verifies all sort orders.
func TestSQLiteRepository_List_SortOrders(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	// Create docs with known names and sizes.
	docA := newTestDoc("sort-a", "Alpha Doc", "hash-sa")
	docA.SizeBytes = 300
	docB := newTestDoc("sort-b", "Beta Doc", "hash-sb")
	docB.SizeBytes = 100
	docC := newTestDoc("sort-c", "Charlie Doc", "hash-sc")
	docC.SizeBytes = 200

	for _, d := range []*Document{docA, docB, docC} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// Sort by title ASC.
	docs, _, err := repo.List(ctx, ListQuery{SortBy: "title", Order: "asc"})
	if err != nil {
		t.Fatalf("List title asc: %v", err)
	}
	if len(docs) != 3 || docs[0].Title != "Alpha Doc" {
		t.Errorf("Sort title asc first: got %q, want %q", docs[0].Title, "Alpha Doc")
	}

	// Sort by title DESC.
	docs, _, err = repo.List(ctx, ListQuery{SortBy: "title", Order: "desc"})
	if err != nil {
		t.Fatalf("List title desc: %v", err)
	}
	if len(docs) != 3 || docs[0].Title != "Charlie Doc" {
		t.Errorf("Sort title desc first: got %q, want %q", docs[0].Title, "Charlie Doc")
	}

	// Sort by size_bytes ASC.
	docs, _, err = repo.List(ctx, ListQuery{SortBy: "size_bytes", Order: "asc"})
	if err != nil {
		t.Fatalf("List size_bytes asc: %v", err)
	}
	if len(docs) != 3 || docs[0].SizeBytes != 100 {
		t.Errorf("Sort size_bytes asc first: got %d, want 100", docs[0].SizeBytes)
	}

	// Sort by created_at DESC (default dir).
	docs, _, err = repo.List(ctx, ListQuery{SortBy: "created_at"})
	if err != nil {
		t.Fatalf("List created_at: %v", err)
	}
	if len(docs) != 3 {
		t.Errorf("Sort created_at count: got %d, want 3", len(docs))
	}
}

// TestSQLiteRepository_List_Pagination verifies limit/offset.
func TestSQLiteRepository_List_Pagination(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		d := newTestDoc("page-"+string(rune('a'+i)), "Page Doc "+string(rune('A'+i)), "hash-pg"+string(rune('a'+i)))
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// First page.
	docs, total, err := repo.List(ctx, ListQuery{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List page 1: %v", err)
	}
	if total != 5 {
		t.Errorf("Total: got %d, want 5", total)
	}
	if len(docs) != 2 {
		t.Errorf("Page 1 count: got %d, want 2", len(docs))
	}

	// Last page.
	docs, _, err = repo.List(ctx, ListQuery{Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("List page 3: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("Page 3 count: got %d, want 1", len(docs))
	}
}

// TestSQLiteRepository_List_ImportJobID verifies import_job_id filtering.
func TestSQLiteRepository_List_ImportJobID(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	withJob := newTestDoc("doc-job1", "Job Doc", "hash-job1")
	withJob.ImportJobID = "job-42"
	noJob := newTestDoc("doc-nojob", "No Job Doc", "hash-nojob")

	for _, d := range []*Document{withJob, noJob} {
		if err := repo.Create(ctx, d); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	docs, total, err := repo.List(ctx, ListQuery{ImportJobID: "job-42"})
	if err != nil {
		t.Fatalf("List import_job_id: %v", err)
	}
	if total != 1 || docs[0].ID != "doc-job1" {
		t.Errorf("List import_job_id: got %d docs", total)
	}
}

// TestSQLiteRepository_CreateDoc_WithNullableFields verifies nullable fields are handled.
func TestSQLiteRepository_CreateDoc_WithNullableFields(t *testing.T) {
	db := openTestDB(t)
	repo := NewSQLiteRepository(db)
	ctx := context.Background()

	now := time.Now().UTC()
	importedAt := now
	uploadedAt := now

	doc := &Document{
		ID:           "doc-null",
		Title:        "Nullable Fields",
		OriginalPath: "/path/to/file.md",
		StorageKey:   "docs/doc-null/current.md",
		ContentHash:  "hash-null",
		Status:       StatusActive,
		SyncStatus:   SyncStatusSynced,
		SourceType:   SourceTypeImported,
		ImportJobID:  "import-job-1",
		ReviewStatus: ReviewStatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		UploadedAt:   &uploadedAt,
		ImportedAt:   &importedAt,
		Summary:      "A summary",
		HeadingText:  "Nullable Fields",
	}

	if err := repo.Create(ctx, doc); err != nil {
		t.Fatalf("Create with nullable fields: %v", err)
	}

	got, err := repo.GetByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.OriginalPath != "/path/to/file.md" {
		t.Errorf("OriginalPath: got %q, want %q", got.OriginalPath, "/path/to/file.md")
	}
	if got.ImportJobID != "import-job-1" {
		t.Errorf("ImportJobID: got %q, want %q", got.ImportJobID, "import-job-1")
	}
	if got.UploadedAt == nil {
		t.Error("UploadedAt should not be nil")
	}
	if got.ImportedAt == nil {
		t.Error("ImportedAt should not be nil")
	}
}
