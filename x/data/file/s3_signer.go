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
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	if payloadHash == "" {
		if req.Body != nil {
			payloadHash = "UNSIGNED-PAYLOAD"
		} else {
			payloadHash = emptyStringSHA256()
		}
	}

	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	if req.Host == "" {
		req.Host = req.URL.Host
	}

	canonicalRequest := s.buildCanonicalRequest(req, payloadHash)
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
	stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)
	signature := s.calculateSignature(dateStamp, stringToSign)
	req.Header.Set("Authorization", s.buildAuthorizationHeader(signature, credentialScope, req))

	return nil
}

// PresignRequest generates a presigned URL for the request.
func (s *S3Signer) PresignRequest(req *http.Request, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	expirySeconds := int(expiry.Seconds())
	if expirySeconds <= 0 || expirySeconds > 604800 {
		expirySeconds = 900
	}

	query := req.URL.Query()
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/%s/aws4_request",
		s.accessKey, dateStamp, s.region, s.service))
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", fmt.Sprintf("%d", expirySeconds))
	query.Set("X-Amz-SignedHeaders", "host")

	req.URL.RawQuery = query.Encode()
	canonicalRequest := s.buildCanonicalRequestForPresign(req)

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s.region, s.service)
	stringToSign := s.buildStringToSign(amzDate, credentialScope, canonicalRequest)
	signature := s.calculateSignature(dateStamp, stringToSign)

	query.Set("X-Amz-Signature", signature)
	req.URL.RawQuery = query.Encode()

	return req.URL.String(), nil
}

func (s *S3Signer) buildCanonicalRequest(req *http.Request, payloadHash string) string {
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	canonicalHeaders, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)

	return strings.Join([]string{
		req.Method,
		canonicalURI,
		s.buildCanonicalQueryString(req.URL.Query()),
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
}

func (s *S3Signer) buildCanonicalRequestForPresign(req *http.Request) string {
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	return strings.Join([]string{
		req.Method,
		canonicalURI,
		s.buildCanonicalQueryString(req.URL.Query()),
		"host:" + req.Host + "\n",
		"host",
		"UNSIGNED-PAYLOAD",
	}, "\n")
}

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

func (s *S3Signer) buildCanonicalHeaders(headers http.Header, host string) (string, string) {
	headerMap := map[string]string{"host": host}
	headerKeys := []string{"host"}

	for k, v := range headers {
		lowerKey := strings.ToLower(k)
		if strings.HasPrefix(lowerKey, "x-amz-") {
			headerMap[lowerKey] = strings.Join(v, ",")
			headerKeys = append(headerKeys, lowerKey)
		}
	}

	sort.Strings(headerKeys)

	var canonical []string
	for _, k := range headerKeys {
		canonical = append(canonical, k+":"+headerMap[k])
	}

	return strings.Join(canonical, "\n") + "\n", strings.Join(headerKeys, ";")
}

func (s *S3Signer) buildStringToSign(amzDate, credentialScope, canonicalRequest string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

func (s *S3Signer) calculateSignature(dateStamp, stringToSign string) string {
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(s.region))
	kService := hmacSHA256(kRegion, []byte(s.service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return hex.EncodeToString(hmacSHA256(kSigning, []byte(stringToSign)))
}

func (s *S3Signer) buildAuthorizationHeader(signature, credentialScope string, req *http.Request) string {
	_, signedHeaders := s.buildCanonicalHeaders(req.Header, req.Host)
	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		s.accessKey, credentialScope, signedHeaders, signature)
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func emptyStringSHA256() string {
	hash := sha256.Sum256([]byte{})
	return hex.EncodeToString(hash[:])
}
