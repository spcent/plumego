package stringsx

import (
	"testing"
	"unicode/utf8"
)

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
	if got := MaskSecret("密码测试123"); got != "密码***23" {
		t.Fatalf("expected unicode-safe mask, got %q", got)
	} else if !utf8.ValidString(got) {
		t.Fatalf("masked secret is not valid utf8: %q", got)
	}
	if got := MaskSecret("密码"); got != "**" {
		t.Fatalf("expected short unicode secret to be fully masked, got %q", got)
	}
}
