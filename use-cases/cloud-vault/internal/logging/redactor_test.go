package logging

import (
	"bytes"
	"testing"
)

func TestRedactingWriter_BearerToken(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ"
	_, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if output == input {
		t.Errorf("Bearer token was not redacted")
	}
	if !bytes.Contains([]byte(output), []byte("Bearer <redacted>")) {
		t.Errorf("Expected 'Bearer <redacted>', got: %s", output)
	}
}

func TestRedactingWriter_APIKey(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "api_key=sk_test_1234567890abcdef"
	_, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if output == input {
		t.Errorf("API key was not redacted")
	}
	if !bytes.Contains([]byte(output), []byte("<redacted>")) {
		t.Errorf("Expected redacted output, got: %s", output)
	}
}

func TestRedactingWriter_Password(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "password=mysecretpassword123"
	_, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if output == input {
		t.Errorf("Password was not redacted")
	}
	if !bytes.Contains([]byte(output), []byte("<redacted>")) {
		t.Errorf("Expected redacted output, got: %s", output)
	}
}

func TestRedactingWriter_MultiplePatterns(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "User login with api_key=abc123 and password=secret456"
	_, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	// Both should be redacted
	if bytes.Contains([]byte(output), []byte("abc123")) {
		t.Errorf("API key was not redacted: %s", output)
	}
	if bytes.Contains([]byte(output), []byte("secret456")) {
		t.Errorf("Password was not redacted: %s", output)
	}
}

func TestRedactingWriter_NoSensitiveData(t *testing.T) {
	var buf bytes.Buffer
	rw := NewRedactingWriter(&buf)

	input := "INFO: User logged in successfully"
	_, err := rw.Write([]byte(input))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := buf.String()
	if output != input {
		t.Errorf("Non-sensitive data should pass through unchanged, got: %s", output)
	}
}
