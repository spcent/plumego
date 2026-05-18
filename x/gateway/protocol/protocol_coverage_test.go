package protocol

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// ---- Registry edge cases ----

func TestRegistryFindNoMatch(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&testAdapter{name: "graphql", path: "/graphql"})

	got := reg.Find(&HTTPRequest{URL: "/rest/users"})
	if got != nil {
		t.Fatalf("expected nil for unmatched URL, got adapter %q", got.Name())
	}
}

func TestRegistryFindEmptyRegistry(t *testing.T) {
	reg := NewRegistry()
	got := reg.Find(&HTTPRequest{URL: "/anything"})
	if got != nil {
		t.Fatalf("expected nil from empty registry, got %+v", got)
	}
}

func TestRegistryFirstMatchWins(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&testAdapter{name: "first", path: "/api"})
	reg.Register(&testAdapter{name: "second", path: "/api"})

	got := reg.Find(&HTTPRequest{URL: "/api/v1"})
	if got == nil || got.Name() != "first" {
		t.Fatalf("expected first registered adapter to win, got %+v", got)
	}
}

func TestRegistryCountMatchesRegistrations(t *testing.T) {
	reg := NewRegistry()
	if reg.Count() != 0 {
		t.Fatalf("empty registry count = %d, want 0", reg.Count())
	}

	reg.Register(&testAdapter{name: "a", path: "/a"})
	reg.Register(&testAdapter{name: "b", path: "/b"})

	if reg.Count() != 2 {
		t.Fatalf("count = %d, want 2", reg.Count())
	}
}

func TestRegistryAdaptersIsDefensiveCopy(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&testAdapter{name: "orig", path: "/orig"})

	adapters1 := reg.Adapters()
	adapters1[0] = &testAdapter{name: "mutated", path: "/mutated"}

	adapters2 := reg.Adapters()
	if adapters2[0].Name() != "orig" {
		t.Fatalf("Adapters() should return a defensive copy; registry was mutated to %q", adapters2[0].Name())
	}
}

// ---- GRPCRequest / GRPCResponse helper types ----

func TestGRPCRequest_NilMeta(t *testing.T) {
	req := &GRPCRequest{MethodName: "ListUsers", Meta: nil}

	if req.Method() != "ListUsers" {
		t.Fatalf("Method = %q, want ListUsers", req.Method())
	}
	meta := req.Metadata()
	if len(meta) != 0 {
		t.Fatalf("Metadata with nil Meta should be empty, got %v", meta)
	}
}

func TestGRPCRequest_NilHeaderMap(t *testing.T) {
	req := &GRPCRequest{HeaderMap: nil}
	hdrs := req.Headers()
	if hdrs == nil {
		t.Fatal("Headers() should never return nil")
	}
	if len(hdrs) != 0 {
		t.Fatalf("expected empty headers, got %v", hdrs)
	}
}

