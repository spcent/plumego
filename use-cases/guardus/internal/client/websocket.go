package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spcent/plumego/x/websocket"
)

// QueryWebSocket dials address, writes body once, and returns the next message.
func QueryWebSocket(address, body string, headers map[string]string, config *Config) (bool, []byte, error) {
	const Origin = "http://localhost/"
	wsHeaders := http.Header{}
	wsHeaders.Set("Origin", Origin)
	for name, value := range headers {
		wsHeaders.Set(name, value)
	}

	ctx := context.Background()
	if config != nil && config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	httpClient := http.DefaultClient
	if config != nil {
		tlsConfig := &tls.Config{InsecureSkipVerify: config.Insecure}
		if config.HasTLSConfig() && config.TLS.isValid() == nil {
			tlsConfig = configureTLS(tlsConfig, *config.TLS)
		}
		httpClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: tlsConfig},
		}
	}

	conn, _, err := websocket.Dial(ctx, address, &websocket.DialOptions{
		HTTPHeader: wsHeaders,
		HTTPClient: httpClient,
	})
	if err != nil {
		return false, nil, fmt.Errorf("error dialing websocket: %w", err)
	}
	defer func() { _ = conn.WriteClose(websocket.CloseNormalClosure, "") }()

	body = parseLocalAddressPlaceholder(body, addrLocal{})
	if err := conn.WriteMessage(websocket.OpcodeText, []byte(body)); err != nil {
		return false, nil, fmt.Errorf("error writing websocket body: %w", err)
	}
	op, msg, err := conn.ReadMessage()
	if err != nil {
		return false, nil, fmt.Errorf("error reading websocket message: %w", err)
	}
	if op != websocket.OpcodeText && op != websocket.OpcodeBinary {
		return false, nil, fmt.Errorf("unexpected websocket message opcode: %d", op)
	}
	return true, msg, nil
}

// addrLocal is a stub net.Addr because the websocket extension does not expose
// the underlying connection's local address. Body templates that reference
// [LOCAL_ADDRESS] in the WS path will see an empty string.
type addrLocal struct{}

func (addrLocal) Network() string { return "ws" }
func (addrLocal) String() string  { return "" }
