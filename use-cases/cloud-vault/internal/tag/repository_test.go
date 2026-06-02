package tag

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

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

func newTestRepo(t *testing.T) *SQLiteRepository {
	t.Helper()
	return NewSQLiteRepository(openTestDB(t))
}

// tagTestEnv holds both repo and a helper to create documents (needed for FK).
type tagTestEnv struct {
	repo   *SQLiteRepository
	docSvc *document.Service
}

func setupTagEnv(t *testing.T) *tagTestEnv {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)
	return &tagTestEnv{
		repo:   NewSQLiteRepository(db),
		docSvc: docSvc,
	}
}

func (e *tagTestEnv) createDoc(t *testing.T, title string) string {
	t.Helper()
	result, err := e.docSvc.Create(context.Background(), document.CreateRequest{
		Title:   title,
		Content: "# " + title,
	})
	if err != nil {
		t.Fatalf("create doc %q: %v", title, err)
	}
	return result.ID
}

func TestSQLiteRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	tag := &Tag{ID: newID(), Name: "golang", Color: "#00ADD8", Source: "manual"}
	if err := repo.Create(ctx, tag); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "golang" {
		t.Errorf("Name: want 'golang', got %q", got.Name)
	}
	if got.Color != "#00ADD8" {
		t.Errorf("Color: want '#00ADD8', got %q", got.Color)
	}
}

func TestSQLiteRepository_Create_duplicate(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	tag := &Tag{ID: newID(), Name: "dup"}
	if err := repo.Create(ctx, tag); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	tag2 := &Tag{ID: newID(), Name: "dup"}
	err := repo.Create(ctx, tag2)
	if !errors.Is(err, ErrDuplicate) {
		t.Errorf("want ErrDuplicate, got %v", err)
	}
}

func TestSQLiteRepository_GetByID_notFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	_, err := repo.GetByID(ctx, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestSQLiteRepository_List(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		if err := repo.Create(ctx, &Tag{ID: newID(), Name: name}); err != nil {
			t.Fatalf("Create %q: %v", name, err)
		}
	}

	tags, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tags) != 3 {
		t.Errorf("want 3 tags, got %d", len(tags))
	}
}

func TestSQLiteRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	tag := &Tag{ID: newID(), Name: "original", Color: ""}
	if err := repo.Create(ctx, tag); err != nil {
		t.Fatalf("Create: %v", err)
	}

	tag.Name = "updated"
	tag.Color = "#ff0000"
	if err := repo.Update(ctx, tag); err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(ctx, tag.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "updated" {
		t.Errorf("Name: want 'updated', got %q", got.Name)
	}
	if got.Color != "#ff0000" {
		t.Errorf("Color: want '#ff0000', got %q", got.Color)
	}
}

func TestSQLiteRepository_Update_notFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	err := repo.Update(ctx, &Tag{ID: "no-such-id", Name: "x"})
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestSQLiteRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	tag := &Tag{ID: newID(), Name: "to-delete"}
	if err := repo.Create(ctx, tag); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, tag.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, tag.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound after delete, got %v", err)
	}
}

func TestSQLiteRepository_Delete_notFound(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	err := repo.Delete(ctx, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

func TestSQLiteRepository_GetDocumentTags(t *testing.T) {
	ctx := context.Background()
	env := setupTagEnv(t)

	tagA := &Tag{ID: newID(), Name: "go"}
	tagB := &Tag{ID: newID(), Name: "rust"}
	env.repo.Create(ctx, tagA)
	env.repo.Create(ctx, tagB)

	docID := env.createDoc(t, "TaggedDoc")
	if err := env.repo.SetDocumentTags(ctx, docID, []string{tagA.ID, tagB.ID}); err != nil {
		t.Fatalf("SetDocumentTags: %v", err)
	}

	tags, err := env.repo.GetDocumentTags(ctx, docID)
	if err != nil {
		t.Fatalf("GetDocumentTags: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("want 2 tags, got %d", len(tags))
	}
}

func TestSQLiteRepository_SetDocumentTags_replacesAll(t *testing.T) {
	ctx := context.Background()
	env := setupTagEnv(t)

	tagA := &Tag{ID: newID(), Name: "a"}
	tagB := &Tag{ID: newID(), Name: "b"}
	tagC := &Tag{ID: newID(), Name: "c"}
	env.repo.Create(ctx, tagA)
	env.repo.Create(ctx, tagB)
	env.repo.Create(ctx, tagC)

	docID := env.createDoc(t, "SetDoc")
	env.repo.SetDocumentTags(ctx, docID, []string{tagA.ID, tagB.ID})
	// Replace with only tagC
	if err := env.repo.SetDocumentTags(ctx, docID, []string{tagC.ID}); err != nil {
		t.Fatalf("SetDocumentTags: %v", err)
	}

	tags, err := env.repo.GetDocumentTags(ctx, docID)
	if err != nil {
		t.Fatalf("GetDocumentTags: %v", err)
	}
	if len(tags) != 1 || tags[0].ID != tagC.ID {
		t.Errorf("want only tagC, got %+v", tags)
	}
}

func TestSQLiteRepository_RemoveDocumentTag(t *testing.T) {
	ctx := context.Background()
	env := setupTagEnv(t)

	tagA := &Tag{ID: newID(), Name: "x"}
	tagB := &Tag{ID: newID(), Name: "y"}
	env.repo.Create(ctx, tagA)
	env.repo.Create(ctx, tagB)

	docID := env.createDoc(t, "RemoveDoc")
	env.repo.SetDocumentTags(ctx, docID, []string{tagA.ID, tagB.ID})

	if err := env.repo.RemoveDocumentTag(ctx, docID, tagA.ID); err != nil {
		t.Fatalf("RemoveDocumentTag: %v", err)
	}

	tags, err := env.repo.GetDocumentTags(ctx, docID)
	if err != nil {
		t.Fatalf("GetDocumentTags: %v", err)
	}
	if len(tags) != 1 || tags[0].ID != tagB.ID {
		t.Errorf("want only tagB remaining, got %+v", tags)
	}
}

func TestSQLiteRepository_EnsureTag_creates(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	tag, err := repo.EnsureTag(ctx, "ensure-new", "#123", "imported")
	if err != nil {
		t.Fatalf("EnsureTag: %v", err)
	}
	if tag.Name != "ensure-new" {
		t.Errorf("Name: want 'ensure-new', got %q", tag.Name)
	}
	if tag.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestSQLiteRepository_EnsureTag_existingReturned(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	// Create once
	first, err := repo.EnsureTag(ctx, "existing", "#abc", "manual")
	if err != nil {
		t.Fatalf("first EnsureTag: %v", err)
	}

	// Call again, should return same tag
	second, err := repo.EnsureTag(ctx, "existing", "#abc", "manual")
	if err != nil {
		t.Fatalf("second EnsureTag: %v", err)
	}
	if first.ID != second.ID {
		t.Errorf("IDs differ: %q vs %q", first.ID, second.ID)
	}
}
