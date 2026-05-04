package websocket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func maskedClientFrame(first byte, payload []byte) []byte {
	frame := []byte{first}
	switch {
	case len(payload) <= 125:
		frame = append(frame, 0x80|byte(len(payload)))
	case len(payload) <= 0xFFFF:
		frame = append(frame, 0x80|126, byte(len(payload)>>8), byte(len(payload)))
	default:
		frame = append(frame, 0x80|127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(len(payload)))
		frame = append(frame, ext[:]...)
	}
	frame = append(frame, 0, 0, 0, 0)
	frame = append(frame, payload...)
	return frame
}

func newFrameReadConn(frame []byte) *Conn {
	c := &Conn{
		br:     bufio.NewReader(bytes.NewReader(frame)),
		closeC: make(chan struct{}),
	}
	c.readLimit.Store(1024)
	return c
}

func TestReadFrameRejectsRSVBits(t *testing.T) {
	c := newFrameReadConn(maskedClientFrame(finBit|0x40|OpcodeText, []byte("x")))
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsUnknownOpcode(t *testing.T) {
	c := newFrameReadConn(maskedClientFrame(finBit|0x3, []byte("x")))
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadMessageStreamRejectsContinuationBeforeMessage(t *testing.T) {
	c := newFrameReadConn(maskedClientFrame(finBit|opcodeContinuation, []byte("x")))
	_, _, err := c.ReadMessageStream()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("ReadMessageStream error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsNonMinimalLength126(t *testing.T) {
	frame := []byte{finBit | OpcodeText, 0x80 | 126, 0, 1, 0, 0, 0, 0, 'x'}
	c := newFrameReadConn(frame)
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsNonMinimalLength127(t *testing.T) {
	frame := []byte{finBit | OpcodeText, 0x80 | 127}
	var ext [8]byte
	binary.BigEndian.PutUint64(ext[:], 126)
	frame = append(frame, ext[:]...)
	frame = append(frame, 0, 0, 0, 0)
	frame = append(frame, bytes.Repeat([]byte("x"), 126)...)

	c := newFrameReadConn(frame)
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsClosePayloadLengthOne(t *testing.T) {
	c := newFrameReadConn(maskedClientFrame(finBit|opcodeClose, []byte{0x03}))
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsInvalidCloseCode(t *testing.T) {
	var payload [2]byte
	binary.BigEndian.PutUint16(payload[:], 1006)
	c := newFrameReadConn(maskedClientFrame(finBit|opcodeClose, payload[:]))
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}

func TestReadFrameRejectsInvalidCloseReasonUTF8(t *testing.T) {
	payload := []byte{0x03, 0xe8, 0xff}
	c := newFrameReadConn(maskedClientFrame(finBit|opcodeClose, payload))
	_, _, _, err := c.readFrame()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("readFrame error = %v, want ErrProtocolError", err)
	}
}
