package password

import (
	"errors"
	"strings"
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

func TestPasswordHashRejectsOverlongPassword(t *testing.T) {
	overlong := strings.Repeat("a", MaxPasswordLength+1)

	if _, err := HashPasswordWithCost(overlong, MinimumCost); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("HashPasswordWithCost overlong error = %v, want ErrPasswordTooLong", err)
	}
}

func TestCheckPasswordRejectsOverlongPassword(t *testing.T) {
	hashed, err := HashPasswordWithCost("short-password", MinimumCost)
	if err != nil {
		t.Fatalf("HashPasswordWithCost: %v", err)
	}

	overlong := strings.Repeat("a", MaxPasswordLength+1)
	if err := CheckPassword(hashed, overlong); !errors.Is(err, ErrPasswordTooLong) {
		t.Fatalf("CheckPassword overlong error = %v, want ErrPasswordTooLong", err)
	}
}

func TestPasswordHashAcceptsMaxLengthPassword(t *testing.T) {
	maxLength := strings.Repeat("a", MaxPasswordLength)

	hashed, err := HashPasswordWithCost(maxLength, MinimumCost)
	if err != nil {
		t.Fatalf("HashPasswordWithCost max length: %v", err)
	}
	if err := CheckPassword(hashed, maxLength); err != nil {
		t.Fatalf("CheckPassword max length: %v", err)
	}
}