func TestGRPCRequest_MetadataPopulated(t *testing.T) {
	req := &GRPCRequest{
		MethodName: "SayHello",
		Meta: &GRPCMetadata{
			Service:     "helloworld.Greeter",
			Method:      "SayHello",
			Authority:   "localhost:50051",
			Timeout:     "10s",
			ContentType: "application/grpc",
			Encoding:    "gzip",
			UserAgent:   "grpc-go/1.0",
			Custom:      map[string]string{"x-request-id": "abc"},
		},
	}

	meta := req.Metadata()
	checks := map[string]string{
		"service":      "helloworld.Greeter",
		"method":       "SayHello",
		"authority":    "localhost:50051",
		"timeout":      "10s",
		"content_type": "application/grpc",
		"encoding":     "gzip",
		"user_agent":   "grpc-go/1.0",
		"x-request-id": "abc",
	}
	for k, want := range checks {
		got, ok := meta[k]
		if !ok {
			t.Errorf("metadata missing key %q", k)
			continue
		}
		if got != want {
			t.Errorf("metadata[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestGRPCRequest_BodyReader(t *testing.T) {
	body := bytes.NewBufferString("request body")
	req := &GRPCRequest{BodyReader: body}
	data, err := io.ReadAll(req.Body())
	if err != nil {
		t.Fatalf("Body() read error: %v", err)
	}
	if string(data) != "request body" {
		t.Fatalf("Body = %q, want %q", data, "request body")
	}
}

func TestGRPCResponse_DefaultStatusCode(t *testing.T) {
	resp := &GRPCResponse{Status: 0}
	if resp.StatusCode() != 200 {
		t.Fatalf("zero Status should default to 200, got %d", resp.StatusCode())
	}
}

func TestGRPCResponse_NilHeaderMap(t *testing.T) {
	resp := &GRPCResponse{HeaderMap: nil}
	if resp.Headers() == nil {
		t.Fatal("Headers() should never return nil")
	}
}

func TestGRPCResponse_NilMetaMetadata(t *testing.T) {
	resp := &GRPCResponse{Meta: nil}
	meta := resp.Metadata()
	if len(meta) != 0 {
		t.Fatalf("Metadata with nil Meta should be empty, got %v", meta)
	}
}

func TestGRPCResponse_BodyReader(t *testing.T) {
	body := strings.NewReader("response body")
	resp := &GRPCResponse{Status: 200, BodyReader: body}
	data, err := io.ReadAll(resp.Body())
	if err != nil {
		t.Fatalf("Body() read error: %v", err)
	}
	if string(data) != "response body" {
		t.Fatalf("Body = %q, want response body", data)
	}
}

// ---- GRPCErrorCode mapping ----

func TestGRPCErrorCode_KnownCodes(t *testing.T) {
	cases := []struct {
		grpc int
		http int
	}{
		{0, 200},
		{1, 499},
		{2, 500},
		{3, 400},
		{4, 504},
		{5, 404},
		{6, 409},
		{7, 403},
		{8, 429},
		{9, 400},
		{10, 409},
		{11, 400},
		{12, 501},
		{13, 500},
		{14, 503},
		{15, 500},
		{16, 401},
	}

	for _, tc := range cases {
		got := GRPCErrorCode(tc.grpc)
		if got != tc.http {
			t.Errorf("GRPCErrorCode(%d) = %d, want %d", tc.grpc, got, tc.http)
		}
	}
}

func TestGRPCErrorCode_UnknownCodeReturns500(t *testing.T) {
	for _, code := range []int{17, 99, 1000, -1} {
		if got := GRPCErrorCode(code); got != 500 {
			t.Errorf("GRPCErrorCode(%d) = %d, want 500", code, got)
		}
	}
}

// ---- HTTPRequest fields ----

func TestHTTPRequest_Fields(t *testing.T) {
	body := strings.NewReader("body")
	req := &HTTPRequest{
		Method:   "POST",
		URL:      "/graphql",
		Headers:  map[string][]string{"Content-Type": {"application/json"}},
		Body:     body,
		Metadata: map[string]any{"trace-id": "xyz"},
	}

	if req.Method != "POST" {
		t.Fatalf("Method = %q, want POST", req.Method)
	}
	if req.URL != "/graphql" {
		t.Fatalf("URL = %q, want /graphql", req.URL)
	}
	if req.Headers["Content-Type"][0] != "application/json" {
		t.Fatalf("Content-Type header missing or wrong")
	}
	if req.Metadata["trace-id"] != "xyz" {
		t.Fatalf("Metadata trace-id missing")
	}
}

// ---- testAdapter (defined in adapter_test.go) exercises ProtocolAdapter contract ----

func TestTestAdapterHandles(t *testing.T) {
	a := &testAdapter{name: "grpc", path: "/grpc"}

	if !a.Handles(&HTTPRequest{URL: "/grpc/health"}) {
		t.Fatal("should handle /grpc/health")
	}
	if a.Handles(&HTTPRequest{URL: "/rest/health"}) {
		t.Fatal("should not handle /rest/health")
	}
	if a.Handles(nil) {
		t.Fatal("should not handle nil request")
	}
}

func TestTestAdapterTransform(t *testing.T) {
	a := &testAdapter{name: "grpc", path: "/grpc"}
	req, err := a.Transform(context.Background(), &HTTPRequest{URL: "/grpc/call"})
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if req == nil {
		t.Fatal("Transform should return a non-nil Request")
	}
}

func TestTestAdapterExecute(t *testing.T) {
	a := &testAdapter{name: "grpc", path: "/grpc"}
	req, _ := a.Transform(context.Background(), &HTTPRequest{URL: "/grpc/call"})
	resp, err := a.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp == nil {
		t.Fatal("Execute should return a non-nil Response")
	}
	if resp.StatusCode() != 200 {
		t.Fatalf("StatusCode = %d, want 200", resp.StatusCode())
	}
}
