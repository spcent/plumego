package stringsx

import "testing"

func TestMaskSecret(t *testing.T) {
	if got := MaskSecret(""); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
	if got := MaskSecret("abcd"); got != "****" {
		t.Fatalf("expected ****, got %q", got)
	}
	if got := MaskSecret("abcdef"); got != "ab**ef" {
		t.Fatalf("expected ab**ef, got %q", got)
	}
}
