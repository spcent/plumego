package websocket

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
)

// computeAcceptKey computes the WebSocket accept key
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// headerContains reports whether the comma-separated header value contains val
// (case-insensitive). Iterates without allocating a slice.
func headerContains(h http.Header, key, val string) bool {
	v := h.Get(key)
	for v != "" {
		var token string
		if i := strings.IndexByte(v, ','); i >= 0 {
			token, v = strings.TrimSpace(v[:i]), v[i+1:]
		} else {
			token, v = strings.TrimSpace(v), ""
		}
		if strings.EqualFold(token, val) {
			return true
		}
	}
	return false
}

// isOriginAllowed checks if the request origin is in the allowed list.
// Returns true if allowedOrigins is nil/empty (skip validation) or contains "*" or the specific origin.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return true
	}
	for _, allowed := range allowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}
	return false
}

// ServerConfig configures WebSocket server options.
//
// Auth must implement RoomAuthenticator. Use NewSimpleRoomAuth or NewSecureRoomAuth
// to create a concrete implementation.
type ServerConfig struct {
	Hub               *Hub
	Auth              RoomAuthenticator
	QueueSize         int
	SendTimeout       time.Duration
	SendBehavior      SendBehavior
	ReadLimit         int64 // Optional max inbound frame payload size. 0 means use defaults.
	MessageValidation MessageValidationConfig
	AllowedOrigins    []string // Allowed origins for CORS. Use ["*"] to allow all origins.
}

type messageSizeProvider interface {
	MaxMessageSize() int64
}

func maxIntValue() int64 {
	return int64(^uint(0) >> 1)
}

func readLimitToInt(limit int64) int {
	if limit <= 0 {
		return 0
	}
	maxInt := maxIntValue()
	if limit > maxInt {
		return int(maxInt)
	}
	return int(limit)
}

func normalizeServerConfig(cfg ServerConfig) (ServerConfig, error) {
	if cfg.Hub == nil {
		return cfg, errors.New("websocket server misconfigured: hub is nil")
	}
	if cfg.Auth == nil {
		return cfg, errors.New("websocket server misconfigured: authenticator is nil")
	}
	if cfg.QueueSize < 0 {
		return cfg, errors.New("websocket server misconfigured: queue size cannot be negative")
	}
	if cfg.SendBehavior < SendBlock || cfg.SendBehavior > SendClose {
		return cfg, errors.New("websocket server misconfigured: invalid send behavior")
	}
	if cfg.ReadLimit < 0 {
		return cfg, errors.New("websocket server misconfigured: read limit cannot be negative")
	}
	if cfg.ReadLimit == 0 {
		if p, ok := cfg.Auth.(messageSizeProvider); ok {
			if lim := p.MaxMessageSize(); lim > 0 {
				cfg.ReadLimit = lim
			}
		}
	}
	return cfg, nil
}

func resolveValidationConfig(cfg ServerConfig) MessageValidationConfig {
	validationCfg := cfg.MessageValidation
	if validationCfg == (MessageValidationConfig{}) {
		validationCfg = DefaultMessageValidationConfig()
	}
	if cfg.ReadLimit > 0 {
		maxLen := readLimitToInt(cfg.ReadLimit)
		if maxLen > 0 && (validationCfg.MaxLength == 0 || validationCfg.MaxLength > maxLen) {
			validationCfg.MaxLength = maxLen
		}
	}
	return validationCfg
}

// ServeWSWithAuth performs the WebSocket handshake with JWT and room-password
// authentication. It accepts tokens from the Authorization header
// ("Bearer <token>") or the ?token= query parameter, and room passwords from
// the ?room_password= query parameter.
//
// Origin validation is explicitly set to allow all origins (["*"]).
// Use ServeWSWithConfig with a non-empty AllowedOrigins list for strict
// CSRF protection.
func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth RoomAuthenticator, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	ServeWSWithConfig(w, r, ServerConfig{
		Hub:            hub,
		Auth:           auth,
		QueueSize:      queueSize,
		SendTimeout:    sendTimeout,
		SendBehavior:   behavior,
		AllowedOrigins: []string{"*"}, // explicit allow-all; callers requiring CSRF protection should use ServeWSWithConfig
	})
}

