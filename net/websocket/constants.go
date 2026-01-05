package websocket

import "time"

// WebSocket protocol constants
const (
	guid                     = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	opcodeContinuation byte  = 0x0
	OpcodeText         byte  = 0x1
	OpcodeBinary       byte  = 0x2
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
