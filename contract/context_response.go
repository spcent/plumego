package contract

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spcent/plumego/utils/pool"
)

var (
	// ErrUnsafeRedirect is returned when a redirect URL points to an external host.
	ErrUnsafeRedirect = errors.New("unsafe redirect: external URLs are not allowed")
)

func (c *Ctx) ErrorJSON(status int, errCode string, message string, details map[string]any) error {
	category := CategoryForStatus(status)
	if category == "" {
		category = CategoryBusiness
	}
	payload := APIError{
		Status:   status,
		Code:     errCode,
		Message:  message,
		Details:  details,
		TraceID:  c.TraceID,
		Category: category,
	}
	return c.JSON(status, payload)
}

// Response writes a standardized success response that includes trace id when available.
func (c *Ctx) Response(status int, data any, meta map[string]any) error {
	if c == nil {
		return ErrContextNil
	}
	return WriteResponse(c.W, c.R, status, data, meta)
}

// JSON writes a JSON response with the given status code.
func (c *Ctx) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/json")
	c.W.WriteHeader(status)

	// Use pooled buffer for encoding to reduce allocations
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return err
	}

	// Write the buffer content to response writer
	_, err := c.W.Write(buf.Bytes())
	return err
}

// Text writes a plain text response with the given status code.
func (c *Ctx) Text(status int, text string) error {
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	_, err := io.WriteString(c.W, text)
	return err
}

// Bytes writes a binary response with the given status code.
func (c *Ctx) Bytes(status int, data []byte) error {
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.WriteHeader(status)
	_, err := c.W.Write(data)
	return err
}

// Redirect sends a redirect response to the client.
// Note: This accepts any URL including external ones. Use SafeRedirect to
// restrict redirects to same-origin paths only.
func (c *Ctx) Redirect(status int, location string) error {
	http.Redirect(c.W, c.R, location, status)
	return nil
}

// SafeRedirect sends a redirect response but only allows relative paths and
// same-origin URLs. It returns ErrUnsafeRedirect if the location points to
// an external host, preventing open redirect vulnerabilities.
func (c *Ctx) SafeRedirect(status int, location string) error {
	if err := validateRedirectURL(location, c.R); err != nil {
		return err
	}
	http.Redirect(c.W, c.R, location, status)
	return nil
}

// validateRedirectURL checks that a redirect URL is safe (relative or same-origin).
func validateRedirectURL(location string, r *http.Request) error {
	// Reject protocol-relative URLs (e.g. "//evil.com/path")
	if strings.HasPrefix(location, "//") {
		return ErrUnsafeRedirect
	}

	parsed, err := url.Parse(location)
	if err != nil {
		return ErrUnsafeRedirect
	}

	// Relative URLs (no scheme and no host) are safe
	if parsed.Scheme == "" && parsed.Host == "" {
		return nil
	}

	// Absolute URL: check that the host matches the request host
	if r != nil && r.Host != "" {
		requestHost := r.Host
		if strings.EqualFold(parsed.Host, requestHost) {
			return nil
		}
	}

	return ErrUnsafeRedirect
}

// File serves a file to the client.
func (c *Ctx) File(path string) error {
	http.ServeFile(c.W, c.R, path)
	return nil
}
