package desktop

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spcent/plumego/log"
)

// LocalServer wraps the HTTP server for desktop mode with security constraints.
type LocalServer struct {
	server    *http.Server
	port      int
	token     string
	logger    log.StructuredLogger
	dataDir   string
}

// NewLocalServer creates a new local HTTP server bound to 127.0.0.1.
// It generates a random token for security and selects a random available port.
func NewLocalServer(handler http.Handler, dataDir string, logger log.StructuredLogger) (*LocalServer, error) {
	// Generate random security token
	token, err := generateToken(32)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// Find available port
	port, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("find port: %w", err)
	}

	// Wrap handler with token middleware
	wrappedHandler := tokenAuthMiddleware(handler, token, logger)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:         addr,
		Handler:      wrappedHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return &LocalServer{
		server:  server,
		port:    port,
		token:   token,
		logger:  logger,
		dataDir: dataDir,
	}, nil
}

// Start starts the local server in a goroutine.
func (s *LocalServer) Start(ctx context.Context) error {
	s.logger.Info("starting local server", map[string]any{
		"addr": s.server.Addr,
		"port": s.port,
	})

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("local server error", map[string]any{"error": err.Error()})
		}
	}()

	// Wait for server to be ready
	return s.waitForReady(ctx)
}

// Stop gracefully shuts down the server.
func (s *LocalServer) Stop(ctx context.Context) error {
	s.logger.Info("stopping local server")
	return s.server.Shutdown(ctx)
}

// Port returns the port the server is listening on.
func (s *LocalServer) Port() int {
	return s.port
}

// Token returns the security token.
func (s *LocalServer) Token() string {
	return s.token
}

// BaseURL returns the full base URL with token.
func (s *LocalServer) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d?token=%s", s.port, s.token)
}

// waitForReady waits for the server to be ready to accept connections.
func (s *LocalServer) waitForReady(ctx context.Context) error {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := net.DialTimeout("tcp", s.server.Addr, 100*time.Millisecond)
			if err == nil {
				conn.Close()
				return nil
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
	return fmt.Errorf("server did not become ready within 5 seconds")
}

// findAvailablePort finds an available port on localhost.
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// generateToken generates a cryptographically secure random token.
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// tokenAuthMiddleware validates the token on each request.
func tokenAuthMiddleware(next http.Handler, token string, logger log.StructuredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check token in query parameter or header
		queryToken := r.URL.Query().Get("token")
		headerToken := r.Header.Get("X-Auth-Token")

		if queryToken != token && headerToken != token {
			logger.Warn("invalid token attempt", map[string]any{
				"remote_addr": r.RemoteAddr,
				"path":        r.URL.Path,
			})
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
