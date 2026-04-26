package contract

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

type bindJSONPayload struct {
	Name string `json:"name"`
}

type countingReadCloser struct {
	reader *strings.Reader
	reads  int
}

func newCountingReadCloser(body string) *countingReadCloser {
	return &countingReadCloser{reader: strings.NewReader(body)}
}

func (r *countingReadCloser) Read(p []byte) (int, error) {
	r.reads++
	return r.reader.Read(p)
}

func (r *countingReadCloser) Close() error {
	return nil
}

func TestNewCtxPopulatesStableFields(t *testing.T) {
	deadline := time.Now().Add(time.Minute)
	baseCtx, cancel := context.WithDeadline(t.Context(), deadline)
	defer cancel()

	req := httptest.NewRequest(http.MethodGet, "/users/123?foo=bar", nil).WithContext(baseCtx)
	ctx := NewCtx(httptest.NewRecorder(), req, map[string]string{"id": "123"})

	if ctx.W == nil || ctx.R == nil {
		t.Fatalf("expected net/http primitives to be retained")
	}
	if ctx.Params["id"] != "123" {
		t.Fatalf("expected param to be propagated")
	}
	if got := ctx.R.URL.Query().Get("foo"); got != "bar" {
		t.Fatalf("expected query to remain on request URL, got %q", got)
	}
	gotDeadline, ok := ctx.R.Context().Deadline()
	if !ok || !gotDeadline.Equal(deadline) {
		t.Fatalf("expected existing request deadline to be preserved")
	}
	if ctx.RequestID() != "" {
		t.Fatalf("expected empty request id on request without request context, got %q", ctx.RequestID())
	}
}

func TestBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"demo"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	var payload struct {
		Name string `json:"name"`
	}
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}
	if payload.Name != "demo" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestBindJSONMaxBodySize(t *testing.T) {
	body := bytes.NewBufferString(`{"name":"too-long"}`)
	ctx := NewCtxWithConfig(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil, RequestConfig{
		MaxBodySize:     10,
		EnableBodyCache: true,
	})

	var payload bindJSONPayload
	err := ctx.BindJSON(&payload)
	if err == nil {
		t.Fatalf("expected error for body too large")
	}

	var bindErr *bindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected bindError, got %T", err)
	}
	if bindErr.Status != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, bindErr.Status)
	}
}

