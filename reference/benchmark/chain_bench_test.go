package benchmark_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
)

// ---------------------------------------------------------------------------
// stdlib-compatible middleware chain benchmarks (plumego / chi / stdlib)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Gin middleware chain benchmarks
// ---------------------------------------------------------------------------

func BenchmarkGinChain1NoOp(b *testing.B) {
	runGinChainBenchmark(b, 1, ginNoOpHandler())
}

func BenchmarkGinChain3NoOp(b *testing.B) {
	runGinChainBenchmark(b, 3, ginNoOpHandler())
}

func BenchmarkGinChain5NoOp(b *testing.B) {
	runGinChainBenchmark(b, 5, ginNoOpHandler())
}

func BenchmarkGinChain1JSON(b *testing.B) {
	runGinChainBenchmark(b, 1, ginJSONHandler())
}

func BenchmarkGinChain3JSON(b *testing.B) {
	runGinChainBenchmark(b, 3, ginJSONHandler())
}

func BenchmarkGinChain5JSON(b *testing.B) {
	runGinChainBenchmark(b, 5, ginJSONHandler())
}

// ---------------------------------------------------------------------------
// Echo middleware chain benchmarks
// ---------------------------------------------------------------------------

func BenchmarkEchoChain1NoOp(b *testing.B) {
	runEchoChainBenchmark(b, 1, echoNoOpHandler())
}

func BenchmarkEchoChain3NoOp(b *testing.B) {
	runEchoChainBenchmark(b, 3, echoNoOpHandler())
}

func BenchmarkEchoChain5NoOp(b *testing.B) {
	runEchoChainBenchmark(b, 5, echoNoOpHandler())
}

func BenchmarkEchoChain1JSON(b *testing.B) {
	runEchoChainBenchmark(b, 1, echoJSONHandler())
}

func BenchmarkEchoChain3JSON(b *testing.B) {
	runEchoChainBenchmark(b, 3, echoJSONHandler())
}

func BenchmarkEchoChain5JSON(b *testing.B) {
	runEchoChainBenchmark(b, 5, echoJSONHandler())
}

// ---------------------------------------------------------------------------
// stdlib chain runner
// ---------------------------------------------------------------------------

func runChainBenchmark(b *testing.B, depth int, final http.Handler, wantStatus int) {
	b.Helper()
	h := applyStdlibMiddleware(final, stdlibMiddlewareStack(depth)...)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != wantStatus {
			b.Fatalf("unexpected status: got %d want %d", rec.Code, wantStatus)
		}
	}
}

func stdlibMiddlewareStack(depth int) []func(http.Handler) http.Handler {
	stack := make([]func(http.Handler) http.Handler, 0, depth)
	for i := range depth {
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

func applyStdlibMiddleware(final http.Handler, stack ...func(http.Handler) http.Handler) http.Handler {
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
	}{ID: "bench", Status: "ok"}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// ---------------------------------------------------------------------------
// Gin chain runner
// ---------------------------------------------------------------------------

func runGinChainBenchmark(b *testing.B, depth int, final gin.HandlerFunc) {
	b.Helper()
	r := gin.New()
	handlers := make(gin.HandlersChain, 0, depth+1)
	handlers = append(handlers, ginMiddlewareStack(depth)...)
	handlers = append(handlers, final)
	r.GET("/benchmark", handlers...)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
	}
}

func ginMiddlewareStack(depth int) []gin.HandlerFunc {
	stack := make([]gin.HandlerFunc, 0, depth)
	for i := range depth {
		layer := strconv.Itoa(i + 1)
		stack = append(stack, func(c *gin.Context) {
			c.Header("X-Benchmark-Layer", layer)
			c.Next()
		})
	}
	return stack
}

func ginNoOpHandler() gin.HandlerFunc {
	return func(c *gin.Context) { c.Status(http.StatusNoContent) }
}

func ginJSONHandler() gin.HandlerFunc {
	payload := gin.H{"id": "bench", "status": "ok"}
	return func(c *gin.Context) { c.JSON(http.StatusOK, payload) }
}

// ---------------------------------------------------------------------------
// Echo chain runner
// ---------------------------------------------------------------------------

func runEchoChainBenchmark(b *testing.B, depth int, final echo.HandlerFunc) {
	b.Helper()
	e := echo.New()
	for _, mw := range echoMiddlewareStack(depth) {
		e.Use(mw)
	}
	e.GET("/benchmark", final)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
	}
}

func echoMiddlewareStack(depth int) []echo.MiddlewareFunc {
	stack := make([]echo.MiddlewareFunc, 0, depth)
	for i := range depth {
		layer := strconv.Itoa(i + 1)
		stack = append(stack, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Response().Header().Add("X-Benchmark-Layer", layer)
				return next(c)
			}
		})
	}
	return stack
}

func echoNoOpHandler() echo.HandlerFunc {
	return func(c echo.Context) error { return c.NoContent(http.StatusNoContent) }
}

func echoJSONHandler() echo.HandlerFunc {
	payload := map[string]string{"id": "bench", "status": "ok"}
	return func(c echo.Context) error { return c.JSON(http.StatusOK, payload) }
}
