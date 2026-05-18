package benchmark_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func BenchmarkChain1NoOp(b *testing.B) {
	runChainBenchmark(b, 1, benchmarkNoOpHandler(), http.StatusNoContent)
}

func BenchmarkChain3NoOp(b *testing.B) {
	runChainBenchmark(b, 3, benchmarkNoOpHandler(), http.StatusNoContent)
}

func BenchmarkChain5NoOp(b *testing.B) {
	runChainBenchmark(b, 5, benchmarkNoOpHandler(), http.StatusNoContent)
}

func BenchmarkChain1JSON(b *testing.B) {
	runChainBenchmark(b, 1, benchmarkJSONHandler(), http.StatusOK)
}

func BenchmarkChain3JSON(b *testing.B) {
	runChainBenchmark(b, 3, benchmarkJSONHandler(), http.StatusOK)
}

func BenchmarkChain5JSON(b *testing.B) {
	runChainBenchmark(b, 5, benchmarkJSONHandler(), http.StatusOK)
}

func runChainBenchmark(b *testing.B, depth int, final http.Handler, wantStatus int) {
	b.Helper()
	h := applyBenchmarkMiddleware(final, benchmarkMiddlewareStack(depth)...)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != wantStatus {
			b.Fatalf("unexpected status: got %d want %d", rec.Code, wantStatus)
		}
	}
}

func benchmarkMiddlewareStack(depth int) []func(http.Handler) http.Handler {
	stack := make([]func(http.Handler) http.Handler, 0, depth)
	for i := 0; i < depth; i++ {
		layer := strconv.Itoa(i + 1)
		stack = append(stack, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("X-Benchmark-Layer", layer)
				next.ServeHTTP(w, r)
			})
		})
	}
	return stack
}

func applyBenchmarkMiddleware(final http.Handler, stack ...func(http.Handler) http.Handler) http.Handler {
	h := final
	for i := len(stack) - 1; i >= 0; i-- {
		h = stack[i](h)
	}
	return h
}

func benchmarkJSONHandler() http.Handler {
	payload := struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}{
		ID:     "bench",
		Status: "ok",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
