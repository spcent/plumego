package devserver

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	defaultAPITestTimeout = 5 * time.Second
	maxAPITestBodyBytes   = 128 * 1024
)

type APITestRequest struct {
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     string            `json:"query,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      string            `json:"body,omitempty"`
	TimeoutMS int               `json:"timeout_ms,omitempty"`
}

type APITestResponse struct {
	Success       bool              `json:"success"`
	Status        int               `json:"status"`
	DurationMS    int64             `json:"duration_ms"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	BodyBase64    string            `json:"body_base64,omitempty"`
	BodyEncoding  string            `json:"body_encoding,omitempty"`
	BodyTruncated bool              `json:"body_truncated,omitempty"`
	Bytes         int               `json:"bytes"`
}

func (a *Analyzer) DoAPITest(req APITestRequest) (APITestResponse, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	if !isAllowedMethod(method) {
		return APITestResponse{}, fmt.Errorf("unsupported method: %s", method)
	}

	targetURL, err := buildAppURL(a.appURL, req.Path, req.Query)
	if err != nil {
		return APITestResponse{}, err
	}

	var body io.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	}

	timeout := defaultAPITestTimeout
	if req.TimeoutMS > 0 {
		timeout = time.Duration(req.TimeoutMS) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return APITestResponse{}, err
	}

	for key, value := range req.Headers {
		if strings.TrimSpace(key) == "" {
			continue
		}
		httpReq.Header.Set(key, value)
	}

	if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
		trimmed := strings.TrimSpace(req.Body)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			httpReq.Header.Set("Content-Type", "application/json")
		}
	}

	start := time.Now()
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return APITestResponse{}, err
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, maxAPITestBodyBytes+1)
	payload, err := io.ReadAll(limited)
	if err != nil {
		return APITestResponse{}, err
	}

	truncated := false
	if len(payload) > maxAPITestBodyBytes {
		truncated = true
		payload = payload[:maxAPITestBodyBytes]
	}

	headers := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	out := APITestResponse{
		Success:       true,
		Status:        resp.StatusCode,
		DurationMS:    time.Since(start).Milliseconds(),
		Headers:       headers,
		BodyTruncated: truncated,
		Bytes:         len(payload),
	}

	if len(payload) == 0 {
		return out, nil
	}

	if utf8.Valid(payload) {
		out.Body = string(bytes.TrimSuffix(payload, []byte{0}))
		return out, nil
	}

	out.BodyBase64 = base64.StdEncoding.EncodeToString(payload)
	out.BodyEncoding = "base64"
	return out, nil
}

func buildAppURL(base, path, rawQuery string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid app URL")
	}

	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		trimmed = "/"
	}

	if strings.Contains(trimmed, "://") {
		return "", fmt.Errorf("absolute URLs are not allowed")
	}

	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}

	parsedPath, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid path")
	}

	query := parsedPath.Query()
	if rawQuery != "" {
		extra, err := url.ParseQuery(rawQuery)
		if err != nil {
			return "", fmt.Errorf("invalid query")
		}
		for key, values := range extra {
			for _, value := range values {
				query.Add(key, value)
			}
		}
	}

	parsedPath.RawQuery = query.Encode()

	finalURL := baseURL.ResolveReference(parsedPath)
	return finalURL.String(), nil
}

func isAllowedMethod(method string) bool {
	switch method {
	case http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions:
		return true
	default:
		return false
	}
}
