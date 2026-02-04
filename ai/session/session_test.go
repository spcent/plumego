package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/spcent/plumego/ai/provider"
	"github.com/spcent/plumego/ai/tokenizer"
)

func TestManager_Create(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	opts := CreateOptions{
		TenantID: "tenant-1",
		UserID:   "user-1",
		AgentID:  "agent-1",
		Model:    "claude-3-opus",
	}

	session, err := manager.Create(context.Background(), opts)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}

	if session.TenantID != opts.TenantID {
		t.Errorf("TenantID = %v, want %v", session.TenantID, opts.TenantID)
	}

	if session.UserID != opts.UserID {
		t.Errorf("UserID = %v, want %v", session.UserID, opts.UserID)
	}
}

func TestManager_Get(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	// Create session
	created, err := manager.Create(context.Background(), CreateOptions{
		TenantID: "tenant-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Get session
	retrieved, err := manager.Get(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID = %v, want %v", retrieved.ID, created.ID)
	}
}

func TestManager_Get_NotFound(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	_, err := manager.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Error("Get() should return error for nonexistent session")
	}
}

func TestManager_Get_Expired(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	// Create session with very short TTL
	session, err := manager.Create(context.Background(), CreateOptions{
		TenantID: "tenant-1",
		TTL:      1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Try to get expired session
	_, err = manager.Get(context.Background(), session.ID)
	if err == nil {
		t.Error("Get() should return error for expired session")
	}
}

func TestManager_AppendMessage(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage, WithTokenizer(tokenizer.NewSimpleTokenizer("test")))

	session, err := manager.Create(context.Background(), CreateOptions{
		TenantID: "tenant-1",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Append message
	msg := provider.NewTextMessage(provider.RoleUser, "Hello, world!")
	if err := manager.AppendMessage(context.Background(), session.ID, msg); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	// Verify message was added
	updated, err := manager.Get(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if len(updated.Messages) != 1 {
		t.Errorf("Message count = %v, want 1", len(updated.Messages))
	}

	if updated.Messages[0].GetText() != "Hello, world!" {
		t.Errorf("Message text = %v, want 'Hello, world!'", updated.Messages[0].GetText())
	}
}

func TestManager_GetMessages(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	session, err := manager.Create(context.Background(), CreateOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Add multiple messages
	messages := []string{"Message 1", "Message 2", "Message 3", "Message 4", "Message 5"}
	for _, text := range messages {
		msg := provider.NewTextMessage(provider.RoleUser, text)
		if err := manager.AppendMessage(context.Background(), session.ID, msg); err != nil {
			t.Fatalf("AppendMessage() error = %v", err)
		}
	}

	// Get messages with pagination
	retrieved, err := manager.GetMessages(context.Background(), session.ID, 1, 3)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Message count = %v, want 3", len(retrieved))
	}

	if retrieved[0].GetText() != "Message 2" {
		t.Errorf("First message = %v, want 'Message 2'", retrieved[0].GetText())
	}
}

func TestManager_GetActiveContext(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage, WithTokenizer(tokenizer.NewSimpleTokenizer("test")))

	session, err := manager.Create(context.Background(), CreateOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Add messages
	for i := 0; i < 10; i++ {
		msg := provider.NewTextMessage(provider.RoleUser, "This is a test message")
		if err := manager.AppendMessage(context.Background(), session.ID, msg); err != nil {
			t.Fatalf("AppendMessage() error = %v", err)
		}
	}

	// Get active context with token limit
	active, err := manager.GetActiveContext(context.Background(), session.ID, 50)
	if err != nil {
		t.Fatalf("GetActiveContext() error = %v", err)
	}

	// Should return fewer messages to fit token budget
	if len(active) >= 10 {
		t.Errorf("Active context should be trimmed, got %v messages", len(active))
	}
}

func TestManager_SetContext(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	session, err := manager.Create(context.Background(), CreateOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Set context
	if err := manager.SetContext(context.Background(), session.ID, "key1", "value1"); err != nil {
		t.Fatalf("SetContext() error = %v", err)
	}

	// Get context
	value, err := manager.GetContext(context.Background(), session.ID, "key1")
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if value != "value1" {
		t.Errorf("Context value = %v, want 'value1'", value)
	}
}

func TestManager_Delete(t *testing.T) {
	storage := NewMemoryStorage()
	manager := NewManager(storage)

	session, err := manager.Create(context.Background(), CreateOptions{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Delete session
	if err := manager.Delete(context.Background(), session.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = manager.Get(context.Background(), session.ID)
	if err == nil {
		t.Error("Get() should return error for deleted session")
	}
}

func TestMemoryStorage(t *testing.T) {
	storage := NewMemoryStorage()

	session := &Session{
		ID:       "test-1",
		TenantID: "tenant-1",
		Messages: []provider.Message{
			provider.NewTextMessage(provider.RoleUser, "Hello"),
		},
		Context: map[string]any{
			"key": "value",
		},
	}

	// Save
	if err := storage.Save(context.Background(), session); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := storage.Load(context.Background(), "test-1")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.ID != session.ID {
		t.Errorf("ID = %v, want %v", loaded.ID, session.ID)
	}

	// Delete
	if err := storage.Delete(context.Background(), "test-1"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = storage.Load(context.Background(), "test-1")
	if err == nil {
		t.Error("Load() should return error for deleted session")
	}
}

func TestMemoryStorage_Count(t *testing.T) {
	storage := NewMemoryStorage()

	if count := storage.Count(); count != 0 {
		t.Errorf("Initial count = %v, want 0", count)
	}

	// Add sessions
	for i := 0; i < 5; i++ {
		session := &Session{ID: fmt.Sprintf("sess-%d", i)}
		storage.Save(context.Background(), session)
	}

	if count := storage.Count(); count != 5 {
		t.Errorf("Count = %v, want 5", count)
	}

	// Clear
	storage.Clear()

	if count := storage.Count(); count != 0 {
		t.Errorf("Count after Clear() = %v, want 0", count)
	}
}
