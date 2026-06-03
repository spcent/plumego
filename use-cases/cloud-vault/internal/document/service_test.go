package document

import (
	"context"
	"path/filepath"
	"testing"

	"cloud-vault/internal/database"
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

func setupService(t *testing.T) (*Service, *database.DB, storage.ObjectStorage) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewSQLiteRepository(db)
	svc := NewService(repo, store)
	return svc, db, store
}

func setupServiceWithVersioning(t *testing.T, cfg VersioningConfig) (*Service, *database.DB, storage.ObjectStorage) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	repo := NewSQLiteRepository(db)
	svc := NewServiceWithVersioning(repo, store, cfg)
	return svc, db, store
}

func TestService_Create_Basic(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	req := CreateRequest{
		Title:   "Test Document",
		Content: "# Hello\n\nThis is a test document.",
	}

	result, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if result.ID == "" {
		t.Error("ID is empty")
	}
	if result.Title != "Test Document" {
		t.Errorf("Title = %q, want %q", result.Title, "Test Document")
	}
	if result.Version != 1 {
		t.Errorf("Version = %d, want 1", result.Version)
	}
	if !result.Changed {
		t.Error("Changed = false, want true")
	}
}

func TestService_Create_EmptyTitle(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	// Should extract title from content
	req := CreateRequest{
		Title:   "",
		Content: "# Auto Title\n\nContent here.",
	}

	result, err := svc.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if result.Title != "Auto Title" {
		t.Errorf("Title = %q, want %q", result.Title, "Auto Title")
	}
}

func TestService_Get_ByID(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	content := "# Test\n\nContent for retrieval."
	created, err := svc.Create(ctx, CreateRequest{Title: "Get Test", Content: content})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	detail, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if detail.ID != created.ID {
		t.Errorf("ID = %q, want %q", detail.ID, created.ID)
	}
	if detail.Content != content {
		t.Errorf("Content = %q, want %q", detail.Content, content)
	}
	if detail.Version != 1 {
		t.Errorf("Version = %d, want 1", detail.Version)
	}
}

func TestService_Update_Success(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Title:   "Update Test",
		Content: "v1 content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Update(ctx, created.ID, UpdateRequest{
		Title:       "Update Test v2",
		Content:     "v2 content updated",
		BaseVersion: 1,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Version != 2 {
		t.Errorf("Version = %d, want 2", updated.Version)
	}
	if !updated.Changed {
		t.Error("Changed = false, want true")
	}

	// Verify content
	detail, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get after update: %v", err)
	}
	if detail.Content != "v2 content updated" {
		t.Errorf("Content = %q, want %q", detail.Content, "v2 content updated")
	}
}

func TestService_Update_VersionConflict(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Title:   "Conflict Test",
		Content: "original",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Try to update with wrong base version
	_, err = svc.Update(ctx, created.ID, UpdateRequest{
		Title:       "New Title",
		Content:     "new content",
		BaseVersion: 99, // Wrong version
	})
	if err != ErrVersionConflict {
		t.Errorf("Update with wrong version: got %v, want ErrVersionConflict", err)
	}
}