func TestBindJSONRejectsNegativeRequestMaxBodySize(t *testing.T) {
	body := newCountingReadCloser(`{"name":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = body
	ctx := NewCtxWithConfig(httptest.NewRecorder(), req, nil, RequestConfig{
		MaxBodySize:     -1,
		EnableBodyCache: true,
	})

	var payload bindJSONPayload
	err := ctx.BindJSON(&payload)
	if err == nil {
		t.Fatal("expected invalid max body size error")
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected errors.Is(err, ErrInvalidParam) to be true, got %v", err)
	}
	if body.reads != 0 {
		t.Fatalf("expected body not to be read, got %d reads", body.reads)
	}
}

func TestBindJSONRejectsNegativeOptionMaxBodySize(t *testing.T) {
	body := newCountingReadCloser(`{"name":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = body
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var payload bindJSONPayload
	err := ctx.BindJSON(&payload, BindOptions{MaxBodySize: -1})
	if err == nil {
		t.Fatal("expected invalid max body size error")
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected errors.Is(err, ErrInvalidParam) to be true, got %v", err)
	}
	if body.reads != 0 {
		t.Fatalf("expected body not to be read, got %d reads", body.reads)
	}
}

func TestBindJSONRejectsMultipleOptions(t *testing.T) {
	body := newCountingReadCloser(`{"name":"demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = body
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var payload bindJSONPayload
	err := ctx.BindJSON(&payload, BindOptions{}, BindOptions{DisallowUnknownFields: true})
	if err == nil {
		t.Fatal("expected multiple bind options error")
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected errors.Is(err, ErrInvalidParam) to be true, got %v", err)
	}
	if body.reads != 0 {
		t.Fatalf("expected body not to be read, got %d reads", body.reads)
	}
}

func TestBindJSONBodyCacheToggle(t *testing.T) {
	payloadBody := `{"name":"demo"}`

	ctx := NewCtxWithConfig(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payloadBody)), nil, RequestConfig{
		EnableBodyCache: true,
	})

	var payload bindJSONPayload
	if err := ctx.BindJSON(&payload); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	cachedBody, err := io.ReadAll(ctx.R.Body)
	if err != nil {
		t.Fatalf("expected to read cached body, got %v", err)
	}
	if string(cachedBody) != payloadBody {
		t.Fatalf("expected cached body to match payload")
	}

	ctxNoCache := NewCtxWithConfig(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(payloadBody)), nil, RequestConfig{
		EnableBodyCache: false,
	})

	if err := ctxNoCache.BindJSON(&payload); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	uncachedBody, err := io.ReadAll(ctxNoCache.R.Body)
	if err != nil {
		t.Fatalf("expected to read body, got %v", err)
	}
	if len(uncachedBody) != 0 {
		t.Fatalf("expected request body to be consumed when cache is disabled")
	}
}

func TestBindJSONErrors(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		wantMsg      string
		wantSentinel error
	}{
		{name: "empty", body: "", wantMsg: "request body is empty", wantSentinel: ErrEmptyRequestBody},
		{name: "invalid", body: "{", wantMsg: "invalid JSON payload", wantSentinel: ErrInvalidJSON},
		{name: "extra", body: `{"name":"demo"} {}`, wantMsg: "unexpected extra data in request body", wantSentinel: ErrUnexpectedExtraData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body)), nil)
			var payload bindJSONPayload
			err := ctx.BindJSON(&payload)
			if err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}

			var bindErr *bindError
			if !errors.As(err, &bindErr) {
				t.Fatalf("expected bindError, got %T", err)
			}
			if bindErr.Message != tt.wantMsg {
				t.Fatalf("unexpected message: %s", bindErr.Message)
			}
			if !errors.Is(err, tt.wantSentinel) {
				t.Fatalf("errors.Is(%v) returned false; want true", tt.wantSentinel)
			}
		})
	}
}

func TestBindJSONNilContextAndRequestReturnErrors(t *testing.T) {
	var nilCtx *Ctx
	var payload bindJSONPayload
	if err := nilCtx.BindJSON(&payload); !errors.Is(err, ErrContextNil) {
		t.Fatalf("expected ErrContextNil, got %v", err)
	}

	ctx := NewCtx(httptest.NewRecorder(), nil, nil)
	if err := ctx.BindJSON(&payload); !errors.Is(err, ErrRequestNil) {
		t.Fatalf("expected ErrRequestNil, got %v", err)
	}
}

func TestBindJSONNilBodyIsEmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Body = nil
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var payload bindJSONPayload
	if err := ctx.BindJSON(&payload); !errors.Is(err, ErrEmptyRequestBody) {
		t.Fatalf("expected ErrEmptyRequestBody, got %v", err)
	}
}

func TestBindJSONThenValidateStruct(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"demo@example.com","password":"superpass","username":"demo"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	type payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Username string `json:"username" validate:"required,min=3"`
	}

	var p payload
	if err := ctx.BindJSON(&p); err != nil {
		t.Fatalf("expected successful bind, got %v", err)
	}
	if err := ValidateStruct(&p); err != nil {
		t.Fatalf("expected successful validate, got %v", err)
	}
}

func TestValidateStructErrorsAfterBindJSON(t *testing.T) {
	body := bytes.NewBufferString(`{"email":"invalid","password":"short","username":"de"}`)
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", body), nil)

	type payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		Username string `json:"username" validate:"required,min=3"`
	}

	var p payload
	if err := ctx.BindJSON(&p); err != nil {
		t.Fatalf("expected bind to succeed, got %v", err)
	}

	err := ValidateStruct(&p)
	if err == nil {
		t.Fatalf("expected validation error")
	}
	if !strings.Contains(err.Error(), "Email") {
		t.Fatalf("unexpected validation error message: %s", err.Error())
	}
}

func TestNewCtxDoesNotCreateRequestTimeout(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtxWithConfig(httptest.NewRecorder(), req, nil, RequestConfig{})

	if _, ok := ctx.R.Context().Deadline(); ok {
		t.Fatalf("NewCtxWithConfig must not create hidden request deadlines")
	}
}

func TestParamsAndRequestContextHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(WithRequestContext(t.Context(), RequestContext{Params: map[string]string{"id": "42"}}))

	rc := RequestContextFromContext(req.Context())
	if rc.Params["id"] != "42" {
		t.Fatalf("request context did not surface params")
	}
}

func TestRequestContextParamsAreCopied(t *testing.T) {
	params := map[string]string{"id": "42"}
	ctx := WithRequestContext(t.Context(), RequestContext{
		Params:       params,
		RoutePattern: "/users/:id",
		RouteName:    "users.show",
	})

	params["id"] = "mutated"

	rc := RequestContextFromContext(ctx)
	if rc.Params["id"] != "42" {
		t.Fatalf("expected stored params to be isolated, got %q", rc.Params["id"])
	}

	rc.Params["id"] = "returned-mutated"
	again := RequestContextFromContext(ctx)
	if again.Params["id"] != "42" {
		t.Fatalf("expected returned params mutation to be isolated, got %q", again.Params["id"])
	}
	if again.RoutePattern != "/users/:id" || again.RouteName != "users.show" {
		t.Fatalf("expected route metadata to be preserved, got %+v", again)
	}
}

func TestBindErrorHelpers(t *testing.T) {
	err := &bindError{Message: "oops", Err: errors.New("root")}
	if err.Error() != "oops" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
	if !errors.Is(err, err.Err) {
		t.Fatalf("expected unwrap to work")
	}

	var nilErr *bindError
	if nilErr.Error() != "" {
		t.Fatalf("nil receiver should return empty string")
	}
	if nilErr.Unwrap() != nil {
		t.Fatalf("nil receiver unwrap should be nil")
	}
}

func TestCtxParamHelpers(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	params := map[string]string{"name": "demo"}
	ctx := NewCtx(httptest.NewRecorder(), req, params)

	params["name"] = "mutated"

	if val, ok := ctx.Param("name"); !ok || val != "demo" {
		t.Fatalf("Param lookup failed")
	}
	if _, err := ctx.MustParam("missing"); err == nil {
		t.Fatalf("MustParam should error on missing key")
	}
}

func TestBindQuery(t *testing.T) {
	type filter struct {
		Name   string   `query:"name"`
		Page   int      `query:"page"`
		Limit  int64    `query:"limit"`
		Score  float64  `query:"score"`
		Active bool     `query:"active"`
		Tags   []string `query:"tags"`
		Ignore string
	}

	req := httptest.NewRequest(http.MethodGet, "/?name=alice&page=2&limit=50&score=9.5&active=true&tags=go&tags=http", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Name != "alice" || f.Page != 2 || f.Limit != 50 || f.Score != 9.5 || !f.Active {
		t.Fatalf("unexpected scalar fields: %+v", f)
	}
	if len(f.Tags) != 2 || f.Tags[0] != "go" || f.Tags[1] != "http" {
		t.Fatalf("expected tags=[go http], got %v", f.Tags)
	}
	if f.Ignore != "" {
		t.Fatal("field without query tag should not be set")
	}
}

func TestBindQueryMissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Name string `query:"name"`
		Page int    `query:"page"`
	}

	var f filter
	if err := ctx.BindQuery(&f); err != nil {
		t.Fatalf("missing params should not error, got %v", err)
	}
	if f.Name != "" || f.Page != 0 {
		t.Fatal("missing params should leave zero values")
	}
}

func TestBindQueryInvalidType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=notanumber", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct {
		Page int `query:"page"`
	}

	var f filter
	err := ctx.BindQuery(&f)
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
	var bindErr *bindError
	if !errors.As(err, &bindErr) {
		t.Fatalf("expected bindError, got %T", err)
	}
	if bindErr.Status != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", bindErr.Status)
	}
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("expected errors.Is(err, ErrInvalidParam) to be true, got %v", err)
	}
	var numErr *strconv.NumError
	if !errors.As(err, &numErr) {
		t.Fatalf("expected strconv.NumError to remain reachable, got %v", err)
	}
}

func TestBindQueryUnsupportedTaggedFieldReturnsError(t *testing.T) {
	type filter struct {
		When time.Time `query:"when"`
	}

	req := httptest.NewRequest(http.MethodGet, "/?when=now", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	var f filter
	err := ctx.BindQuery(&f)
	if err == nil {
		t.Fatal("expected unsupported query field error")
	}
	if !errors.Is(err, ErrInvalidBindDst) {
		t.Fatalf("expected ErrInvalidBindDst, got %v", err)
	}
}

func TestBindQueryRejectsEmptyTagName(t *testing.T) {
	tests := []struct {
		name string
		tag  any
	}{
		{
			name: "empty name with option",
			tag: &struct {
				Value string `query:",omitempty"`
			}{},
		},
		{
			name: "blank name",
			tag: &struct {
				Value string `query:"   ,omitempty"`
			}{},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/?=empty", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ctx.BindQuery(tt.tag)
			if err == nil {
				t.Fatal("expected empty query tag name error")
			}
			if !errors.Is(err, ErrInvalidBindDst) {
				t.Fatalf("expected ErrInvalidBindDst, got %v", err)
			}
		})
	}
}

func TestBindQueryNilContextAndRequestReturnErrors(t *testing.T) {
	type filter struct {
		Name string `query:"name"`
	}
	var f filter

	var nilCtx *Ctx
	if err := nilCtx.BindQuery(&f); !errors.Is(err, ErrContextNil) {
		t.Fatalf("expected ErrContextNil, got %v", err)
	}

	ctx := NewCtx(httptest.NewRecorder(), nil, nil)
	if err := ctx.BindQuery(&f); !errors.Is(err, ErrRequestNil) {
		t.Fatalf("expected ErrRequestNil, got %v", err)
	}
}

func TestBindQueryNonPointer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := NewCtx(httptest.NewRecorder(), req, nil)

	type filter struct{ Name string }
	var f filter
	err := ctx.BindQuery(f)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}
