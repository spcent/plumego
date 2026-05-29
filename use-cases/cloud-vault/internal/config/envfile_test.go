package config

import (
	"strings"
	"testing"
)

func TestParseDotEnv(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "basic key=value",
			input: "KEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "multiple lines",
			input: "A=1\nB=2\nC=3",
			want:  map[string]string{"A": "1", "B": "2", "C": "3"},
		},
		{
			name:  "double quoted value",
			input: `KEY="hello world"`,
			want:  map[string]string{"KEY": "hello world"},
		},
		{
			name:  "single quoted value",
			input: `KEY='hello world'`,
			want:  map[string]string{"KEY": "hello world"},
		},
		{
			name:  "blank lines ignored",
			input: "A=1\n\nB=2\n\n",
			want:  map[string]string{"A": "1", "B": "2"},
		},
		{
			name:  "comments ignored",
			input: "# comment\nA=1\n# another\nB=2",
			want:  map[string]string{"A": "1", "B": "2"},
		},
		{
			name:  "inline comments stripped",
			input: "KEY=value # comment",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "whitespace trimmed",
			input: "  KEY  =  value  ",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "empty value",
			input: "KEY=",
			want:  map[string]string{"KEY": ""},
		},
		{
			name:  "no equals sign skipped",
			input: "INVALID\nKEY=value",
			want:  map[string]string{"KEY": "value"},
		},
		{
			name:  "equals in value preserved",
			input: "KEY=val=ue",
			want:  map[string]string{"KEY": "val=ue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDotEnv(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parseDotEnv() error = %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("len(got) = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("got[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestStripQuotes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`hello`, "hello"},
		{`""`, ""},
		{`''`, ""},
		{`"`, `"`},
		{`"hello'`, `"hello'`},
	}
	for _, tt := range tests {
		if got := stripQuotes(tt.input); got != tt.want {
			t.Errorf("stripQuotes(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCombineLookup(t *testing.T) {
	dotenv := map[string]string{
		"A": "from-dotenv",
		"B": "from-dotenv",
	}
	envLookup := func(key string) (string, bool) {
		if key == "A" {
			return "from-env", true
		}
		return "", false
	}

	combined := combineLookup(dotenv, envLookup)

	// Env var wins
	if v, ok := combined("A"); !ok || v != "from-env" {
		t.Errorf("combined(A) = (%q, %v), want (from-env, true)", v, ok)
	}
	// Dotenv fallback
	if v, ok := combined("B"); !ok || v != "from-dotenv" {
		t.Errorf("combined(B) = (%q, %v), want (from-dotenv, true)", v, ok)
	}
	// Missing
	if v, ok := combined("C"); ok {
		t.Errorf("combined(C) = (%q, %v), want (, false)", v, ok)
	}
}
