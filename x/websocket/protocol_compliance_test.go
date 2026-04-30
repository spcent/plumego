package websocket

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestReadFrameRejectsMalformedProtocolInputs(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "rsv bit set",
			data: maskedFrameWithHeader(finBit|0x40|OpcodeText, 0, nil, nil),
		},
		{
			name: "reserved data opcode",
			data: maskedFrameWithHeader(finBit|0x3, 0, nil, nil),
		},
		{
			name: "reserved control opcode",
			data: maskedFrameWithHeader(finBit|0xB, 0, nil, nil),
		},
		{
			name: "non-minimal 126 length",
			data: maskedFrameWithHeader(finBit|OpcodeText, 126, []byte{0, 1}, []byte("x")),
		},
		{
			name: "non-minimal 127 length",
			data: maskedFrameWithHeader(finBit|OpcodeBinary, 127, uint64Bytes(126), bytes.Repeat([]byte("x"), 126)),
		},
		{
			name: "close payload length one",
			data: maskedFrame(opcodeClose, true, []byte{0x03}),
		},
		{
			name: "invalid close code",
			data: maskedFrame(opcodeClose, true, []byte{0x03, 0xEE}),
		},
		{
			name: "invalid close reason utf8",
			data: maskedFrame(opcodeClose, true, []byte{0x03, 0xE8, 0xFF}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newProtocolTestConn(tt.data)
			_, _, _, err := conn.readFrame()
			if !errors.Is(err, ErrProtocolError) {
				t.Fatalf("readFrame error = %v, want %v", err, ErrProtocolError)
			}
		})
	}
}

func TestReadFrameAcceptsValidControlAndDataFrames(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantOp  byte
		wantFin bool
		want    []byte
	}{
		{
			name:    "text",
			data:    maskedFrame(OpcodeText, true, []byte("hello")),
			wantOp:  OpcodeText,
			wantFin: true,
			want:    []byte("hello"),
		},
		{
			name:    "binary",
			data:    maskedFrame(OpcodeBinary, true, []byte{1, 2, 3}),
			wantOp:  OpcodeBinary,
			wantFin: true,
			want:    []byte{1, 2, 3},
		},
		{
			name:    "ping",
			data:    maskedFrame(opcodePing, true, []byte("p")),
			wantOp:  opcodePing,
			wantFin: true,
			want:    []byte("p"),
		},
		{
			name:    "pong",
			data:    maskedFrame(opcodePong, true, []byte("p")),
			wantOp:  opcodePong,
			wantFin: true,
			want:    []byte("p"),
		},
		{
			name:    "close",
			data:    maskedFrame(opcodeClose, true, []byte{0x03, 0xE8}),
			wantOp:  opcodeClose,
			wantFin: true,
			want:    []byte{0x03, 0xE8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newProtocolTestConn(tt.data)
			op, fin, payload, err := conn.readFrame()
			if err != nil {
				t.Fatalf("readFrame error: %v", err)
			}
			if op != tt.wantOp || fin != tt.wantFin || !bytes.Equal(payload, tt.want) {
				t.Fatalf("readFrame = (%d, %v, %v), want (%d, %v, %v)", op, fin, payload, tt.wantOp, tt.wantFin, tt.want)
			}
		})
	}
}

func TestReadMessageRejectsUnexpectedContinuation(t *testing.T) {
	conn := newProtocolTestConn(maskedFrame(opcodeContinuation, true, []byte("orphan")))

	_, _, err := conn.ReadMessage()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("ReadMessage error = %v, want %v", err, ErrProtocolError)
	}
}

func TestReadMessageAcceptsFragmentedText(t *testing.T) {
	data := append(maskedFrame(OpcodeText, false, []byte("hel")), maskedFrame(opcodeContinuation, true, []byte("lo"))...)
	conn := newProtocolTestConn(data)

	op, payload, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage error: %v", err)
	}
	if op != OpcodeText || string(payload) != "hello" {
		t.Fatalf("ReadMessage = (%d, %q), want text hello", op, string(payload))
	}
}

func TestReadMessageRejectsNewDataFrameDuringFragment(t *testing.T) {
	data := append(maskedFrame(OpcodeText, false, []byte("hel")), maskedFrame(OpcodeText, true, []byte("lo"))...)
	conn := newProtocolTestConn(data)

	_, _, err := conn.ReadMessage()
	if !errors.Is(err, ErrProtocolError) {
		t.Fatalf("ReadMessage error = %v, want %v", err, ErrProtocolError)
	}
}

func newProtocolTestConn(data []byte) *Conn {
	var writes bytes.Buffer
	conn := &Conn{
		br:     bufio.NewReader(bytes.NewReader(data)),
		bw:     bufio.NewWriter(&writes),
		closeC: make(chan struct{}),
	}
	conn.readLimit.Store(16 << 20)
	conn.pingPeriod.Store(int64(defaultPingPeriod))
	conn.pongWait.Store(int64(defaultPongWait))
	return conn
}

func maskedFrame(op byte, fin bool, payload []byte) []byte {
	first := op
	if fin {
		first |= finBit
	}
	switch {
	case len(payload) <= 125:
		return maskedFrameWithHeader(first, byte(len(payload)), nil, payload)
	case len(payload) <= 0xFFFF:
		var ext [2]byte
		binary.BigEndian.PutUint16(ext[:], uint16(len(payload)))
		return maskedFrameWithHeader(first, 126, ext[:], payload)
	default:
		return maskedFrameWithHeader(first, 127, uint64Bytes(uint64(len(payload))), payload)
	}
}

func maskedFrameWithHeader(first, prefix byte, ext, payload []byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(first)
	buf.WriteByte(0x80 | prefix)
	buf.Write(ext)
	mask := [4]byte{1, 2, 3, 4}
	buf.Write(mask[:])
	for i, b := range payload {
		buf.WriteByte(b ^ mask[i%len(mask)])
	}
	return buf.Bytes()
}

func uint64Bytes(v uint64) []byte {
	var ext [8]byte
	binary.BigEndian.PutUint64(ext[:], v)
	return ext[:]
}

func TestReadFrameStillRejectsUnmaskedFrames(t *testing.T) {
	conn := &Conn{
		br:     bufio.NewReader(bytes.NewReader([]byte{finBit | OpcodeText, 0})),
		bw:     bufio.NewWriter(io.Discard),
		closeC: make(chan struct{}),
	}
	conn.readLimit.Store(16 << 20)

	_, _, _, err := conn.readFrame()
	if !errors.Is(err, ErrUnmaskedFrame) {
		t.Fatalf("readFrame error = %v, want %v", err, ErrUnmaskedFrame)
	}
}
