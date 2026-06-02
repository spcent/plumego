package importer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-vault/internal/database"
	"cloud-vault/internal/document"
	"cloud-vault/internal/storage"
)

// --- helpers ---

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

func setupService(t *testing.T) (*Service, string) {
	t.Helper()
	db := openTestDB(t)
	repo := NewRepository(db)

	docRepo := document.NewSQLiteRepository(db)
	store := storage.NewLocalStorage(t.TempDir())
	docSvc := document.NewService(docRepo, store)

	safeRoot := t.TempDir()
	t.Setenv(importerSafeRootEnv, safeRoot)

	svc := NewService(repo, docSvc, Config{})
	return svc, safeRoot
}

// --- resolveAndValidateSourcePath tests ---

func TestResolveAndValidateSourcePath_empty(t *testing.T) {
	root := t.TempDir()
	t.Setenv(importerSafeRootEnv, root)

	_, err := resolveAndValidateSourcePath("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestResolveAndValidateSourcePath_whitespaceOnly(t *testing.T) {
	root := t.TempDir()
	t.Setenv(importerSafeRootEnv, root)

	_, err := resolveAndValidateSourcePath("   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only path")
	}
}

func TestResolveAndValidateSourcePath_valid(t *testing.T) {
	root := t.TempDir()
	t.Setenv(importerSafeRootEnv, root)

	// Create a subdirectory inside the safe root.
	sub := filepath.Join(root, "subdir")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	resolved, err := resolveAndValidateSourcePath("subdir")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != sub {
		t.Errorf("want %q, got %q", sub, resolved)
	}
}

