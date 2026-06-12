package user

import (
	"context"
	"errors"
	"testing"
)

const strongPassword = "Str0ng!Password#2026"

func TestRegisterAndAuthenticate(t *testing.T) {
	svc := NewService(NewMemoryStore())
	ctx := context.Background()

	u, err := svc.Register(ctx, "Alice@Example.COM", "Alice", strongPassword)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if u.Email != "alice@example.com" {
		t.Fatalf("email not lowercased: %q", u.Email)
	}
	if u.PasswordHash == strongPassword || u.PasswordHash == "" {
		t.Fatal("password must be stored as a hash")
	}

	got, err := svc.Authenticate(ctx, "alice@example.com", strongPassword)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if got.ID != u.ID {
		t.Fatalf("authenticated wrong user: %s != %s", got.ID, u.ID)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	svc := NewService(NewMemoryStore())
	ctx := context.Background()
	if _, err := svc.Register(ctx, "a@b.com", "A", strongPassword); err != nil {
		t.Fatalf("first register: %v", err)
	}
	// Case-insensitive duplicate.
	if _, err := svc.Register(ctx, "A@B.COM", "A2", strongPassword); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	svc := NewService(NewMemoryStore())
	if _, err := svc.Register(context.Background(), "a@b.com", "A", "weak"); !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestAuthenticateWrongPassword(t *testing.T) {
	svc := NewService(NewMemoryStore())
	ctx := context.Background()
	if _, err := svc.Register(ctx, "a@b.com", "A", strongPassword); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := svc.Authenticate(ctx, "a@b.com", "Wrong!Password#2026"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthenticateUnknownEmailSameError(t *testing.T) {
	svc := NewService(NewMemoryStore())
	if _, err := svc.Authenticate(context.Background(), "ghost@b.com", strongPassword); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials for unknown email, got %v", err)
	}
}

func TestByIDNotFound(t *testing.T) {
	svc := NewService(NewMemoryStore())
	if _, err := svc.ByID(context.Background(), "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
