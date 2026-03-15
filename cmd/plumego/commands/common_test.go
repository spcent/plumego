package commands

import (
	"flag"
	"testing"
)

func TestContainsHelpFlag(t *testing.T) {
	// Test case 1: No help flag
	args := []string{"--flag", "value"}
	if containsHelpFlag(args) {
		t.Fatalf("expected no help flag, got true")
	}

	// Test case 2: Short help flag
	args = []string{"-h"}
	if !containsHelpFlag(args) {
		t.Fatalf("expected help flag, got false")
	}

	// Test case 3: Long help flag
	args = []string{"--help"}
	if !containsHelpFlag(args) {
		t.Fatalf("expected help flag, got false")
	}

	// Test case 4: Help flag with other flags
	args = []string{"--flag", "value", "--help"}
	if !containsHelpFlag(args) {
		t.Fatalf("expected help flag, got false")
	}
}

func TestParseFlags(t *testing.T) {
	// Test case: Parse flags and return positionals
	var flagValue string
	var boolValue bool

	args := []string{"--flag", "test", "pos1", "pos2"}
	positionals, err := ParseFlags("test", args, func(fs *flag.FlagSet) {
		flagValue = *fs.String("flag", "default", "test flag")
		boolValue = *fs.Bool("bool", false, "test bool")
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if flagValue != "test" {
		t.Fatalf("expected flag value 'test', got: %s", flagValue)
	}

	if boolValue != false {
		t.Fatalf("expected bool value false, got: %t", boolValue)
	}

	if len(positionals) != 2 {
		t.Fatalf("expected 2 positionals, got: %d", len(positionals))
	}

	if positionals[0] != "pos1" || positionals[1] != "pos2" {
		t.Fatalf("expected positionals ['pos1', 'pos2'], got: %v", positionals)
	}
}

func TestNewAppError(t *testing.T) {
	// Test case: Create new app error
	err := NewAppError(1, "test error", "test detail")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "test error" {
		t.Fatalf("expected error message 'test error', got: %s", err.Error())
	}

	if err.Code() != 1 {
		t.Fatalf("expected error code 1, got: %d", err.Code())
	}
}
