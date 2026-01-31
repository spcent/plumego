package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/spcent/plumego/validator"
)

// BindJSON binds the request JSON body to the provided destination structure.
// It performs minimal validation and returns a BindError on failure.
func (c *Ctx) BindJSON(dst any) error {
	data, err := c.bodyBytes()
	if err != nil {
		if errors.Is(err, ErrRequestBodyTooLarge) {
			return &BindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
		return &BindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return &BindError{Status: http.StatusBadRequest, Message: "request body is empty"}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	// DisallowUnknownFields could be enabled if you want strict mode:
	// decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: "invalid JSON payload", Err: err}
	}

	// Ensure no trailing data
	if decoder.Decode(&struct{}{}) != io.EOF {
		return &BindError{Status: http.StatusBadRequest, Message: "unexpected extra JSON data"}
	}

	return nil
}

// BindAndValidateJSON binds the request body to dst and validates it using struct tags.
func (c *Ctx) BindAndValidateJSON(dst any) error {
	if err := c.BindJSON(dst); err != nil {
		return err
	}

	if err := validator.Validate(dst); err != nil {
		return &BindError{Status: http.StatusBadRequest, Message: err.Error(), Err: err}
	}

	return nil
}

func (c *Ctx) bodyBytes() ([]byte, error) {
	c.bodyReadOnce.Do(func() {
		reader := io.Reader(c.R.Body)
		maxBodySize := int64(0)
		if c.Config != nil && c.Config.MaxBodySize > 0 {
			maxBodySize = c.Config.MaxBodySize
			if c.W != nil {
				reader = http.MaxBytesReader(c.W, c.R.Body, maxBodySize)
			} else {
				reader = io.LimitReader(reader, maxBodySize+1)
			}
		}

		c.body, c.bodyErr = io.ReadAll(reader)
		if c.bodyErr != nil {
			var maxErr *http.MaxBytesError
			if errors.As(c.bodyErr, &maxErr) {
				c.bodyErr = ErrRequestBodyTooLarge
				c.body = nil
			}
			return
		}
		if maxBodySize > 0 && int64(len(c.body)) > maxBodySize {
			c.bodyErr = ErrRequestBodyTooLarge
			c.body = nil
			return
		}
		c.bodySize.Store(int64(len(c.body)))
		if c.Config == nil || c.Config.EnableBodyCache {
			c.R.Body = io.NopCloser(bytes.NewBuffer(c.body))
		}
	})
	return c.body, c.bodyErr
}
