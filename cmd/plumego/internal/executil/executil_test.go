package executil

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestRunBoundsOutput(t *testing.T) {
	result, err := Run(context.Background(), Options{
		Name:        os.Args[0],
		Args:        []string{"-test.run=TestHelperProcess"},
		Env:         []string{"PLUMEGO_EXECUTIL_HELPER=1"},
		Timeout:     2 * time.Second,
		OutputLimit: 8,
	})
	if err != nil {
		t.Fatalf("run helper: %v", err)
	}
	if result.Stdout != "12345678\n[plumego: output truncated after 8 bytes; 8 bytes omitted]\n" {
		t.Fatalf("unexpected stdout: %q", result.Stdout)
	}
	if !result.StdoutTruncated || result.StderrTruncated {
		t.Fatalf("unexpected truncation flags: %#v", result)
	}
}

func TestRunTimesOut(t *testing.T) {
	result, err := Run(context.Background(), Options{
		Name:    os.Args[0],
		Args:    []string{"-test.run=TestSleepHelperProcess"},
		Env:     []string{"PLUMEGO_EXECUTIL_SLEEP_HELPER=1"},
		Timeout: 10 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !result.TimedOut {
		t.Fatalf("expected timeout flag, got %#v", result)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("PLUMEGO_EXECUTIL_HELPER") != "1" {
		return
	}
	fmt.Print("1234567890abcdef")
	os.Exit(0)
}

func TestSleepHelperProcess(t *testing.T) {
	if os.Getenv("PLUMEGO_EXECUTIL_SLEEP_HELPER") != "1" {
		return
	}
	time.Sleep(time.Second)
	os.Exit(0)
}