// ServeWSWithConfig performs the WebSocket handshake with full configuration
// options including origin validation (CSRF protection).
func ServeWSWithConfig(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	normalized, err := normalizeServerConfig(cfg)
	if err != nil {
		contract.WriteError(w, r, contract.NewInternalError(err.Error()))
		return
	}
	cfg = normalized

	// Origin validation (CSRF protection)
	origin := r.Header.Get("Origin")
	if origin != "" && !isOriginAllowed(origin, cfg.AllowedOrigins) {
		cfg.Hub.securityRejections.Add(1)
		contract.WriteError(w, r, contract.NewForbiddenError("forbidden origin"))
		return
	}

	// Basic HTTP validation first
	if r.Method != http.MethodGet {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusMethodNotAllowed, Code: "METHOD_NOT_ALLOWED", Message: "method not allowed", Category: contract.CategoryClient})
		return
	}
	if !headerContains(r.Header, "Connection", "Upgrade") || !headerContains(r.Header, "Upgrade", "websocket") {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "bad request", Category: contract.CategoryClient})
		return
	}

	// Validate WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: "bad request", Category: contract.CategoryClient})
		return
	}
	if err := ValidateWebSocketKey(key); err != nil {
		cfg.Hub.invalidWSKeys.Add(1)
		contract.WriteError(w, r, contract.APIError{Status: http.StatusBadRequest, Code: "BAD_REQUEST", Message: err.Error(), Category: contract.CategoryClient})
		return
	}

	// Auth: room and JWT
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	// Check room password
	roomPwd := r.URL.Query().Get("room_password")
	if !cfg.Auth.CheckRoomPassword(room, roomPwd) {
		cfg.Hub.securityRejections.Add(1)
		contract.WriteError(w, r, contract.NewForbiddenError("forbidden: bad room password"))
		return
	}
	if err := cfg.Hub.CanJoin(room); err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, ErrRoomFull) {
			status = http.StatusTooManyRequests
		}
		contract.WriteError(w, r, contract.APIError{Status: status, Code: "JOIN_DENIED", Message: err.Error(), Category: contract.CategoryClient})
		return
	}

	// Check token if present
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if t := r.URL.Query().Get("token"); t != "" {
		token = t
	}
	if token != "" {
		payload, err := cfg.Auth.VerifyJWT(token)
		if err != nil {
			cfg.Hub.securityRejections.Add(1)
			contract.WriteError(w, r, contract.NewForbiddenError("forbidden: invalid token"))
			return
		}
		userInfo = ExtractUserInfo(payload)
		cfg.Hub.successfulAuths.Add(1)
	}

	accept := computeAcceptKey(key)
	hj, ok := w.(http.Hijacker)
	if !ok {
		contract.WriteError(w, r, contract.NewInternalError("server does not support hijacking"))
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		contract.WriteError(w, r, contract.NewInternalError("hijack failed"))
		return
	}

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n" +
		"\r\n"
	if _, err := buf.WriteString(resp); err != nil {
		conn.Close()
		return
	}
	if err := buf.Flush(); err != nil {
		conn.Close()
		return
	}

	// Reuse the bufio.ReadWriter returned by Hijack to avoid a redundant
	// buffer allocation (NewConn would otherwise create default-sized buffers
	// that are immediately discarded).
	c := newConnFromHijack(conn, buf.Reader, buf.Writer, cfg.QueueSize, cfg.SendTimeout, cfg.SendBehavior)
	if cfg.ReadLimit > 0 {
		c.SetReadLimit(cfg.ReadLimit)
	}
	c.UserInfo = userInfo

	// Register in hub
	if err := cfg.Hub.TryJoin(room, c); err != nil {
		// Close the Conn wrapper so writerPump/pongMonitor goroutines stop
		// cleanly via the closeC channel.
		c.Close()
		return
	}

	// Cleanup on close: remove from all rooms once the connection is gone.
	go func() {
		<-c.closeC
		cfg.Hub.RemoveConn(c)
	}()

	// Read frames from the client and broadcast to the room.
	go func() {
		validationCfg := resolveValidationConfig(cfg)
		for {
			op, rstream, err := c.ReadMessageStream()
			if err != nil {
				if err != io.EOF {
					cfg.Hub.logger.Printf("ReadMessageStream error: %v", err)
				}
				c.Close()
				return
			}
			buf := msgBufPool.Get().(*bytes.Buffer)
			buf.Reset()
			if _, err := io.Copy(buf, rstream); err != nil {
				_ = rstream.Close()
				msgBufPool.Put(buf)
				cfg.Hub.logger.Printf("ReadMessageStream copy error: %v", err)
				c.Close()
				return
			}
			if err := rstream.Close(); err != nil {
				msgBufPool.Put(buf)
				cfg.Hub.logger.Printf("ReadMessageStream close error: %v", err)
				c.Close()
				return
			}

			// Validate text messages before broadcasting.
			if op == OpcodeText {
				if err := ValidateTextMessage(buf.Bytes(), validationCfg); err != nil {
					cfg.Hub.logger.Printf("dropped invalid text message: %v", err)
					msgBufPool.Put(buf)
					continue
				}
			}

			// Copy data before returning buf to pool; BroadcastRoom enqueues
			// it asynchronously so the pool buffer must not be reused yet.
			data := make([]byte, buf.Len())
			copy(data, buf.Bytes())
			msgBufPool.Put(buf)
			cfg.Hub.BroadcastRoom(room, op, data)
		}
	}()
}

