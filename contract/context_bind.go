package contract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// BindJSON binds the request JSON body to the provided destination structure.
// It performs minimal decoding and returns a bindError on failure.
// An optional BindOptions value can tighten per-call JSON behavior.
func (c *Ctx) BindJSON(dst any, opts ...BindOptions) error {
	bindOpts, err := normalizeJSONBindOptions(opts)
	if err != nil {
		return err
	}

	data, err := c.bodyBytes()
	if err != nil {
		if errors.Is(err, ErrRequestBodyTooLarge) {
			// err IS ErrRequestBodyTooLarge (set directly by bodyBytes); pass through.
			return &bindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: err}
		}
		return &bindError{Status: http.StatusBadRequest, Message: "failed to read request body", Err: err}
	}

	if bindOpts.MaxBodySize > 0 && int64(len(data)) > bindOpts.MaxBodySize {
		return &bindError{Status: http.StatusRequestEntityTooLarge, Message: ErrRequestBodyTooLarge.Error(), Err: ErrRequestBodyTooLarge}
	}
	// bindOpts.MaxBodySize, if positive, enforces a stricter per-call cap on the
	// already-read body. RequestConfig.MaxBodySize enforces read-time limits.
	return decodeJSONBody(data, dst, bindOpts.DisallowUnknownFields)
}

// joinSentinel wraps sentinel and cause together so that errors.Is(e, sentinel)
// is true while the original cause is still reachable for logging and diagnosis.
// If cause is nil or identical to sentinel, sentinel is returned directly.
func joinSentinel(sentinel, cause error) error {
	if cause == nil || cause == sentinel {
		return sentinel
	}
	return fmt.Errorf("%w: %w", sentinel, cause)
}

func decodeJSONBody(data []byte, dst any, disallowUnknown bool) error {
	if len(bytes.TrimSpace(data)) == 0 {
		return &bindError{Status: http.StatusBadRequest, Message: ErrEmptyRequestBody.Error(), Err: ErrEmptyRequestBody}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if disallowUnknown {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(dst); err != nil {
		return &bindError{Status: http.StatusBadRequest, Message: ErrInvalidJSON.Error(), Err: joinSentinel(ErrInvalidJSON, err)}
	}

	if decoder.Decode(&struct{}{}) != io.EOF {
		return &bindError{Status: http.StatusBadRequest, Message: ErrUnexpectedExtraData.Error(), Err: ErrUnexpectedExtraData}
	}

	return nil
}

// BindQuery binds URL query parameters to the provided struct using the "query" struct tag.
// It supports scalar primitives, pointer-to-scalar fields, and primitive slices.
// Fields without a "query" tag are skipped. A tag value of "-" also skips the field.
// Validation is an explicit second step via ValidateStruct.
func (c *Ctx) BindQuery(dst any) error {
	return bindQuery(c.Query, dst)
}

// bindQuery maps URL query values to struct fields using the "query" struct tag.
func bindQuery(values url.Values, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &bindError{Status: http.StatusBadRequest, Message: "bind destination must be a non-nil pointer to a struct"}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return &bindError{Status: http.StatusBadRequest, Message: "bind destination must be a pointer to a struct"}
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("query")
		if tag == "" || tag == "-" {
			continue
		}

		// query:"name,omitempty" is accepted for parity with common Go tags.
		// Query binding already behaves as omitempty-by-default for absent values.
		name := strings.SplitN(tag, ",", 2)[0]
		queryVal := values.Get(name)
		queryVals := values[name]

		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		if err := setFieldFromQuery(fv, queryVal, queryVals); err != nil {
			return &bindError{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("invalid query parameter %q: %v", name, err),
				Err:     err,
			}
		}
	}

	return nil
}

// setFieldFromQuery assigns a query value to a struct field based on its type.
func setFieldFromQuery(fv reflect.Value, val string, vals []string) error {
	if val == "" && len(vals) == 0 {
		return nil
	}

	switch fv.Kind() {
	case reflect.Ptr:
		if fv.Type().Elem().Kind() == reflect.Ptr {
			return nil
		}
		elem := reflect.New(fv.Type().Elem()).Elem()
		if err := setFieldFromQuery(elem, val, vals); err != nil {
			return err
		}
		ptr := reflect.New(fv.Type().Elem())
		ptr.Elem().Set(elem)
		fv.Set(ptr)
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrInvalidParam, err)
		}
		fv.SetBool(b)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			fv.Set(reflect.ValueOf(append([]string(nil), vals...)))
			return nil
		}
		result := reflect.MakeSlice(fv.Type(), 0, len(vals))
		for _, item := range vals {
			elem := reflect.New(fv.Type().Elem()).Elem()
			if err := setFieldFromQuery(elem, item, []string{item}); err != nil {
				return err
			}
			result = reflect.Append(result, elem)
		}
		fv.Set(result)
	}
	return nil
}

// FormFile returns the first file for the provided form key.
// It is a convenience wrapper around http.Request.FormFile.
func (c *Ctx) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.R.FormFile(name)
	return fh, err
}

// SaveUploadedFile copies an uploaded file to a destination path on disk.
// Parent directories are created automatically if they do not exist.
func (c *Ctx) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if dir := filepath.Dir(dst); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

func normalizeJSONBindOptions(opts []BindOptions) (BindOptions, error) {
	switch len(opts) {
	case 0:
		return BindOptions{}, nil
	case 1:
		return opts[0], nil
	default:
		return BindOptions{}, &bindError{
			Status:  http.StatusBadRequest,
			Message: "BindJSON accepts at most one BindOptions value",
		}
	}
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
