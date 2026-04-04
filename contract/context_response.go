package contract

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	payload := NewErrorBuilder().
		Status(status).
		Category(category).
		Code(errCode).
		Message(message).
		TraceID(c.TraceID).
		Details(details).
		Build()
	return WriteError(c.W, c.R, payload)
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
	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

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

// SetCookie adds a Set-Cookie header using the provided cookie configuration.
// Use http.Cookie{} to specify all attributes including SameSite.
// If cookie.Path is empty it defaults to "/".
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	http.SetCookie(c.W, cookie)
}

// Cookie returns the named cookie from the request.
// If the cookie is not found, it returns http.ErrNoCookie.
func (c *Ctx) Cookie(name string) (string, error) {
	cookie, err := c.R.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
