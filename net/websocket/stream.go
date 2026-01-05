package websocket

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// streamReader implements io.ReadCloser to stream frames for one message.
type streamReader struct {
	parent    *Conn
	op        byte
	buf       bytes.Buffer
	done      bool
	readErr   error
	readMu    sync.Mutex
	closeOnce sync.Once
	closed    chan struct{}
}

func (sr *streamReader) Read(p []byte) (int, error) {
	sr.readMu.Lock()
	defer sr.readMu.Unlock()

	for {
		if sr.buf.Len() > 0 {
			return sr.buf.Read(p)
		}
		if sr.done {
			if sr.readErr != nil {
				return 0, sr.readErr
			}
			return 0, io.EOF
		}
		// need to pull next frame(s) from connection
		op, fin, payload, err := sr.parent.readFrame()
		if err != nil {
			sr.readErr = err
			sr.done = true
			return 0, err
		}
		// control frames may appear in middle
		switch op {
		case opcodePing:
			_ = sr.parent.writeFrame(opcodePong, true, payload)
			continue
		case opcodePong:
			atomic.StoreInt64(&sr.parent.lastPong, time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = sr.parent.writeFrame(opcodeClose, true, payload)
			sr.readErr = io.EOF
			sr.done = true
			return 0, io.EOF
		case opcodeContinuation:
			// append
			sr.buf.Write(payload)
			if fin {
				sr.done = true
			}
			continue
		case OpcodeText, OpcodeBinary:
			// if new data opcode arrives while not started, treat as start
			// if already started assembling and new data opcode arrives -> protocol error
			if sr.op == 0 {
				sr.op = op
				sr.buf.Write(payload)
				if fin {
					sr.done = true
				}
				continue
			} else {
				// shouldn't get new data opcode while assembling
				sr.readErr = errors.New("protocol error: new data opcode while assembling")
				sr.done = true
				return 0, sr.readErr
			}
		default:
			// ignore other opcodes
			continue
		}
	}
}

func (sr *streamReader) Close() error {
	sr.closeOnce.Do(func() { close(sr.closed) })
	return nil
}

// ReadMessageStream returns (opcode, io.ReadCloser, error).
// Caller must Close() the returned ReadCloser when finished to allow connection continue.
func (c *Conn) ReadMessageStream() (byte, io.ReadCloser, error) {
	if c.IsClosed() {
		return 0, nil, errors.New("connection closed")
	}
	// read first frame - must be text/binary or control
	for {
		op, fin, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch op {
		case OpcodeText, OpcodeBinary:
			// If fin == true and small, can return a reader that contains payload and EOF immediately
			sr := &streamReader{
				parent: c,
				op:     0,
				closed: make(chan struct{}),
			}
			// write payload into buffer
			sr.buf.Write(payload)
			if fin {
				sr.done = true
			} else {
				// not finished; set op marker and let subsequent continuation frames be read by Read
				sr.op = op
			}
			return op, sr, nil
		case opcodePing:
			_ = c.writeFrame(opcodePong, true, payload)
			continue
		case opcodePong:
			atomic.StoreInt64(&c.lastPong, time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = c.writeFrame(opcodeClose, true, payload)
			return 0, nil, io.EOF
		default:
			// ignore
			continue
		}
	}
}

// ReadMessage reads a complete message into memory
func (c *Conn) ReadMessage() (byte, []byte, error) {
	op, stream, err := c.ReadMessageStream()
	if err != nil {
		return 0, nil, err
	}
	defer stream.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, stream)
	if err != nil {
		return 0, nil, err
	}
	return op, buf.Bytes(), nil
}
