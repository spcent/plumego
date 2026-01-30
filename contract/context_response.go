package contract

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spcent/plumego/utils/pool"
)

func (c *Ctx) ErrorJSON(status int, errCode string, message string, details map[string]any) error {
	payload := APIError{
		Status:   status,
		Code:     errCode,
		Message:  message,
		Details:  details,
		TraceID:  c.TraceID,
		Category: CategoryBusiness,
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
func (c *Ctx) Redirect(status int, location string) error {
	http.Redirect(c.W, c.R, location, status)
	return nil
}

// File serves a file to the client.
func (c *Ctx) File(path string) error {
	http.ServeFile(c.W, c.R, path)
	return nil
}