func TestResolveAndValidateSourcePath_traversal(t *testing.T) {
	root := t.TempDir()
	t.Setenv(importerSafeRootEnv, root)

	_, err := resolveAndValidateSourcePath("../../etc/passwd")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestResolveAndValidateSourcePath_rootItself(t *testing.T) {
	root := t.TempDir()
	t.Setenv(importerSafeRootEnv, root)

	// Joining root with "." resolves to root itself, which should be allowed.
	resolved, err := resolveAndValidateSourcePath(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// resolved should equal root
	rootAbs, _ := filepath.Abs(root)
	if resolved != rootAbs {
		t.Errorf("want %q, got %q", rootAbs, resolved)
	}
}

// --- Service.CreateJob ---

func TestService_CreateJob_valid(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	// Create a directory with some .md files inside the safe root.
	srcDir := filepath.Join(safeRoot, "docs")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.md", "b.md"} {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte("# "+name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	resp, err := svc.CreateJob(ctx, CreateJobRequest{SourcePath: "docs"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if resp.ID == "" {
		t.Error("ID should not be empty")
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount: want 2, got %d", resp.TotalCount)
	}
	if resp.Status != JobStatusPending {
		t.Errorf("Status: want %q, got %q", JobStatusPending, resp.Status)
	}
}

func TestService_CreateJob_invalidPath(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.CreateJob(ctx, CreateJobRequest{SourcePath: ""})
	if err == nil {
		t.Fatal("expected error for empty source_path")
	}
}

func TestService_CreateJob_traversal(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.CreateJob(ctx, CreateJobRequest{SourcePath: "../../somewhere"})
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestService_CreateJob_noMdFiles(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	emptyDir := filepath.Join(safeRoot, "empty")
	if err := os.Mkdir(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	resp, err := svc.CreateJob(ctx, CreateJobRequest{SourcePath: "empty"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if resp.TotalCount != 0 {
		t.Errorf("TotalCount: want 0, got %d", resp.TotalCount)
	}
}

func TestService_CreateJob_nameDefaultsToSourcePath(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	srcDir := filepath.Join(safeRoot, "mydir")
	os.Mkdir(srcDir, 0o755)

	resp, err := svc.CreateJob(ctx, CreateJobRequest{SourcePath: "mydir"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	// Name defaults to SourcePath when not provided.
	if resp.Name == "" {
		t.Error("Name should not be empty")
	}
}

// --- Service.GetJob ---

func TestService_GetJob(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	srcDir := filepath.Join(safeRoot, "getjob")
	os.Mkdir(srcDir, 0o755)

	created, err := svc.CreateJob(ctx, CreateJobRequest{Name: "get-test", SourcePath: "getjob"})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	got, err := svc.GetJob(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID: want %q, got %q", created.ID, got.ID)
	}
}

func TestService_GetJob_notFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	_, err := svc.GetJob(ctx, "no-such-id")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// --- Service.ListJobs ---

func TestService_ListJobs(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	for _, name := range []string{"j1", "j2", "j3"} {
		d := filepath.Join(safeRoot, name)
		os.Mkdir(d, 0o755)
		svc.CreateJob(ctx, CreateJobRequest{Name: name, SourcePath: name})
	}

	result, err := svc.ListJobs(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total: want 3, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("Items: want 3, got %d", len(result.Items))
	}
}

func TestService_ListJobs_defaultLimit(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	result, err := svc.ListJobs(ctx, 0, 0)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if result.Limit != 20 {
		t.Errorf("Limit: want 20 (default), got %d", result.Limit)
	}
}

func TestService_ListJobs_maxLimitCapped(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t)

	result, err := svc.ListJobs(ctx, 999, 0)
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if result.Limit != 100 {
		t.Errorf("Limit: want 100 (capped), got %d", result.Limit)
	}
}

// --- Service.PauseJob ---

func TestService_PauseJob(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	d := filepath.Join(safeRoot, "pause")
	os.Mkdir(d, 0o755)
	created, _ := svc.CreateJob(ctx, CreateJobRequest{Name: "pausable", SourcePath: "pause"})

	if err := svc.PauseJob(ctx, created.ID); err != nil {
		t.Fatalf("PauseJob: %v", err)
	}

	got, err := svc.GetJob(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Status != JobStatusPaused {
		t.Errorf("Status: want %q, got %q", JobStatusPaused, got.Status)
	}
}

func TestService_PauseJob_notPausable(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	d := filepath.Join(safeRoot, "done")
	os.Mkdir(d, 0o755)
	created, _ := svc.CreateJob(ctx, CreateJobRequest{Name: "done-job", SourcePath: "done"})

	// Manually set status to done via cancel (which sets to cancelled) then test
	svc.CancelJob(ctx, created.ID)

	err := svc.PauseJob(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error pausing a cancelled job")
	}
}

// --- Service.CancelJob ---

func TestService_CancelJob(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	d := filepath.Join(safeRoot, "cancel")
	os.Mkdir(d, 0o755)
	created, _ := svc.CreateJob(ctx, CreateJobRequest{Name: "cancellable", SourcePath: "cancel"})

	if err := svc.CancelJob(ctx, created.ID); err != nil {
		t.Fatalf("CancelJob: %v", err)
	}

	got, err := svc.GetJob(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Status != JobStatusCancelled {
		t.Errorf("Status: want %q, got %q", JobStatusCancelled, got.Status)
	}
}

// --- Service.ListItems ---

func TestService_ListItems(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	d := filepath.Join(safeRoot, "items")
	os.Mkdir(d, 0o755)
	for _, name := range []string{"x.md", "y.md", "z.md"} {
		os.WriteFile(filepath.Join(d, name), []byte("# "+name), 0o644)
	}

	created, _ := svc.CreateJob(ctx, CreateJobRequest{Name: "items-job", SourcePath: "items"})

	result, err := svc.ListItems(ctx, created.ID, "", 10, 0)
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("Total: want 3, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Errorf("Items: want 3, got %d", len(result.Items))
	}
}

func TestService_ListItems_statusFilter(t *testing.T) {
	ctx := context.Background()
	svc, safeRoot := setupService(t)

	d := filepath.Join(safeRoot, "filter")
	os.Mkdir(d, 0o755)
	os.WriteFile(filepath.Join(d, "a.md"), []byte("# A"), 0o644)

	created, _ := svc.CreateJob(ctx, CreateJobRequest{Name: "filter-job", SourcePath: "filter"})

	// Filter for "success" — should return 0 since job hasn't run.
	result, err := svc.ListItems(ctx, created.ID, "success", 10, 0)
	if err != nil {
		t.Fatalf("ListItems: %v", err)
	}
	if result.Total != 0 {
		t.Errorf("Total: want 0 for success filter, got %d", result.Total)
	}
}
