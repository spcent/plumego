package protocol

import (
	"context"
	"io"
	"strings"
	"testing"
)

type testAdapter struct {
	name    string
	prefix  string
	matches bool
}

func (a *testAdapter) Name() string { return a.name }

func (a *testAdapter) Handles(req *HTTPRequest) bool {
	if !a.matches || req == nil {
		return false
	}
	return strings.HasPrefix(req.URL, a.prefix)
}

func (a *testAdapter) Transform(ctx context.Context, req *HTTPRequest) (Request, error) {
	return &GRPCRequest{MethodName: a.name}, nil
}

func (a *testAdapter) Execute(ctx context.Context, req Request) (Response, error) {
	return &GRPCResponse{Status: 200, BodyReader: strings.NewReader(req.Method())}, nil
}

func (a *testAdapter) Encode(ctx context.Context, resp Response, writer ResponseWriter) error {
	writer.WriteHeader(resp.StatusCode())
	if body := resp.Body(); body != nil {
		_, err := io.Copy(writerAdapter{w: writer}, body)
		return err
	}
	return nil
}

type writerAdapter struct {
	w ResponseWriter
}

func (wa writerAdapter) Write(p []byte) (int, error) {
	return wa.w.Write(p)
}

func TestRegistryRegisterFindAndCount(t *testing.T) {
	reg := NewRegistry()

	a1 := &testAdapter{name: "a1", prefix: "/grpc", matches: true}
	a2 := &testAdapter{name: "a2", prefix: "/graphql", matches: true}
	reg.Register(a1)
	reg.Register(a2)

	if got := reg.Count(); got != 2 {
		t.Fatalf("expected count 2, got %d", got)
	}

	req := &HTTPRequest{URL: "/graphql/query"}
	found := reg.Find(req)
	if found == nil || found.Name() != "a2" {
		t.Fatalf("expected adapter a2, got %v", found)
	}

	if miss := reg.Find(&HTTPRequest{URL: "/unknown"}); miss != nil {
		t.Fatalf("expected nil for unknown route, got %v", miss)
	}
}

func TestRegistryAdaptersReturnsCopy(t *testing.T) {
	reg := NewRegistry()
	reg.Register(&testAdapter{name: "a1", prefix: "/a", matches: true})

	adapters := reg.Adapters()
	if len(adapters) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(adapters))
	}

	// Mutating the returned slice must not mutate registry internals.
	adapters[0] = &testAdapter{name: "mutated", prefix: "/m", matches: true}

	after := reg.Adapters()
	if after[0].Name() != "a1" {
		t.Fatalf("expected registry adapter to stay unchanged, got %s", after[0].Name())
	}
}
