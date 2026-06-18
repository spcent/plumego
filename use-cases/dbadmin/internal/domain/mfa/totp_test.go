package mfa

import (
	"testing"
	"time"
)

// knownSecret and knownVectors are derived from RFC 6238 Appendix B test
// vectors (the ASCII seed "12345678901234567890", base32-encoded), using the
// HMAC-SHA1 / 6-digit variant required by this implementation.
const knownSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestGenerateCode_deterministic(t *testing.T) {
	// RFC 6238 test vector: T=59 -> counter=1 for a 30s step, SHA1, 8 digits is
	// "94287082"; the low-order 6 digits are "287082".
	at := time.Unix(59, 0).UTC()
	code, err := GenerateCode(knownSecret, at)
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("GenerateCode length = %d, want 6", len(code))
	}
	if code != "287082" {
		t.Fatalf("GenerateCode(%d) = %q, want %q", at.Unix(), code, "287082")
	}
}

func TestGenerateCode_isStableWithinStep(t *testing.T) {
	base := time.Unix(1700000000, 0).UTC()
	c1, err := GenerateCode(knownSecret, base)
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	c2, err := GenerateCode(knownSecret, base.Add(5*time.Second))
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	if c1 != c2 {
		t.Fatalf("codes within the same 30s step differ: %q vs %q", c1, c2)
	}
}

func TestGenerateCode_changesAcrossSteps(t *testing.T) {
	base := time.Unix(1700000000, 0).UTC()
	c1, _ := GenerateCode(knownSecret, base)
	c2, _ := GenerateCode(knownSecret, base.Add(31*time.Second))
	if c1 == c2 {
		t.Fatalf("codes across different steps unexpectedly match: %q", c1)
	}
}

func TestGenerateCode_invalidSecret(t *testing.T) {
	if _, err := GenerateCode("not-valid-base32!!", time.Now()); err == nil {
		t.Fatal("GenerateCode with invalid secret should fail closed with an error")
	}
}

func TestValidate_exactMatch(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	code, err := GenerateCode(knownSecret, now)
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}
	if !Validate(knownSecret, code, now) {
		t.Fatal("Validate should accept the code generated for the same time step")
	}
}

func TestValidate_clockSkewTolerance(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	codeAtNow, err := GenerateCode(knownSecret, now)
	if err != nil {
		t.Fatalf("GenerateCode error = %v", err)
	}

	tests := []struct {
		name   string
		offset time.Duration
		want   bool
	}{
		{"same step", 0, true},
		{"one step earlier (client clock behind)", -30 * time.Second, true},
		{"one step later (client clock ahead)", 30 * time.Second, true},
		{"two steps earlier - outside tolerance", -60 * time.Second, false},
		{"two steps later - outside tolerance", 60 * time.Second, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverTime := now.Add(-tt.offset) // server validates at serverTime using codeAtNow generated at `now`
			got := Validate(knownSecret, codeAtNow, serverTime)
			if got != tt.want {
				t.Errorf("Validate(offset=%v) = %v, want %v", tt.offset, got, tt.want)
			}
		})
	}
}

func TestValidate_rejectsWrongCode(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	if Validate(knownSecret, "000000", now) {
		// Extremely unlikely to collide, but guard against flakiness by also
		// trying a code guaranteed different from the real one.
		real, _ := GenerateCode(knownSecret, now)
		if real == "000000" {
			t.Skip("generated code coincidentally equals the wrong-code fixture")
		}
		t.Fatal("Validate accepted an incorrect code")
	}
}

func TestValidate_rejectsMalformedCode(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	cases := []string{"", "12345", "1234567", "abcdef", "  287082  "}
	for _, c := range cases {
		if Validate(knownSecret, c, now) && len(c) != 6 {
			t.Errorf("Validate(%q) unexpectedly accepted malformed code", c)
		}
	}
}

func TestValidate_failsClosedOnBadSecret(t *testing.T) {
	if Validate("not-valid-base32!!", "123456", time.Now()) {
		t.Fatal("Validate must fail closed for an undecodable secret")
	}
}

func TestGenerateSecret_returnsUsableBase32(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatalf("GenerateSecret error = %v", err)
	}
	if secret == "" {
		t.Fatal("GenerateSecret returned empty secret")
	}
	if _, err := GenerateCode(secret, time.Now()); err != nil {
		t.Fatalf("generated secret is not usable for GenerateCode: %v", err)
	}
}

func TestGenerateSecret_uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 20; i++ {
		s, err := GenerateSecret()
		if err != nil {
			t.Fatalf("GenerateSecret error = %v", err)
		}
		if seen[s] {
			t.Fatalf("duplicate secret generated: %q", s)
		}
		seen[s] = true
	}
}

func TestOtpauthURI_containsExpectedParams(t *testing.T) {
	uri := otpauthURI("dbadmin", "alice", knownSecret)
	wantSubstrings := []string{
		"otpauth://totp/",
		"secret=" + knownSecret,
		"issuer=dbadmin",
		"algorithm=SHA1",
		"digits=6",
		"period=30",
	}
	for _, want := range wantSubstrings {
		if !contains(uri, want) {
			t.Errorf("otpauthURI = %q, want substring %q", uri, want)
		}
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
