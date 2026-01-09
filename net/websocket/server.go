package websocket

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// computeAcceptKey computes the WebSocket accept key
func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + guid))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// headerContains checks if a header contains a value (case-insensitive)
func headerContains(h http.Header, key, val string) bool {
	v := h.Get(key)
	if v == "" {
		return false
	}
	parts := strings.Split(v, ",")
	for _, p := range parts {
		if strings.EqualFold(strings.TrimSpace(p), val) {
			return true
		}
	}
	return false
}

// ServeWSWithAuth does the handshake and basic auth checks before accepting connection.
// Accepts token from Authorization header "Bearer <token>" or ?token= in query.
// room password from ?room_password= param.
// onConn will be called when Conn is ready.
func ServeWSWithAuth(w http.ResponseWriter, r *http.Request, hub *Hub, auth *simpleRoomAuth, queueSize int, sendTimeout time.Duration, behavior SendBehavior) {
	// Validate WebSocket key
	key := r.Header.Get("Sec-WebSocket-Key")
	if err := ValidateWebSocketKey(key); err != nil {
		metricsMutex.Lock()
		securityMetrics.InvalidWebSocketKeys++
		metricsMutex.Unlock()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !headerContains(r.Header, "Connection", "Upgrade") || !headerContains(r.Header, "Upgrade", "websocket") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if key == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	// auth: room and JWT
	room := r.URL.Query().Get("room")
	if room == "" {
		room = "default"
	}
	// check room password
	roomPwd := r.URL.Query().Get("room_password")
	if !auth.CheckRoomPassword(room, roomPwd) {
		metricsMutex.Lock()
		securityMetrics.RejectedConnections++
		metricsMutex.Unlock()
		http.Error(w, "forbidden: bad room password", http.StatusForbidden)
		return
	}
	if err := hub.CanJoin(room); err != nil {
		status := http.StatusServiceUnavailable
		if errors.Is(err, ErrRoomFull) {
			status = http.StatusTooManyRequests
		}
		http.Error(w, err.Error(), status)
		return
	}
	// check token if present
	token := ""
	var userInfo *UserInfo
	if ah := r.Header.Get("Authorization"); ah != "" && strings.HasPrefix(strings.ToLower(ah), "bearer ") {
		token = strings.TrimSpace(ah[len("bearer "):])
	} else if t := r.URL.Query().Get("token"); t != "" {
		token = t
	}
	if token != "" {
		payload, err := auth.VerifyJWT(token)
		if err != nil {
			metricsMutex.Lock()
			securityMetrics.RejectedConnections++
			metricsMutex.Unlock()
			http.Error(w, "forbidden: invalid token", http.StatusForbidden)
			return
		}
		// Extract user information from JWT payload
		userInfo = ExtractUserInfo(payload)
		metricsMutex.Lock()
		securityMetrics.SuccessfulAuthentications++
		metricsMutex.Unlock()
	}
	accept := computeAcceptKey(key)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server does not support hijacking", http.StatusInternalServerError)
		return
	}
	conn, buf, err := hj.Hijack()
	if err != nil {
		http.Error(w, "hijack failed", http.StatusInternalServerError)
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

	c := NewConn(conn, queueSize, sendTimeout, behavior)
	// Set user information if authenticated
	c.UserInfo = userInfo
	// override br/bw with hijacked conn
	c.br = bufio.NewReaderSize(conn, 8192)
	c.bw = bufio.NewWriterSize(conn, 8192)

	// register in hub
	if err := hub.TryJoin(room, c); err != nil {
		conn.Close()
		return
	}
	// cleanup on close
	go func() {
		<-c.closeC
		hub.Leave(room, c)
		hub.RemoveConn(c)
	}()
	// spawn a goroutine to read frames and push to room broadcast on completed stream
	go func() {
		for {
			op, rstream, err := c.ReadMessageStream()
			if err != nil {
				if err != io.EOF {
					log.Println("ReadMessageStream error:", err)
				}
				c.Close()
				return
			}
			// For demonstration: stream copy to a buffer but streaming-friendly.
			// If message is huge, we stream to temp file or process chunk-by-chunk.
			buf := &bytes.Buffer{}
			_, _ = io.Copy(buf, rstream) // streaming
			_ = rstream.Close()
			// broadcast to room
			hub.BroadcastRoom(room, op, buf.Bytes())
		}
	}()
}

// UpgradeClient performs client-side WebSocket upgrade
// This is useful for testing or when you need to create a client connection
func UpgradeClient(url string, header http.Header) (net.Conn, *bufio.ReadWriter, error) {
	// This is a simplified client upgrade - in production you'd want more features
	return nil, nil, errors.New("client upgrade not implemented")
}
