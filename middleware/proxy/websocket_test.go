package proxy

import "testing"

func TestComputeAcceptKey(t *testing.T) {
	key := "dGhlIHNhbXBsZSBub25jZQ=="
	want := "s3pPLMBiTxaQ9kYGzzhZRbK+xOo="

	if got := computeAcceptKey(key); got != want {
		t.Fatalf("computeAcceptKey mismatch: got %q want %q", got, want)
	}
}
