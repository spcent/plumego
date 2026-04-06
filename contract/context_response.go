package contract

import (
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

// Response writes a standardized success response that includes trace id when available.
func (c *Ctx) Response(status int, data any, meta map[string]any) error {
	if c == nil {
		return ErrContextNil
	}
	return WriteResponse(c.W, c.R, status, data, meta)
}

// Text writes a plain text response with the given status code.
func (c *Ctx) Text(status int, text string) error {
	if c == nil {
		return ErrContextNil
	}
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.WriteHeader(status)
	_, err := io.WriteString(c.W, text)
	return err
}

// Bytes writes a binary response with the given status code.
func (c *Ctx) Bytes(status int, data []byte) error {
	if c == nil {
		return ErrContextNil
	}
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.WriteHeader(status)
	_, err := c.W.Write(data)
	return err
}

// Redirect sends a redirect response but only allows relative paths and
// same-origin URLs. It returns ErrUnsafeRedirect if the location points to
// an external host, preventing open redirect vulnerabilities.
// Use UnsafeRedirect when a cross-origin redirect is explicitly required.
func (c *Ctx) Redirect(status int, location string) error {
	if c == nil {
		return ErrContextNil
	}
	if err := validateRedirectURL(location, c.R); err != nil {
		return err
	}
	http.Redirect(c.W, c.R, location, status)
	return nil
}

// UnsafeRedirect sends a redirect response to any URL, including external hosts.
// Prefer Redirect for same-origin redirects. Only use this when you explicitly
// need to redirect to a different origin and have validated the destination.
func (c *Ctx) UnsafeRedirect(status int, location string) error {
	if c == nil {
		return ErrContextNil
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
	if r != nil {
		requestHost := r.Host
		if requestHost == "" && r.URL != nil {
			requestHost = r.URL.Host
		}
		if strings.EqualFold(parsed.Host, requestHost) {
			return nil
		}
	}

	return ErrUnsafeRedirect
}

// File serves a file to the client.
func (c *Ctx) File(path string) error {
	if c == nil {
		return ErrContextNil
	}
	http.ServeFile(c.W, c.R, path)
	return nil
}

// SetCookie adds a Set-Cookie header using the provided cookie configuration.
// Use http.Cookie{} to specify all attributes including SameSite.
// If cookie.Path is empty it defaults to "/".
// SetCookie does not modify the caller's cookie struct.
func (c *Ctx) SetCookie(cookie *http.Cookie) {
	if c == nil {
		return
	}
	local := *cookie // copy; do not mutate caller's struct
	if local.Path == "" {
		local.Path = "/"
	}
	// Ensure cookies are only sent over secure HTTPS connections.
	local.Secure = true
	http.SetCookie(c.W, &local)
}

// Cookie returns the named cookie from the request.
// If the cookie is not found, it returns http.ErrNoCookie.
func (c *Ctx) Cookie(name string) (string, error) {
	if c == nil {
		return "", ErrContextNil
	}
	cookie, err := c.R.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}
