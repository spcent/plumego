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
// Browser requests with an Origin header require an explicit allowed origin or
// "*" entry. Non-browser requests without Origin skip this check at the call
// site.
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
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

// Message describes one fully read inbound websocket message.
type Message struct {
	Room   string
	Opcode byte
	Data   []byte
}

// MessageHandler handles a fully read inbound websocket message.
type MessageHandler func(conn *Conn, msg Message)

// ServerConfig configures WebSocket server options.
//
// RoomAuth authorizes room access. TokenAuth authenticates bearer tokens when
// AllowUnauthenticated is false or when a token is provided.
type ServerConfig struct {
	Hub                  *Hub
	RoomAuth             RoomAuthorizer
	TokenAuth            TokenAuthenticator
	OnMessage            MessageHandler
	AllowUnauthenticated bool
	AllowQueryToken      bool
	QueueSize            int
	SendTimeout          time.Duration
	SendBehavior         SendBehavior
	ReadLimit            int64 // Optional max inbound frame payload size. 0 means use defaults.
	MessageValidation    MessageValidationConfig
	AllowedOrigins       []string // Allowed origins for CORS. Use ["*"] to allow all origins.
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
	if cfg.RoomAuth == nil {
		return cfg, ErrNilRoomAuthorizer
	}
	if cfg.TokenAuth == nil && !cfg.AllowUnauthenticated {
		return cfg, ErrNilTokenAuthorizer
	}
	if cfg.QueueSize < 0 {
		return cfg, ErrNegativeQueueSize
	}
	if cfg.SendTimeout < 0 {
		return cfg, ErrNegativeSendTimeout
	}
	if cfg.SendBehavior < SendBlock || cfg.SendBehavior > SendClose {
		return cfg, ErrInvalidSendBehavior
	}
	if cfg.ReadLimit < 0 {
		return cfg, ErrNegativeReadLimit
	}
	if cfg.OnMessage == nil {
		return cfg, ErrNilMessageHandler
	}
	if cfg.ReadLimit == 0 {
		if p, ok := cfg.TokenAuth.(messageSizeProvider); ok {
			if lim := p.MaxMessageSize(); lim > 0 {
				cfg.ReadLimit = lim
			}
		} else if p, ok := cfg.RoomAuth.(messageSizeProvider); ok {
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

// ServeRoomFanoutWS performs the WebSocket handshake and broadcasts every
// accepted client message back to the same room.
func ServeRoomFanoutWS(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	if cfg.OnMessage != nil {
		writeWebSocketHandshakeError(w, r, http.StatusInternalServerError, codeWebSocketInvalidConfig, "websocket server misconfigured", contract.CategoryServer)
		return
	}
	cfg.OnMessage = func(_ *Conn, msg Message) {
		cfg.Hub.BroadcastRoom(msg.Room, msg.Opcode, msg.Data)
	}
	ServeWSWithConfig(w, r, cfg)
}

// ServeWSWithConfig performs the WebSocket handshake with full configuration
// options including origin validation (CSRF protection). Accepted messages are
// delivered to cfg.OnMessage; this low-level handler does not broadcast by
// default.
func ServeWSWithConfig(w http.ResponseWriter, r *http.Request, cfg ServerConfig) {
	normalized, err := normalizeServerConfig(cfg)
	if err != nil {
		writeWebSocketHandshakeError(w, r, http.StatusInternalServerError, codeWebSocketInvalidConfig, "websocket server misconfigured", contract.CategoryServer)
		return
	}
	cfg = normalized

	// Origin validation (CSRF protection)
	origin := r.Header.Get("Origin")
	if origin != "" && !isOriginAllowed(origin, cfg.AllowedOrigins) {
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
	if version := r.Header.Get("Sec-WebSocket-Version"); version != "13" {
		writeWebSocketHandshakeError(w, r, http.StatusBadRequest, codeWebSocketVersionUnsupported, "unsupported websocket version", contract.CategoryClient)
		return
	}

	// RoomAuth: room and JWT
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	if err := ValidateRoomName(room); err != nil {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusBadRequest, codeWebSocketRoomInvalid, "invalid websocket room", contract.CategoryClient)
		return
	}

	// Check room password
	if r.URL.Query().Get("room_password") != "" {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketRoomForbidden, "websocket room access denied", contract.CategoryClient)
		return
	}
	roomPwd := r.Header.Get(RoomPasswordHeader)
	if !cfg.RoomAuth.AuthorizeRoom(room, roomPwd) {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketRoomForbidden, "websocket room access denied", contract.CategoryClient)
		return
	}
	if err := cfg.Hub.CanJoin(room); err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, ErrRoomFull) {
			status = http.StatusTooManyRequests
		}
		writeWebSocketHandshakeError(w, r, status, codeWebSocketJoinDenied, "websocket room join denied", contract.CategoryClient)
		return
	}

	// Check token if present
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if cfg.AllowQueryToken {
		t := r.URL.Query().Get("token")
		token = t
	} else if t := r.URL.Query().Get("token"); t != "" {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketInvalidToken, "invalid websocket token", contract.CategoryClient)
		return
	}
	if token != "" {
		payload, err := cfg.TokenAuth.AuthenticateToken(token)
		if err != nil {
			cfg.Hub.securityRejections.Add(1)
			writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketInvalidToken, "invalid websocket token", contract.CategoryClient)
			return
		}
		userInfo = ExtractUserInfo(payload)
		cfg.Hub.successfulAuths.Add(1)
	} else if !cfg.AllowUnauthenticated {
		cfg.Hub.securityRejections.Add(1)
		writeWebSocketHandshakeError(w, r, http.StatusForbidden, codeWebSocketInvalidToken, "invalid websocket token", contract.CategoryClient)
		return
	}

	accept := computeAcceptKey(key)
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

	// Reuse the bufio.ReadWriter returned by Hijack to avoid the redundant
	// default-sized buffers that NewConnE allocates for direct net.Conn callers.
	c := newConnFromHijack(conn, buf.Reader, buf.Writer, cfg.QueueSize, cfg.SendTimeout, cfg.SendBehavior)
	if cfg.ReadLimit > 0 {
		if err := c.SetReadLimit(cfg.ReadLimit); err != nil {
			c.Close()
			return
		}
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
				putMessageBuffer(buf)
				cfg.Hub.logger.Printf("ReadMessageStream copy error: %v", err)
				c.Close()
				return
			}
			if err := rstream.Close(); err != nil {
				putMessageBuffer(buf)
				cfg.Hub.logger.Printf("ReadMessageStream close error: %v", err)
				c.Close()
				return
			}

			// Validate text messages before broadcasting.
			if op == OpcodeText {
				if err := ValidateTextMessage(buf.Bytes(), validationCfg); err != nil {
					cfg.Hub.logger.Printf("dropped invalid text message: %v", err)
					putMessageBuffer(buf)
					continue
				}
			}

			// Copy data before returning buf to pool; BroadcastRoom enqueues
			// it asynchronously so the pool buffer must not be reused yet.
			data := make([]byte, buf.Len())
			copy(data, buf.Bytes())
			putMessageBuffer(buf)
			cfg.OnMessage(c, Message{
				Room:   room,
				Opcode: op,
				Data:   data,
			})
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
