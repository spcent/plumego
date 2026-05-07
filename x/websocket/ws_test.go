package websocket

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

// ---------------- Tests & Benchmark ----------------
// To run unit tests and benchmark, run: go test -v -bench . -run ^$
// There is a simple benchmark that spawns many simulated clients and broadcasts messages.

func TestJWTAndRoomAuth(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	auth := NewSimpleRoomAuth()
	if err := auth.SetRoomPassword("a", "p"); err != nil {
		t.Fatalf("SetRoomPassword: %v", err)
	}
	if !auth.AuthorizeRoom("a", "p") {
		t.Fatal("password check failed")
	}
	tokenAuth, err := NewHS256TokenAuth(secret)
	if err != nil {
		t.Fatalf("NewHS256TokenAuth: %v", err)
	}
	// create a token
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user1","exp":` + fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()) + `}`))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := header + "." + payload + "." + sig
	if _, err := tokenAuth.AuthenticateToken(token); err != nil {
		t.Fatal("verify jwt failed:", err)
	}
}

func startTestServer(t *testing.T) (*http.Server, *Hub, string) {
	workerCount := 4
	jobQueueSize := 1024
	sendQueueSize := 64
	sendTimeout := 50 * time.Millisecond
	sendBehavior := SendBlock
	hub := mustHub(t, workerCount, jobQueueSize)
	secret := []byte("0123456789abcdef0123456789abcdef")
	auth := NewSimpleRoomAuth()
	if err := auth.SetRoomPassword("room1", "pwd1"); err != nil {
		t.Fatalf("SetRoomPassword: %v", err)
	}
	tokenAuth, err := NewHS256TokenAuth(secret)
	if err != nil {
		t.Fatalf("NewHS256TokenAuth: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeRoomFanoutWS(w, r, ServerConfig{
			Hub:                  hub,
			RoomAuth:             auth,
			TokenAuth:            tokenAuth,
			AllowUnauthenticated: true,
			QueueSize:            sendQueueSize,
			SendTimeout:          sendTimeout,
			SendBehavior:         sendBehavior,
			AllowedOrigins:       []string{"*"},
		})
	})
	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(ln)
	return server, hub, "http://" + ln.Addr().String()
}

// minimal WS client helper that performs handshake and basic send/receive frames
type testWSClient struct {
	conn net.Conn
	br   *bufio.Reader
	bw   *bufio.Writer
}

func newTestWSClient(t *testing.T, url string, room, pwd, token string) *testWSClient {
	// parse host
	// url like http://host:port/ws
	parts := strings.Split(url, "://")
	if len(parts) != 2 {
		t.Fatal("bad url")
	}
	hostPart := strings.TrimPrefix(parts[1], "http://")
	host := hostPart
	conn, err := net.Dial("tcp", host)
	if err != nil {
		t.Fatal(err)
	}
	// build request
	keyBytes := make([]byte, 16)
	_, err = rand.Read(keyBytes)
	if err != nil {
		t.Fatal(err)
	}

	key := base64.StdEncoding.EncodeToString(keyBytes)
	path := "/ws"
	if room != "" {
		path += "?room=" + room
	}
	authHeader := ""
	if token != "" {
		authHeader = "Authorization: Bearer " + token + "\r\n"
	}
	roomPasswordHeader := ""
	if pwd != "" {
		roomPasswordHeader = RoomPasswordHeader + ": " + pwd + "\r\n"
	}
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: %s\r\n%s%s\r\n", path, host, key, authHeader, roomPasswordHeader)
	bw := bufio.NewWriter(conn)
	_, _ = bw.WriteString(req)
	_ = bw.Flush()
	br := bufio.NewReader(conn)
	// read response lines
	status, err := br.ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "101") {
		t.Fatalf("handshake failed: %s", status)
	}
	// consume headers until blank line
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	return &testWSClient{conn: conn, br: br, bw: bufio.NewWriter(conn)}
}

func (c *testWSClient) sendFrame(op byte, fin bool, payload []byte) error {
	var header [14]byte
	hlen := 0
	b0 := byte(0)
	if fin {
		b0 |= finBit
	}
	b0 |= op & 0x0F
	header[0] = b0
	hlen = 1
	n := len(payload)
	switch {
	case n <= 125:
		header[hlen] = byte(n) | 0x80 // masked
		hlen++
	case n <= 0xFFFF:
		header[hlen] = 126 | 0x80
		hlen++
		binary.BigEndian.PutUint16(header[hlen:hlen+2], uint16(n))
		hlen += 2
	default:
		header[hlen] = 127 | 0x80
		hlen++
		binary.BigEndian.PutUint64(header[hlen:hlen+8], uint64(n))
		hlen += 8
	}
	// mask key
	maskKey := make([]byte, 4)
	rand.Read(maskKey)
	if _, err := c.bw.Write(header[:hlen]); err != nil {
		return err
	}
	if _, err := c.bw.Write(maskKey); err != nil {
		return err
	}
	// masked payload
	masked := make([]byte, n)
	for i := 0; i < n; i++ {
		masked[i] = payload[i] ^ maskKey[i%4]
	}
	if _, err := c.bw.Write(masked); err != nil {
		return err
	}
	return c.bw.Flush()
}

