package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"
)

// WriteMessage enqueues message to sendQueue. Behavior on full queue depends on c.sendBehavior.
// It waits up to sendTimeout if blocking behavior chosen.
//
// Deprecated: Use WriteMessageContext for better cancellation support.
func (c *Conn) WriteMessage(op byte, data []byte) error {
	ctx := context.Background()
	if c.sendTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.sendTimeout)
		defer cancel()
	}
	return c.WriteMessageContext(ctx, op, data)
}

// WriteMessageContext enqueues message to sendQueue with context support.
// Behavior on full queue depends on c.sendBehavior.
//
// The context can be used for cancellation and custom timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	err := conn.WriteMessageContext(ctx, websocket.OpcodeText, []byte("hello"))
func (c *Conn) WriteMessageContext(ctx context.Context, op byte, data []byte) error {
	if c.IsClosed() {
		return ErrConnClosed
	}

	out := Outbound{Op: op, Data: data}

	// Fast path: try non-blocking send first
	select {
	case c.sendQueue <- out:
		return nil
	default:
		// Queue is full, handle according to behavior
	}

	// Handle full queue based on behavior
	switch c.sendBehavior {
	case SendBlock:
		// Block until space available, context cancelled, or connection closed
		select {
		case c.sendQueue <- out:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-c.closeC:
			return ErrConnClosed
		}

	case SendDrop:
		return ErrQueueFull

	case SendClose:
		c.Close()
		return ErrQueueFullClosed

	default:
		return errors.New("unknown send behavior")
	}
}

// WriteText sends a text message
func (c *Conn) WriteText(data string) error {
	return c.WriteMessage(OpcodeText, []byte(data))
}

// WriteBinary sends a binary message
func (c *Conn) WriteBinary(data []byte) error {
	return c.WriteMessage(OpcodeBinary, data)
}

// WriteJSON serializes and sends a JSON message
func (c *Conn) WriteJSON(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.WriteMessage(OpcodeText, data)
}

// writerPump consumes sendQueue and writes frames to client. It fragments large messages.
func (c *Conn) writerPump() {
	ticker := time.NewTicker(time.Duration(atomic.LoadInt64((*int64)(&c.pingPeriod))))
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case <-c.closeC:
			return
		case out, ok := <-c.sendQueue:
			if !ok {
				return
			}
			// fragment if needed
			data := out.Data
			if len(data) <= maxFragmentSize {
				_ = c.writeFrame(out.Op, true, data)
				continue
			}
			total := len(data)
			offset := 0
			first := true
			for offset < total {
				end := offset + maxFragmentSize
				if end > total {
					end = total
				}
				chunk := data[offset:end]
				fin := end == total
				var op byte
				if first {
					op = out.Op
					first = false
				} else {
					op = opcodeContinuation
				}
				if err := c.writeFrame(op, fin, chunk); err != nil {
					c.Close()
					return
				}
				offset = end
			}
		case <-ticker.C:
			// send ping
			_ = c.writeFrame(opcodePing, true, []byte("ping"))
		}
	}
}

// pongMonitor closes connection if no pong received within pongWait
func (c *Conn) pongMonitor() {
	ticker := time.NewTicker(time.Duration(atomic.LoadInt64((*int64)(&c.pingPeriod))) / 2)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeC:
			return
		case <-ticker.C:
			last := time.Unix(0, atomic.LoadInt64(&c.lastPong))
			if time.Since(last) > time.Duration(atomic.LoadInt64((*int64)(&c.pongWait))) {
				_ = c.Close()
				return
			}
		}
	}
}
