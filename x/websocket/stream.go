package websocket

import (
	"bytes"
	"io"
	"sync"
	"time"
)

const maxPooledMessageBufferCap = 64 << 10

// streamReader implements io.ReadCloser for one bounded message.
//
// It is not a low-memory streaming parser: each frame payload is read into
// memory by readFrame and then copied into this reader's buffer before being
// returned to the caller.
type streamReader struct {
	parent  *Conn
	op      byte
	buf     bytes.Buffer
	done    bool
	readErr error
	total   int64
	readMu  sync.Mutex
}

// streamReaderPool recycles streamReader instances to reduce per-message
// heap allocation. Because streamReader embeds a bytes.Buffer, pooling also
// reuses the buffer's internal backing array, saving an extra allocation for
// every received frame.
var streamReaderPool = sync.Pool{
	New: func() any { return new(streamReader) },
}

func putMessageBuffer(buf *bytes.Buffer) bool {
	if buf.Cap() > maxPooledMessageBufferCap {
		return false
	}
	buf.Reset()
	msgBufPool.Put(buf)
	return true
}

func (sr *streamReader) appendPayload(parent *Conn, payload []byte) error {
	limit := parent.readLimit.Load()
	if limit >= 0 && sr.total+int64(len(payload)) > limit {
		return ErrPayloadTooLarge
	}
	sr.total += int64(len(payload))
	_, _ = sr.buf.Write(payload)
	return nil
}

func (sr *streamReader) Read(p []byte) (int, error) {
	sr.readMu.Lock()
	defer sr.readMu.Unlock()

	// Capture parent once. Close() nils sr.parent, so capturing it here
	// into a local prevents a nil-dereference if Close() races with Read()
	// (which is a caller contract violation, but we defend against it anyway).
	parent := sr.parent
	if parent == nil {
		return 0, io.ErrClosedPipe
	}

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
		op, fin, payload, err := parent.readFrame()
		if err != nil {
			sr.readErr = err
			sr.done = true
			return 0, err
		}
		// control frames may appear in middle
		switch op {
		case opcodePing:
			_ = parent.writeFrame(opcodePong, true, payload)
			continue
		case opcodePong:
			parent.lastPong.Store(time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = parent.writeFrame(opcodeClose, true, payload)
			sr.readErr = io.EOF
			sr.done = true
			return 0, io.EOF
		case opcodeContinuation:
			if err := sr.appendPayload(parent, payload); err != nil {
				sr.readErr = err
				sr.done = true
				return 0, err
			}
			if fin {
				sr.done = true
			}
			continue
		case OpcodeText, OpcodeBinary:
			// if new data opcode arrives while not started, treat as start
			// if already started assembling and new data opcode arrives -> protocol error
			if sr.op == 0 {
				sr.op = op
				if err := sr.appendPayload(parent, payload); err != nil {
					sr.readErr = err
					sr.done = true
					return 0, err
				}
				if fin {
					sr.done = true
				}
				continue
			} else {
				// shouldn't get new data opcode while assembling
				sr.readErr = ErrProtocolError
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
	// Nil the parent pointer so the pool doesn't hold a reference to the Conn.
	sr.parent = nil
	if sr.buf.Cap() > maxPooledMessageBufferCap {
		sr.buf = bytes.Buffer{}
	} else {
		sr.buf.Reset()
	}
	streamReaderPool.Put(sr)
	return nil
}

// ReadMessageStream returns (opcode, io.ReadCloser, error) for one complete
// bounded inbound message.
//
// The returned reader is bounded, not low-memory or zero-copy: continuation
// frames are pulled as the reader advances, but each frame payload is buffered
// in memory by readFrame and copied into the returned reader. Caller must
// Close() the returned ReadCloser when finished to allow pooling and connection
// progress.
func (c *Conn) ReadMessageStream() (byte, io.ReadCloser, error) {
	if c.IsClosed() {
		return 0, nil, ErrConnClosed
	}
	// read first frame - must be text/binary or control
	for {
		op, fin, payload, err := c.readFrame()
		if err != nil {
			return 0, nil, err
		}
		switch op {
		case OpcodeText, OpcodeBinary:
			// Get a pooled reader and reset all fields from any prior use.
			sr := streamReaderPool.Get().(*streamReader)
			sr.parent = c
			sr.op = 0
			sr.done = false
			sr.readErr = nil
			sr.total = 0
			sr.buf.Reset() // keeps the backing array; avoids re-allocation

			// write payload into buffer
			if err := sr.appendPayload(c, payload); err != nil {
				_ = sr.Close()
				return 0, nil, err
			}
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
			c.lastPong.Store(time.Now().UnixNano())
			continue
		case opcodeClose:
			_ = c.writeFrame(opcodeClose, true, payload)
			return 0, nil, io.EOF
		case opcodeContinuation:
			return 0, nil, ErrProtocolError
		default:
			return 0, nil, ErrProtocolError
		}
	}
}

// ReadMessage reads a complete message into memory and returns an owned byte
// slice.
func (c *Conn) ReadMessage() (byte, []byte, error) {
	op, stream, err := c.ReadMessageStream()
	if err != nil {
		return 0, nil, err
	}

	buf := msgBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	_, err = io.Copy(buf, stream)
	_ = stream.Close()
	if err != nil {
		putMessageBuffer(buf)
		return 0, nil, err
	}
	// Copy data before returning buf to pool; callers expect to own the slice.
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	putMessageBuffer(buf)
	return op, data, nil
}
