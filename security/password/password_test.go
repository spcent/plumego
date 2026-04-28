package password

import (
	"errors"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hashed == password {
		t.Error("Hashed password should not be the same as plaintext")
	}
	if len(hashed) == 0 {
		t.Error("Hashed password should not be empty")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	err = CheckPassword(hashed, password)
	if err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}

	// Test incorrect password
	err = CheckPassword(hashed, "wrongpassword")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("CheckPassword incorrect error = %v, want ErrPasswordMismatch", err)
	}

	// Test empty password
	err = CheckPassword(hashed, "")
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Errorf("CheckPassword empty error = %v, want ErrPasswordMismatch", err)
	}
}

func TestCheckPasswordDifferentCosts(t *testing.T) {
	// Test that hashing with different costs still works for verification
	password := "testpassword123"
	hashed1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	err = CheckPassword(hashed1, password)
	if err != nil {
		t.Errorf("CheckPassword failed for correct password: %v", err)
	}
}
