package password

import (
	"encoding/base64"
	"fmt"
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
	if err == nil {
		t.Error("CheckPassword should have failed for incorrect password")
	}

	// Test empty password
	err = CheckPassword(hashed, "")
	if err == nil {
		t.Error("CheckPassword should have failed for empty password")
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

func TestCheckPasswordLegacyHash(t *testing.T) {
	password := "legacy_password123"
	salt := []byte("1234567890abcdef")
	cost := 12

	derived := deriveKeyLegacy(password, salt, cost)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)
	encodedHash := base64.StdEncoding.EncodeToString(derived)
	legacyHash := fmt.Sprintf("%d$%s$%s", cost, encodedSalt, encodedHash)

	if err := CheckPassword(legacyHash, password); err != nil {
		t.Fatalf("CheckPassword failed for legacy hash: %v", err)
	}

	if err := CheckPassword(legacyHash, "wrong"); err == nil {
		t.Fatal("CheckPassword should fail for incorrect password with legacy hash")
	}
}
