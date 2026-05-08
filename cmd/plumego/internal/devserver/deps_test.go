package devserver

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDependencyDiagnosticBufferBoundsOutput(t *testing.T) {
	buf := newDependencyDiagnosticBuffer(8)
	n, err := buf.Write([]byte("1234567890"))
	if err != nil || n != 10 {
		t.Fatalf("write n=%d err=%v", n, err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "12345678") {
		t.Fatalf("expected retained prefix, got %q", out)
	}
	if !strings.Contains(out, "dependency diagnostic output truncated after 8 bytes") ||
		!strings.Contains(out, "2 bytes omitted") {
		t.Fatalf("expected truncation marker, got %q", out)
	}
}

func TestDepsCacheCoalescesConcurrentMisses(t *testing.T) {
	cache := newDepsCache()
	var calls atomic.Int32
	release := make(chan struct{})
	cache.build = func(ctx context.Context, dir string) (*DependencyGraph, error) {
		calls.Add(1)
		select {
		case <-release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return &DependencyGraph{
			GeneratedAt: time.Now(),
			Summary: DepSummary{
				MainModule: "example.com/app",
			},
		}, nil
	}

	const workers = 8
	start := make(chan struct{})
	errs := make(chan error, workers)
	graphs := make(chan *DependencyGraph, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			graph, err := cache.Get(context.Background(), ".", false)
			if err != nil {
				errs <- err
				return
			}
			graphs <- graph
		}()
	}

	close(start)
	deadline := time.After(2 * time.Second)
	for calls.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("dependency graph build did not start")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	time.Sleep(50 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Fatalf("build calls before release = %d, want 1", got)
	}
	close(release)
	wg.Wait()
	close(errs)
	close(graphs)

	for err := range errs {
		t.Fatalf("Get returned error: %v", err)
	}
	var first *DependencyGraph
	count := 0
	for graph := range graphs {
		if graph == nil {
			t.Fatal("expected graph, got nil")
		}
		if first == nil {
			first = graph
		} else if graph != first {
			t.Fatal("concurrent cache miss did not share in-flight graph result")
		}
		count++
	}
	if count != workers {
		t.Fatalf("graph count = %d, want %d", count, workers)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("build calls = %d, want 1", got)
	}
}
