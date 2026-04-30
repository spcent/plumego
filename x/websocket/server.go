package websocket

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
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

// isOriginAllowed checks if the request origin is explicitly allowed.
func isOriginAllowed(origin string, allowedOrigins []string, allowAll bool) bool {
	if origin == "" {
		return true
	}
	if allowAll {
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
	Hub                  *Hub
	Auth                 RoomAuthenticator
	QueueSize            int
	SendTimeout          time.Duration
	SendBehavior         SendBehavior
	ReadLimit            int64 // Optional max inbound frame payload size. 0 means use defaults.
	MessageValidation    MessageValidationConfig
	AllowedOrigins       []string // Allowed origins for CORS. Use ["*"] to allow all origins.
	AllowAllOrigins      bool     // Explicitly disable origin checks for development or trusted non-browser clients.
	AllowUnauthenticated bool     // Explicitly allow connections without JWT; room passwords still apply.
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
		return cfg, ErrNilHub
	}
	if cfg.Auth == nil {
		return cfg, ErrNilAuthenticator
	}
	if cfg.QueueSize < 0 {
		return cfg, ErrNegativeQueueSize
	}
	if cfg.SendBehavior < SendBlock || cfg.SendBehavior > SendClose {
		return cfg, ErrInvalidSendBehavior
	}
	if cfg.ReadLimit < 0 {
		return cfg, ErrNegativeReadLimit
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
// Origin validation is explicitly disabled for this compatibility helper.
// Use ServeWSWithConfig with a non-empty AllowedOrigins list for strict
// CSRF protection.
func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth RoomAuthenticator, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	ServeWSWithConfig(w, r, ServerConfig{
		Hub:             hub,
		Auth:            auth,
		QueueSize:       queueSize,
		SendTimeout:     sendTimeout,
		SendBehavior:    behavior,
		AllowAllOrigins: true, // explicit compatibility behavior; callers requiring CSRF protection should use ServeWSWithConfig
	})
}

// ServeWSWithConfig performs the WebSocket handshake with full configuration
// options including origin validation (CSRF protection).
func ServeWSWithConfig(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	normalized, err := normalizeServerConfig(cfg)
	if err != nil {
		writeWebSocketHandshakeError(w, r, http.StatusInternalServerError, codeWebSocketInvalidConfig, "websocket server misconfigured", contract.CategoryServer)
		return
	}
	cfg = normalized

	// Origin validation (CSRF protection)
	origin := r.Header.Get("Origin")
	if !isOriginAllowed(origin, cfg.AllowedOrigins, cfg.AllowAllOrigins) {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketForbiddenOrigin, "forbidden origin", contract.CategoryClient)
		return
	}

	// Basic HTTP validation first
	if r.Method != http.MethodGet {
		writeWebSocketHandshakeError(w, r, http.StatusMethodNotAllowed, contract.CodeMethodNotAllowed, "method not allowed", contract.CategoryClient)
		return
	}
	if !headerContains(r.Header, "Connection", "Upgrade") || !headerContains(r.Header, "Upgrade", "websocket") {
		writeWebSocketHandshakeError(w, r, http.StatusBadRequest, codeWebSocketBadUpgrade, "websocket upgrade required", contract.CategoryClient)
		return
	}

	// Validate WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		writeWebSocketHandshakeError(w, r, http.StatusBadRequest, codeWebSocketKeyMissing, "websocket key required", contract.CategoryClient)
		return
	}
	if err := ValidateWebSocketKey(key); err != nil {
		cfg.Hub.invalidWSKeys.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusBadRequest, codeWebSocketKeyInvalid, "invalid websocket key", contract.CategoryClient)
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
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketRoomForbidden, "websocket room access denied", contract.CategoryClient)
		return
	}
	if err := cfg.Hub.CanJoin(room); err != nil {
		status := websocketJoinDeniedStatus(err)
		writeWebSocketHandshakeError(w, r, status, codeWebSocketJoinDenied, "websocket room join denied", contract.CategoryClient)
		return
	}

	// Check token. Missing tokens are rejected unless callers explicitly allow
	// unauthenticated connections and rely on room-password checks only.
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if t := r.URL.Query().Get("token"); t != "" {
		token = t
	}
	if token == "" {
		if !cfg.AllowUnauthenticated {
			cfg.Hub.securityRejections.Add(1)
			writeWebSocketHandshakeError(w, r, http.StatusUnauthorized, codeWebSocketTokenRequired, "websocket token required", contract.CategoryClient)
			return
		}
	} else {
		payload, err := cfg.Auth.VerifyJWT(token)
		if err != nil {
			cfg.Hub.securityRejections.Add(1)
			writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketInvalidToken, "invalid websocket token", contract.CategoryClient)
			return
		}
		userInfo = ExtractUserInfo(payload)
		cfg.Hub.successfulAuths.Add(1)
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		writeWebSocketHandshakeError(w, r, http.StatusInternalServerError, codeWebSocketHijackUnsupported, "websocket hijack unsupported", contract.CategoryServer)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		writeWebSocketHandshakeError(w, r, http.StatusInternalServerError, codeWebSocketHandshakeFailed, "websocket handshake failed", contract.CategoryServer)
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

	// Register in the hub before writing 101. This closes the race where the
	// pre-check passes, capacity is consumed by another connection, and this
	// request would otherwise be upgraded before the real join failure.
	if err := cfg.Hub.TryJoin(room, c); err != nil {
		writeHijackedWebSocketHandshakeError(buf.Writer, websocketJoinDeniedStatus(err), codeWebSocketJoinDenied, "websocket room join denied", contract.CategoryClient)
		c.Close()
		return
	}

	accept := computeAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n" +
		"\r\n"
	if _, err := buf.WriteString(resp); err != nil {
		cfg.Hub.RemoveConn(c)
		c.Close()
		return
	}
	if err := buf.Flush(); err != nil {
		cfg.Hub.RemoveConn(c)
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

func writeWebSocketHandshakeError(w http.ResponseWriter, r *http.Request, status int, code, message string, category contract.ErrorCategory) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Status(status).
		Code(code).
		Message(message).
		Category(category).
		Build())
}

func writeHijackedWebSocketHandshakeError(bw *bufio.Writer, status int, code, message string, category contract.ErrorCategory) {
	apiErr := contract.NewErrorBuilder().
		Status(status).
		Code(code).
		Message(message).
		Category(category).
		Build()
	body, err := json.Marshal(contract.ErrorResponse{Error: apiErr})
	if err != nil {
		body = []byte(`{"error":{"code":"WEBSOCKET_HANDSHAKE_FAILED","message":"websocket handshake failed"}}`)
	}

	_, _ = bw.WriteString("HTTP/1.1 " + strconv.Itoa(status) + " " + http.StatusText(status) + "\r\n")
	_, _ = bw.WriteString("Content-Type: application/json\r\n")
	_, _ = bw.WriteString("Connection: close\r\n")
	_, _ = bw.WriteString("Content-Length: " + strconv.Itoa(len(body)) + "\r\n")
	_, _ = bw.WriteString("\r\n")
	_, _ = bw.Write(body)
	_ = bw.Flush()
}

func websocketJoinDeniedStatus(err error) int {
	if errors.Is(err, ErrRoomFull) {
		return http.StatusTooManyRequests
	}
	return http.StatusServiceUnavailable
}
