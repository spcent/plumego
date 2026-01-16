package websocket

import "time"

// WebSocket protocol constants
const (
	guid                     = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	opcodeContinuation byte  = 0x0
	opcodeText         byte  = 0x1
	opcodeBinary       byte  = 0x2
	opcodeClose        byte  = 0x8
	opcodePing         byte  = 0x9
	opcodePong         byte  = 0xA
	finBit             byte  = 0x80
	defaultBufSize     int   = 4096
	defaultPingPeriod        = 20 * time.Second
	defaultPongWait          = 30 * time.Second
	maxControlPayload  int64 = 125
	maxFragmentSize    int   = 64 * 1024 // 64KB
)

// Opcode constants for WebSocket frame types.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	// Send a text message
//	err := conn.WriteMessage(websocket.OpcodeText, []byte("Hello"))
//
//	// Send a binary message
//	err := conn.WriteMessage(websocket.OpcodeBinary, []byte{0x01, 0x02, 0x03})
const (
	// OpcodeText is used for UTF-8 encoded text messages
	OpcodeText byte = 0x1

	// OpcodeBinary is used for binary data messages
	OpcodeBinary byte = 0x2
)
