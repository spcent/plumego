package ai

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

// mockDocSaver implements DocumentSaver for testing
type mockDocSaver struct {
	savedDocs []savedDoc
}

type savedDoc struct {
	title     string
	content   string
	sourceIDs []string
}

func (m *mockDocSaver) SaveAIDocument(ctx context.Context, title, content, storageKey string, sourceIDs []string) (string, error) {
	m.savedDocs = append(m.savedDocs, savedDoc{title, content, sourceIDs})
	return "mock-doc-id", nil
}

func setupAI(t *testing.T) (*Service, *document.Service, *database.DB, *mockDocSaver) {
	t.Helper()
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())

	// Document service
	docRepo := document.NewSQLiteRepository(db)
	docSvc := document.NewService(docRepo, store)

	// AI service with mock provider
	aiRepo := NewRepository(db.DB)
	provider := NewMockProvider()
	docSaver := &mockDocSaver{}
	aiSvc := NewService(aiRepo, store, config.AIConfig{Enabled: true, Model: "mock-model"}, provider, docSaver)

	return aiSvc, docSvc, db, docSaver
}

func TestAI_EnqueueSummary(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	// Create a document to summarize
	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Summarize Me",
		Content: "# Important Document\n\nThis is the content to summarize.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Enqueue summary task
	task, err := aiSvc.EnqueueSummary(ctx, doc.ID)
	if err != nil {
		t.Fatalf("EnqueueSummary: %v", err)
	}

	if task.ID == "" {
		t.Error("Task ID is empty")
	}
	if task.TaskType != TaskTypeDocumentSummary {
		t.Errorf("TaskType = %q, want %q", task.TaskType, TaskTypeDocumentSummary)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("Status = %q, want %q", task.Status, TaskStatusPending)
	}
}

func TestAI_ProcessTask_Summary(t *testing.T) {
	aiSvc, docSvc, _, docSaver := setupAI(t)
	ctx := context.Background()

	// Create document
	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Process Test",
		Content: "# Test\n\nContent for processing.",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Enqueue and process
	task, err := aiSvc.EnqueueSummary(ctx, doc.ID)
	if err != nil {
		t.Fatalf("EnqueueSummary: %v", err)
	}

	err = aiSvc.ProcessTask(ctx, task)
	if err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}

	// Verify task completed
	updatedTask, err := aiSvc.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if updatedTask.Status != TaskStatusCompleted {
		t.Errorf("Status = %q, want %q", updatedTask.Status, TaskStatusCompleted)
	}

	// Verify summary stored
	summary, err := aiSvc.GetSummary(ctx, doc.ID)
	if err != nil {
		t.Fatalf("GetSummary: %v", err)
	}
	if summary == nil {
		t.Fatal("Summary is nil")
	}
	if summary.Summary == "" {
		t.Error("Summary content is empty")
	}

	// Verify document saved
	if len(docSaver.savedDocs) != 1 {
		t.Errorf("Saved docs = %d, want 1", len(docSaver.savedDocs))
	}
}

func TestAI_EnqueueQA(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	// Create documents for Q&A
	doc1, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Doc 1",
		Content: "# Doc 1\n\nFirst document.",
	})
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}

	doc2, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Doc 2",
		Content: "# Doc 2\n\nSecond document.",
	})
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}

	// Enqueue Q&A
	task, err := aiSvc.EnqueueQA(ctx, "What is this?", []string{doc1.ID, doc2.ID})
	if err != nil {
		t.Fatalf("EnqueueQA: %v", err)
	}

	if task.TaskType != TaskTypeSearchAnswer {
		t.Errorf("TaskType = %q, want %q", task.TaskType, TaskTypeSearchAnswer)
	}
}

func TestAI_CancelTask(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	doc, err := docSvc.Create(ctx, document.CreateRequest{
		Title:   "Cancel Test",
		Content: "Content",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	task, err := aiSvc.EnqueueSummary(ctx, doc.ID)
	if err != nil {
		t.Fatalf("EnqueueSummary: %v", err)
	}

	// Cancel
	err = aiSvc.CancelTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("CancelTask: %v", err)
	}

	// Verify cancelled
	updatedTask, err := aiSvc.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if updatedTask.Status != TaskStatusCancelled {
		t.Errorf("Status = %q, want %q", updatedTask.Status, TaskStatusCancelled)
	}
}

func TestAI_ListTasks(t *testing.T) {
	aiSvc, docSvc, _, _ := setupAI(t)
	ctx := context.Background()

	// Create multiple tasks
	for i := 0; i < 3; i++ {
		doc, err := docSvc.Create(ctx, document.CreateRequest{
			Title:   "Doc " + string(rune('A'+i)),
			Content: "Content " + string(rune('A'+i)),
		})
		if err != nil {
			t.Fatalf("Create %d: %v", i, err)
		}
		_, err = aiSvc.EnqueueSummary(ctx, doc.ID)
		if err != nil {
			t.Fatalf("EnqueueSummary %d: %v", i, err)
		}
	}

	// List all
	tasks, total, err := aiSvc.ListTasks(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if total != 3 {
		t.Errorf("Total = %d, want 3", total)
	}
	if len(tasks) != 3 {
		t.Errorf("Tasks count = %d, want 3", len(tasks))
	}

	// List by status
	tasks2, total2, err := aiSvc.ListTasks(ctx, TaskStatusPending, 10, 0)
	if err != nil {
		t.Fatalf("ListTasks pending: %v", err)
	}
	if total2 != 3 {
		t.Errorf("Pending total = %d, want 3", total2)
	}
	if len(tasks2) != 3 {
		t.Errorf("Pending tasks = %d, want 3", len(tasks2))
	}
}
