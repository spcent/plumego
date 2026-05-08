package kvengine

import (
	"bytes"
	"strings"
	"testing"
)

func TestDetectFormatUnknownFailsClosed(t *testing.T) {
	_, err := DetectFormat(bytes.NewReader([]byte("???? snapshot")))
	if err == nil || !strings.Contains(err.Error(), "unknown snapshot format") {
		t.Fatalf("DetectFormat error = %v, want unknown snapshot format", err)
	}
}

func TestDetectWALFormatUnknownFailsClosed(t *testing.T) {
	_, err := DetectWALFormat(bytes.NewReader([]byte("???? wal")))
	if err == nil || !strings.Contains(err.Error(), "unknown WAL format") {
		t.Fatalf("DetectWALFormat error = %v, want unknown WAL format", err)
	}
}
