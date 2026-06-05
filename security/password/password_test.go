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

// TestHashPasswordWithCostInvalidCostTooLow exercises the ErrInvalidCost path
// when cost is below MinimumCost.
func TestHashPasswordWithCostInvalidCostTooLow(t *testing.T) {
	_, err := HashPasswordWithCost("password", MinimumCost-1)
	if !errors.Is(err, ErrInvalidCost) {
		t.Fatalf("error = %v, want ErrInvalidCost", err)
	}
}

// TestHashPasswordWithCostInvalidCostTooHigh exercises the ErrInvalidCost path
// when cost is above MaximumCost.
func TestHashPasswordWithCostInvalidCostTooHigh(t *testing.T) {
	_, err := HashPasswordWithCost("password", MaximumCost+1)
	if !errors.Is(err, ErrInvalidCost) {
		t.Fatalf("error = %v, want ErrInvalidCost", err)
	}
}

// TestCheckPasswordInvalidHashFormats covers the various ErrInvalidHash paths
// in CheckPassword (not enough segments, non-numeric cost, cost out of range,
// bad salt base64, wrong salt length, bad hash base64, wrong hash length).
func TestCheckPasswordInvalidHashFormats(t *testing.T) {
	cases := []struct {
		name   string
		hashed string
	}{
		{"too few segments", "noCost$onlyTwo"},
		{"non-numeric cost", "notanint$c2FsdA==$aGFzaA=="},
		{"cost too low", "99999$c2FsdA==$aGFzaA=="},
		{"cost too high", "2000001$c2FsdA==$aGFzaA=="},
		{"bad salt base64", "100000$not!base64$aGFzaA=="},
		// Valid base64 but wrong decoded length (not 16 bytes).
		{"wrong salt length", "100000$dGVzdA==$aGFzaA=="},
		// Valid cost + valid 16-byte salt + bad hash base64.
		{"bad hash base64", "100000$AAAAAAAAAAAAAAAAAAAAAA==$not!base64"},
		// Valid cost + valid 16-byte salt + valid base64 but wrong hash length (not 32 bytes).
		{"wrong hash length", "100000$AAAAAAAAAAAAAAAAAAAAAA==$aGFzaA=="},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CheckPassword(tc.hashed, "password")
			if !errors.Is(err, ErrInvalidHash) {
				t.Fatalf("CheckPassword(%q) error = %v, want ErrInvalidHash", tc.hashed, err)
			}
		})
	}
}