func TestService_Update_NoChange(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Title:   "No Change Test",
		Content: "same content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update with identical content and title
	updated, err := svc.Update(ctx, created.ID, UpdateRequest{
		Title:       "No Change Test",
		Content:     "same content",
		BaseVersion: 1,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if updated.Changed {
		t.Error("Changed = true, want false for identical content")
	}
	if updated.Version != 1 {
		t.Errorf("Version = %d, want 1 (no increment)", updated.Version)
	}
}

func TestService_Delete_SoftDelete(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Title:   "Delete Test",
		Content: "will be deleted",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = svc.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Should not be retrievable after delete
	_, err = svc.Get(ctx, created.ID)
	if err != ErrNotFound {
		t.Errorf("Get after delete: got %v, want ErrNotFound", err)
	}
}

func TestService_List_Pagination(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	// Create 5 documents
	for i := 0; i < 5; i++ {
		_, err := svc.Create(ctx, CreateRequest{
			Title:   "Doc " + string(rune('A'+i)),
			Content: "Content " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
	}

	// List with limit
	result, err := svc.List(ctx, ListQuery{Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items count = %d, want 2", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("Total = %d, want 5", result.Total)
	}

	// List with offset
	result2, err := svc.List(ctx, ListQuery{Limit: 10, Offset: 3})
	if err != nil {
		t.Fatalf("List with offset: %v", err)
	}
	if len(result2.Items) != 2 {
		t.Errorf("Items count with offset = %d, want 2", len(result2.Items))
	}
}

func TestService_ExistsByHash(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	content := "# Unique Content\n\nFor hash testing."
	created, err := svc.Create(ctx, CreateRequest{Title: "Hash Test", Content: content})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Compute hash
	hash := HashContent(content)

	exists, err := svc.ExistsByHash(ctx, hash)
	if err != nil {
		t.Fatalf("ExistsByHash: %v", err)
	}
	if !exists {
		t.Error("ExistsByHash = false, want true")
	}

	// Non-existent hash
	exists2, err := svc.ExistsByHash(ctx, "nonexistenthash123")
	if err != nil {
		t.Fatalf("ExistsByHash nonexistent: %v", err)
	}
	if exists2 {
		t.Error("ExistsByHash = true for nonexistent hash")
	}

	// Delete and verify hash no longer exists
	err = svc.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	exists3, err := svc.ExistsByHash(ctx, hash)
	if err != nil {
		t.Fatalf("ExistsByHash after delete: %v", err)
	}
	if exists3 {
		t.Error("ExistsByHash = true after delete")
	}
}

func TestService_GetVersions(t *testing.T) {
	svc, _, _ := setupService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{
		Title:   "Version Test",
		Content: "v1",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Update twice
	_, err = svc.Update(ctx, created.ID, UpdateRequest{
		Title:       "Version Test",
		Content:     "v2",
		BaseVersion: 1,
	})
	if err != nil {
		t.Fatalf("Update v2: %v", err)
	}

	_, err = svc.Update(ctx, created.ID, UpdateRequest{
		Title:       "Version Test",
		Content:     "v3",
		BaseVersion: 2,
	})
	if err != nil {
		t.Fatalf("Update v3: %v", err)
	}

	// Get versions
	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if len(versions.Items) != 3 {
		t.Errorf("Version count = %d, want 3", len(versions.Items))
	}

	// Versions should be in descending order
	if versions.Items[0].Version != 3 {
		t.Errorf("First version = %d, want 3", versions.Items[0].Version)
	}
}

func TestService_BoundedVersionRetention_PrunesOldAutoSnapshots(t *testing.T) {
	svc, _, store := setupServiceWithVersioning(t, VersioningConfig{Policy: VersionPolicyBounded, KeepLatest: 2})
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{Title: "Bounded", Content: "v1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	base := 1
	for version := 2; version <= 5; version++ {
		updated, err := svc.Update(ctx, created.ID, UpdateRequest{
			Title:       "Bounded",
			Content:     "v" + string(rune('0'+version)),
			BaseVersion: base,
		})
		if err != nil {
			t.Fatalf("Update v%d: %v", version, err)
		}
		base = updated.Version
	}

	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if len(versions.Items) != 2 {
		t.Fatalf("Version count = %d, want 2", len(versions.Items))
	}
	if versions.Items[0].Version != 5 || versions.Items[1].Version != 4 {
		t.Fatalf("Kept versions = %#v, want v5 and v4", versions.Items)
	}
	if exists, err := store.Exists(ctx, VersionKey(created.ID, 3)); err != nil || exists {
		t.Fatalf("Version 3 object exists=%v err=%v, want pruned", exists, err)
	}
}

func TestService_CreateSnapshot_PinsCurrentVersion(t *testing.T) {
	svc, _, _ := setupServiceWithVersioning(t, VersioningConfig{Policy: VersionPolicyBounded, KeepLatest: 1})
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{Title: "Pinned", Content: "v1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	snapshot, err := svc.CreateSnapshot(ctx, created.ID, "before large edit")
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if !snapshot.Pinned || snapshot.Kind != VersionKindManual || snapshot.Note != "before large edit" {
		t.Fatalf("Snapshot metadata = %#v", snapshot)
	}

	updated, err := svc.Update(ctx, created.ID, UpdateRequest{Title: "Pinned", Content: "v2", BaseVersion: 1})
	if err != nil {
		t.Fatalf("Update v2: %v", err)
	}
	_, err = svc.Update(ctx, created.ID, UpdateRequest{Title: "Pinned", Content: "v3", BaseVersion: updated.Version})
	if err != nil {
		t.Fatalf("Update v3: %v", err)
	}

	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if len(versions.Items) != 2 {
		t.Fatalf("Version count = %d, want pinned v1 plus latest v3", len(versions.Items))
	}
	var sawPinnedV1 bool
	for _, item := range versions.Items {
		if item.Version == 1 && item.Pinned {
			sawPinnedV1 = true
		}
	}
	if !sawPinnedV1 {
		t.Fatalf("Pinned v1 missing from versions: %#v", versions.Items)
	}
}

func TestService_OverwritePolicy_SkipsAutoSnapshots(t *testing.T) {
	svc, _, _ := setupServiceWithVersioning(t, VersioningConfig{Policy: VersionPolicyOverwrite})
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{Title: "Overwrite", Content: "v1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.Update(ctx, created.ID, UpdateRequest{Title: "Overwrite", Content: "v2", BaseVersion: 1}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if len(versions.Items) != 0 {
		t.Fatalf("Version count = %d, want 0", len(versions.Items))
	}

	snapshot, err := svc.CreateSnapshot(ctx, created.ID, "manual")
	if err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	if snapshot.Version != 2 || !snapshot.Pinned {
		t.Fatalf("Snapshot = %#v, want pinned current v2", snapshot)
	}
}

func TestService_RestoreVersion_CreatesNewCurrentVersion(t *testing.T) {
	svc, _, _ := setupServiceWithVersioning(t, VersioningConfig{Policy: VersionPolicyFull})
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{Title: "Restore", Content: "# Restore\n\nv1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	updated, err := svc.Update(ctx, created.ID, UpdateRequest{Title: "Restore", Content: "# Restore\n\nv2", BaseVersion: 1})
	if err != nil {
		t.Fatalf("Update v2: %v", err)
	}

	restored, err := svc.RestoreVersion(ctx, created.ID, 1)
	if err != nil {
		t.Fatalf("RestoreVersion: %v", err)
	}
	if restored.Version != updated.Version+1 {
		t.Fatalf("Restored version = %d, want %d", restored.Version, updated.Version+1)
	}
	detail, err := svc.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if detail.Content != "# Restore\n\nv1" {
		t.Fatalf("Content = %q, want restored v1", detail.Content)
	}
	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}
	if versions.Items[0].Version != restored.Version || !versions.Items[0].Pinned {
		t.Fatalf("Latest snapshot = %#v, want pinned restored version", versions.Items[0])
	}
}

func TestService_RestoreVersion_OverwritePolicyPreservesCurrentBeforeRestore(t *testing.T) {
	svc, _, _ := setupServiceWithVersioning(t, VersioningConfig{Policy: VersionPolicyOverwrite})
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateRequest{Title: "Overwrite Restore", Content: "# A\n\nv1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.CreateSnapshot(ctx, created.ID, "baseline"); err != nil {
		t.Fatalf("CreateSnapshot: %v", err)
	}
	updated, err := svc.Update(ctx, created.ID, UpdateRequest{Title: "Overwrite Restore", Content: "# A\n\nv2", BaseVersion: 1})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if _, err := svc.RestoreVersion(ctx, created.ID, 1); err != nil {
		t.Fatalf("RestoreVersion: %v", err)
	}
	versions, err := svc.GetVersions(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetVersions: %v", err)
	}

	var sawBeforeRestore bool
	for _, item := range versions.Items {
		if item.Version == updated.Version && item.Pinned && item.Note == "before restore to v1" {
			sawBeforeRestore = true
		}
	}
	if !sawBeforeRestore {
		t.Fatalf("Missing before-restore snapshot in %#v", versions.Items)
	}
}
