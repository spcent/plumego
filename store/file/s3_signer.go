package file

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// S3Signer implements AWS Signature Version 4 signing.
type S3Signer struct {
	accessKey string
	secretKey string
	region    string
	service   string
}

// NewS3Signer creates a new S3 signer.
func NewS3Signer(accessKey, secretKey, region string) *S3Signer {
	return &S3Signer{
		accessKey: accessKey,
		secretKey: secretKey,
		region:    region,
		service:   "s3",
	}
}

// SignRequest signs an HTTP request using AWS Signature V4.
func (s *S3Signer) SignRequest(req *http.Request, payloadHash string) error {
	// Set timestamp
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	// If no payload hash provided, use default
	if payloadHash == "" {
		if req.Body != nil {
			payloadHash = "UNSIGNED-PAYLOAD"
		} else {
			payloadHash = emptyStringSHA256()
		}
	}

	// Set required headers
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	// Ensure Host is set
	if req.Host == "" {
		req.Host = req.URL.Host
	}

	// Step 1: Create canonical request
	canonicalRequest := s.buildCanonicalRequest(req, payloadHash)

	// Step 2: Create string to sign
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
	stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)

	// Step 3: Calculate signature
	signature := s.calculateSignature(dateStamp, stringToSign)

	// Step 4: Add signature to Authorization header
	authorization := s.buildAuthorizationHeader(signature, credentialScope, req)
	req.Header.Set("Authorization", authorization)

	return nil
}

// PresignRequest generates a presigned URL for the request.
func (s *S3Signer) PresignRequest(req *http.Request, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	expirySeconds := int(expiry.Seconds())
	if expirySeconds <= 0 || expirySeconds > 604800 { // Max 7 days
		expirySeconds = 900 // Default 15 minutes
	}

	// Build query parameters
	query := req.URL.Query()
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/%s/aws4_request",
		s.accessKey, dateStamp, s.region, s.service))
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", fmt.Sprintf("%d", expirySeconds))
	query.Set("X-Amz-SignedHeaders", "host")

	// Build canonical request
	req.URL.RawQuery = query.Encode()
	canonicalRequest := s.buildCanonicalRequestForPresign(req)

	// Build string to sign
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
	stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)

	// Calculate signature
	signature := s.calculateSignature(dateStamp, stringToSign)

	// Add signature to URL
	query.Set("X-Amz-Signature", signature)
	req.URL.RawQuery = query.Encode()

	return req.URL.String(), nil
}

// buildCanonicalRequest builds the canonical request string.
func (s *S3Signer) buildCanonicalRequest(req *http.Request, payloadHash string) string {
	// 1. HTTP method
	method := req.Method

	// 2. Canonical URI
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// 3. Canonical query string
	canonicalQueryString := s.buildCanonicalQueryString(req.URL.Query())

	// 4. Canonical headers
	canonicalHeaders, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)

	// 5. Combine canonical request
	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	return canonicalRequest
}

// buildCanonicalRequestForPresign builds canonical request for presigned URLs.
func (s *S3Signer) buildCanonicalRequestForPresign(req *http.Request) string {
	method := req.Method
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalQueryString := s.buildCanonicalQueryString(req.URL.Query())
	canonicalHeaders := "host:" + req.Host + "\n"
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := strings.Join([]string{
		method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	return canonicalRequest
}

// buildCanonicalQueryString builds the canonical query string.
func (s *S3Signer) buildCanonicalQueryString(values url.Values) string {
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		for _, v := range values[k] {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}

	return strings.Join(parts, "&")
}

// buildCanonicalHeaders builds the canonical headers string.
func (s *S3Signer) buildCanonicalHeaders(headers http.Header, host string) (string, string) {
	// Headers to sign
	var headerKeys []string
	headerMap := make(map[string]string)

	// Add required host header
	headerMap["host"] = host
	headerKeys = append(headerKeys, "host")

	// Add other headers
	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		if strings.HasPrefix(lowerKey, "x-amz-") {
			headerMap[lowerKey] = strings.Join(v, ",")
			headerKeys = append(headerKeys, lowerKey)
		}
	}

	sort.Strings(headerKeys)

	// Build canonical headers string
	var canonicalHeaders []string
	for _, k := range headerKeys {
		canonicalHeaders = append(canonicalHeaders, k+":"+headerMap[k])
	}

	canonicalHeadersStr := strings.Join(canonicalHeaders, "\n") + "\n"
	signedHeadersStr := strings.Join(headerKeys, ";")

	return canonicalHeadersStr, signedHeadersStr
}

// buildStringToSign builds the string to sign.
func (s *S3Signer) buildStringToSign(amzDate, credentialScope, canonicalRequest string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

// calculateSignature calculates the signature.
func (s *S3Signer) calculateSignature(dateStamp, stringToSign string) string {
	// Generate signing key
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte(s.service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))

	// Calculate signature
	signature := hmacSHA256(kSigning, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

// buildAuthorizationHeader builds the Authorization header value.
func (s *S3Signer) buildAuthorizationHeader(signature, credentialScope string, req *http.Request) string {
	_, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)

	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey,
		credentialScope,
		signedHeaders,
		signature,
	)
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// emptyStringSHA256 returns the SHA256 hash of an empty string.
func emptyStringSHA256() string {
	hash := sha256.Sum256([]byte{})
	return hex.EncodeToString(hash[:])
}
