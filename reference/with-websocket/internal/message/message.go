// Package message defines the typed JSON envelope for the WebSocket echo demo.
package message

import "encoding/json"

// Event is the JSON message exchanged over the WebSocket connection.
//
// The demo handler responds to {"type":"ping"} with {"type":"pong"}.
// All other text frames and all binary frames are echoed verbatim.
type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// Decode parses data as a JSON-encoded Event. Returns an error if data is
// not valid JSON or cannot be decoded into Event.
func Decode(data []byte) (Event, error) {
	var ev Event
	if err := json.Unmarshal(data, &ev); err != nil {
		return Event{}, err
	}
	return ev, nil
}

// Encode serializes ev to JSON.
func Encode(ev Event) ([]byte, error) {
	return json.Marshal(ev)
}