func (c *testWSClient) sendText(s string) error {
	return c.sendFrame(OpcodeText, true, []byte(s))
}

func (c *testWSClient) readFrame() (byte, bool, []byte, error) {
	var h [2]byte
	if _, err := io.ReadFull(c.br, h[:]); err != nil {
		return 0, false, nil, err
	}
	fin := h[0]&finBit != 0
	op := h[0] & 0x0F
	mask := h[1]&0x80 != 0
	len7 := int64(h[1] & 0x7F)
	var payloadLen int64
	switch len7 {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return 0, false, nil, err
		}
		payloadLen = int64(binary.BigEndian.Uint16(ext[:]))
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(c.br, ext[:]); err != nil {
			return 0, false, nil, err
		}
		payloadLen = int64(binary.BigEndian.Uint64(ext[:]))
	default:
		payloadLen = len7
	}
	var maskKey [4]byte
	if mask {
		if _, err := io.ReadFull(c.br, maskKey[:]); err != nil {
			return 0, false, nil, err
		}
	}
	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		if _, err := io.ReadFull(c.br, payload); err != nil {
			return 0, false, nil, err
		}
	}
	if mask {
		for i := int64(0); i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}
	return op, fin, payload, nil
}

func TestSimpleEchoAndRoom(t *testing.T) {
	server, hub, base := startTestServer(t)
	defer server.Close()
	defer hub.Stop()

	// create token to pass JWT verification; server expects startTestServer's secret
	secret := []byte("0123456789abcdef0123456789abcdef")
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u","exp":` + fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()) + `}`))
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := header + "." + payload + "." + sig

	cli := newTestWSClient(t, base, "room1", "pwd1", token)
	defer cli.conn.Close()
	// send a message
	if err := cli.sendText("hello"); err != nil {
		t.Fatal(err)
	}
	// read broadcast (server will broadcast back)
	op, _, payloadb, err := cli.readFrame()
	if err != nil {
		t.Fatal(err)
	}
	if op != OpcodeText {
		t.Fatalf("expected text op, got %d", op)
	}
	if string(payloadb) != "hello" {
		t.Fatalf("expected hello, got %s", string(payloadb))
	}
}

func TestServeWSWithConfigUsesMessageHandler(t *testing.T) {
	hub := mustHub(t, 1, 16)
	defer hub.Stop()

	received := make(chan Message, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  hub,
			RoomAuth:             NewSimpleRoomAuth(),
			AllowUnauthenticated: true,
			SendBehavior:         SendBlock,
			AllowedOrigins:       []string{"*"},
			OnMessage: func(_ *Conn, msg Message) {
				received <- msg
			},
		})
	})

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(ln)
	defer server.Close()

	cli := newTestWSClient(t, "http://"+ln.Addr().String(), "", "", "")
	defer cli.conn.Close()

	if err := cli.sendText("handled"); err != nil {
		t.Fatal(err)
	}

	select {
	case msg := <-received:
		if msg.Room != "default" {
			t.Fatalf("room = %q, want default", msg.Room)
		}
		if msg.Opcode != OpcodeText {
			t.Fatalf("opcode = %d, want %d", msg.Opcode, OpcodeText)
		}
		if string(msg.Data) != "handled" {
			t.Fatalf("data = %q, want handled", string(msg.Data))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message handler")
	}
}

func TestServeWSWithConfigRecoversMessageHandlerPanic(t *testing.T) {
	hub := mustHub(t, 1, 16)
	defer hub.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  hub,
			RoomAuth:             NewSimpleRoomAuth(),
			AllowUnauthenticated: true,
			SendBehavior:         SendBlock,
			AllowedOrigins:       []string{"*"},
			OnMessage: func(*Conn, Message) {
				panic("handler failed")
			},
		})
	})

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(ln)
	defer server.Close()

	cli := newTestWSClient(t, "http://"+ln.Addr().String(), "", "", "")
	defer cli.conn.Close()

	if err := cli.sendText("boom"); err != nil {
		t.Fatal(err)
	}
	if err := cli.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := cli.readFrame(); err == nil {
		t.Fatal("expected connection close after message handler panic")
	}
}

func TestServeWSWithConfigClosesWhenMessageHandlerQueueFull(t *testing.T) {
	hub := mustHub(t, 1, 16)
	defer hub.Stop()

	blockHandler := make(chan struct{})
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWSWithConfig(w, r, ServerConfig{
			Hub:                  hub,
			RoomAuth:             NewSimpleRoomAuth(),
			AllowUnauthenticated: true,
			QueueSize:            1,
			SendBehavior:         SendBlock,
			AllowedOrigins:       []string{"*"},
			OnMessage: func(*Conn, Message) {
				<-blockHandler
			},
		})
	})

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(ln)
	defer server.Close()
	defer close(blockHandler)

	cli := newTestWSClient(t, "http://"+ln.Addr().String(), "", "", "")
	defer cli.conn.Close()

	if err := cli.sendText("one"); err != nil {
		t.Fatal(err)
	}
	if err := cli.sendText("two"); err != nil {
		t.Fatal(err)
	}
	if err := cli.sendText("three"); err != nil {
		t.Fatal(err)
	}
	if err := cli.conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := cli.readFrame(); err == nil {
		t.Fatal("expected connection close after handler queue fills")
	}
}
