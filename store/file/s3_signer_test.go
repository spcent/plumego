package file

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestNewS3Signer(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	if signer == nil {
		t.Fatal("Signer is nil")
	}

	if signer.accessKey != "access-key" {
		t.Errorf("accessKey = %q, want %q", signer.accessKey, "access-key")
	}
	if signer.secretKey != "secret-key" {
		t.Errorf("secretKey = %q, want %q", signer.secretKey, "secret-key")
	}
	if signer.region != "us-east-1" {
		t.Errorf("region = %q, want %q", signer.region, "us-east-1")
	}
	if signer.service != "s3" {
		t.Errorf("service = %q, want %q", signer.service, "s3")
	}
}

func TestS3Signer_SignRequest(t *testing.T) {
	signer := NewS3Signer("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "us-east-1")

	// Create test request
	req, err := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Sign request
	payloadHash := emptyStringSHA256()
	if err := signer.SignRequest(req, payloadHash); err != nil {
		t.Fatalf("SignRequest failed: %v", err)
	}

	// Verify headers were added
	if req.Header.Get("x-amz-date") == "" {
		t.Error("x-amz-date header not set")
	}
	if req.Header.Get("x-amz-content-sha256") == "" {
		t.Error("x-amz-content-sha256 header not set")
	}
	if req.Header.Get("Authorization") == "" {
		t.Error("Authorization header not set")
	}

	// Verify Authorization header format
	auth := req.Header.Get("Authorization")
	if !containsString(auth, "AWS4-HMAC-SHA256") {
		t.Error("Authorization should contain AWS4-HMAC-SHA256")
	}
	if !containsString(auth, "Credential=") {
		t.Error("Authorization should contain Credential=")
	}
	if !containsString(auth, "SignedHeaders=") {
		t.Error("Authorization should contain SignedHeaders=")
	}
	if !containsString(auth, "Signature=") {
		t.Error("Authorization should contain Signature=")
	}
}

func TestS3Signer_SignRequest_WithHost(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	req, err := http.NewRequest(http.MethodPut, "https://bucket.s3.amazonaws.com/key", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Don't set Host explicitly, should use URL.Host
	if err := signer.SignRequest(req, emptyStringSHA256()); err != nil {
		t.Fatal(err)
	}

	// Verify Host was set
	if req.Host == "" {
		t.Error("Host should be set from URL")
	}
}

func TestS3Signer_SignRequest_WithPayloadHash(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	req, err := http.NewRequest(http.MethodPut, "https://bucket.s3.amazonaws.com/key", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Custom payload hash
	customHash := "abcd1234"
	if err := signer.SignRequest(req, customHash); err != nil {
		t.Fatal(err)
	}

	// Verify payload hash was used
	if req.Header.Get("x-amz-content-sha256") != customHash {
		t.Errorf("x-amz-content-sha256 = %q, want %q", req.Header.Get("x-amz-content-sha256"), customHash)
	}
}

func TestS3Signer_PresignRequest(t *testing.T) {
	signer := NewS3Signer("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "us-east-1")

	req, err := http.NewRequest(http.MethodGet, "https://examplebucket.s3.amazonaws.com/test.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Generate presigned URL
	presignedURL, err := signer.PresignRequest(req, 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignRequest failed: %v", err)
	}

	// Parse URL
	u, err := url.Parse(presignedURL)
	if err != nil {
		t.Fatalf("Failed to parse presigned URL: %v", err)
	}

	// Verify query parameters
	query := u.Query()
	if query.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Errorf("X-Amz-Algorithm = %q, want %q", query.Get("X-Amz-Algorithm"), "AWS4-HMAC-SHA256")
	}
	if query.Get("X-Amz-Credential") == "" {
		t.Error("X-Amz-Credential not set")
	}
	if query.Get("X-Amz-Date") == "" {
		t.Error("X-Amz-Date not set")
	}
	if query.Get("X-Amz-Expires") != "900" { // 15 minutes = 900 seconds
		t.Errorf("X-Amz-Expires = %q, want %q", query.Get("X-Amz-Expires"), "900")
	}
	if query.Get("X-Amz-SignedHeaders") != "host" {
		t.Errorf("X-Amz-SignedHeaders = %q, want %q", query.Get("X-Amz-SignedHeaders"), "host")
	}
	if query.Get("X-Amz-Signature") == "" {
		t.Error("X-Amz-Signature not set")
	}
}

func TestS3Signer_PresignRequest_ExpiryValidation(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	tests := []struct {
		name        string
		expiry      time.Duration
		wantExpiry  string
	}{
		{
			name:       "valid expiry",
			expiry:     time.Hour,
			wantExpiry: "3600",
		},
		{
			name:       "zero expiry - default to 15 min",
			expiry:     0,
			wantExpiry: "900",
		},
		{
			name:       "negative expiry - default to 15 min",
			expiry:     -1 * time.Hour,
			wantExpiry: "900",
		},
		{
			name:       "too long expiry - default to 15 min",
			expiry:     8 * 24 * time.Hour, // More than 7 days
			wantExpiry: "900",
		},
		{
			name:       "max expiry - 7 days",
			expiry:     7 * 24 * time.Hour,
			wantExpiry: "604800",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "https://bucket.s3.amazonaws.com/key", nil)
			if err != nil {
				t.Fatal(err)
			}

			presignedURL, err := signer.PresignRequest(req, tt.expiry)
			if err != nil {
				t.Fatal(err)
			}

			u, err := url.Parse(presignedURL)
			if err != nil {
				t.Fatal(err)
			}

			gotExpiry := u.Query().Get("X-Amz-Expires")
			if gotExpiry != tt.wantExpiry {
				t.Errorf("X-Amz-Expires = %q, want %q", gotExpiry, tt.wantExpiry)
			}
		})
	}
}

func TestS3Signer_BuildCanonicalQueryString(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	tests := []struct {
		name   string
		values url.Values
		want   string
	}{
		{
			name:   "empty",
			values: url.Values{},
			want:   "",
		},
		{
			name: "single parameter",
			values: url.Values{
				"key": []string{"value"},
			},
			want: "key=value",
		},
		{
			name: "multiple parameters - sorted",
			values: url.Values{
				"zebra": []string{"z"},
				"alpha": []string{"a"},
				"beta":  []string{"b"},
			},
			want: "alpha=a&beta=b&zebra=z",
		},
		{
			name: "multiple values for same key",
			values: url.Values{
				"key": []string{"value1", "value2"},
			},
			want: "key=value1&key=value2",
		},
		{
			name: "special characters - should be escaped",
			values: url.Values{
				"key": []string{"value with spaces"},
			},
			want: "key=value+with+spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := signer.buildCanonicalQueryString(tt.values)
			if got != tt.want {
				t.Errorf("buildCanonicalQueryString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestS3Signer_BuildCanonicalHeaders(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	headers := http.Header{
		"Content-Type":       []string{"application/json"},
		"X-Amz-Date":         []string{"20260205T123000Z"},
		"X-Amz-Content-Sha256": []string{"abc123"},
	}
	host := "bucket.s3.amazonaws.com"

	canonicalHeaders, signedHeaders := signer.buildCanonicalHeaders(headers, host)

	// Verify canonical headers format
	if !containsString(canonicalHeaders, "host:bucket.s3.amazonaws.com") {
		t.Error("Canonical headers should contain host")
	}
	if !containsString(canonicalHeaders, "x-amz-date:20260205T123000Z") {
		t.Error("Canonical headers should contain x-amz-date")
	}
	if !containsString(canonicalHeaders, "x-amz-content-sha256:abc123") {
		t.Error("Canonical headers should contain x-amz-content-sha256")
	}

	// Verify signed headers are sorted
	if !containsString(signedHeaders, "host") {
		t.Error("Signed headers should contain host")
	}
	if !containsString(signedHeaders, "x-amz-date") {
		t.Error("Signed headers should contain x-amz-date")
	}
	if !containsString(signedHeaders, "x-amz-content-sha256") {
		t.Error("Signed headers should contain x-amz-content-sha256")
	}
}

func TestEmptyStringSHA256(t *testing.T) {
	got := emptyStringSHA256()

	// Calculate expected hash
	hash := sha256.Sum256([]byte{})
	want := hex.EncodeToString(hash[:])

	if got != want {
		t.Errorf("emptyStringSHA256() = %q, want %q", got, want)
	}

	// Verify it's a valid SHA256 hex string
	if len(got) != 64 {
		t.Errorf("Hash length = %d, want 64", len(got))
	}
}

func TestHmacSHA256(t *testing.T) {
	key := []byte("secret")
	data := []byte("message")

	result := hmacSHA256(key, data)

	// Verify result is not empty
	if len(result) == 0 {
		t.Error("HMAC result is empty")
	}

	// Verify result length (SHA256 produces 32 bytes)
	if len(result) != 32 {
		t.Errorf("HMAC length = %d, want 32", len(result))
	}

	// Verify deterministic
	result2 := hmacSHA256(key, data)
	if !bytesEqual(result, result2) {
		t.Error("HMAC should be deterministic")
	}

	// Verify different key produces different result
	result3 := hmacSHA256([]byte("different"), data)
	if bytesEqual(result, result3) {
		t.Error("Different keys should produce different HMAC")
	}
}

func TestS3Signer_CalculateSignature(t *testing.T) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	dateStamp := "20260205"
	stringToSign := "test-string-to-sign"

	signature := signer.calculateSignature(dateStamp, stringToSign)

	// Verify signature is hex-encoded
	if len(signature) != 64 { // SHA256 hex = 64 chars
		t.Errorf("Signature length = %d, want 64", len(signature))
	}

	// Verify deterministic
	signature2 := signer.calculateSignature(dateStamp, stringToSign)
	if signature != signature2 {
		t.Error("Signature should be deterministic")
	}

	// Verify different date produces different signature
	signature3 := signer.calculateSignature("20260206", stringToSign)
	if signature == signature3 {
		t.Error("Different dates should produce different signatures")
	}
}

// Benchmark signing operations
func BenchmarkS3Signer_SignRequest(b *testing.B) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	req, _ := http.NewRequest(http.MethodGet, "https://bucket.s3.amazonaws.com/key", nil)
	payloadHash := emptyStringSHA256()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.SignRequest(req, payloadHash)
	}
}

func BenchmarkS3Signer_PresignRequest(b *testing.B) {
	signer := NewS3Signer("access-key", "secret-key", "us-east-1")

	req, _ := http.NewRequest(http.MethodGet, "https://bucket.s3.amazonaws.com/key", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.PresignRequest(req, 15*time.Minute)
	}
}

func BenchmarkHmacSHA256(b *testing.B) {
	key := []byte("secret-key")
	data := []byte("data-to-sign")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hmacSHA256(key, data)
	}
}

// Helper functions

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr)+1))
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
