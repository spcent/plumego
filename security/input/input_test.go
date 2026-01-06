package input

import "testing"

func TestIsToken(t *testing.T) {
	cases := []struct {
		value string
		ok    bool
	}{
		{"X-Test", true},
		{"content-type", true},
		{"abc123", true},
		{"", false},
		{"bad value", false},
		{"bad\tvalue", false},
		{"bad,value", false},
		{"bad:value", false},
	}

	for _, tc := range cases {
		if got := IsToken(tc.value); got != tc.ok {
			t.Fatalf("IsToken(%q) = %v, want %v", tc.value, got, tc.ok)
		}
	}
}

func TestIsHeaderValue(t *testing.T) {
	if !IsHeaderValue("nosniff") {
		t.Fatalf("expected safe header value")
	}

	if IsHeaderValue("bad\nvalue") {
		t.Fatalf("expected newline to be rejected")
	}

	if IsHeaderValue("bad\rvalue") {
		t.Fatalf("expected carriage return to be rejected")
	}

	if IsHeaderValue(string([]byte{0xff})) {
		t.Fatalf("expected invalid utf-8 to be rejected")
	}
}
