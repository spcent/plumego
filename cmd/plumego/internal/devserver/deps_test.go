package devserver

import (
	"strings"
	"testing"
)

func TestDependencyDiagnosticBufferBoundsOutput(t *testing.T) {
	buf := newDependencyDiagnosticBuffer(8)
	n, err := buf.Write([]byte("1234567890"))
	if err != nil || n != 10 {
		t.Fatalf("write n=%d err=%v", n, err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "12345678") {
		t.Fatalf("expected retained prefix, got %q", out)
	}
	if !strings.Contains(out, "dependency diagnostic output truncated after 8 bytes") ||
		!strings.Contains(out, "2 bytes omitted") {
		t.Fatalf("expected truncation marker, got %q", out)
	}
}
