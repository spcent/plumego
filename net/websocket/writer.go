package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// WriteMessage enqueues a message to the send queue.
// Behavior on a full queue depends on c.sendBehavior.
// It waits up to sendTimeout when SendBlock behavior is chosen.
//
// Use WriteMessageContext when you need explicit cancellation or a custom
// deadline beyond the configured sendTimeout.
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
	period := time.Duration(c.pingPeriod.Load())
	ticker := time.NewTicker(period)
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
				if err := c.writeFrame(out.Op, true, data); err != nil {
					c.Close()
					return
				}
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
			// Pick up any runtime change to ping period made via SetPingPeriod.
			if newPeriod := time.Duration(c.pingPeriod.Load()); newPeriod != period {
				period = newPeriod
				ticker.Reset(period)
			}
			if err := c.writeFrame(opcodePing, true, []byte("ping")); err != nil {
				c.Close()
				return
			}
		}
	}
}

// pongMonitor closes the connection if no pong is received within pongWait.
// The check interval is pongWait/3 so that a dead connection is detected
// within at most 4/3 * pongWait of the last failed pong, regardless of how
// pingPeriod is configured.
func (c *Conn) pongMonitor() {
	period := time.Duration(c.pongWait.Load()) / 3
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-c.closeC:
			return
		case <-ticker.C:
			// Pick up any runtime change to pongWait made via SetPongWait.
			if newPeriod := time.Duration(c.pongWait.Load()) / 3; newPeriod != period {
				period = newPeriod
				ticker.Reset(period)
			}
			last := time.Unix(0, c.lastPong.Load())
			if time.Since(last) > time.Duration(c.pongWait.Load()) {
				_ = c.Close()
				return
			}
		}
	}
}
