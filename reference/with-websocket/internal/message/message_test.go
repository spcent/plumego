package message

import (
	"encoding/json"
	"testing"
)

func TestDecodeRoundtrip(t *testing.T) {
	ev := Event{Type: "ping", Data: json.RawMessage(`{"seq":1}`)}
	b, err := Encode(ev)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, err := Decode(b)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Type != ev.Type {
		t.Fatalf("Type = %q, want %q", got.Type, ev.Type)
	}
	if string(got.Data) != string(ev.Data) {
		t.Fatalf("Data = %s, want %s", got.Data, ev.Data)
	}
}

func TestDecodeFieldMapping(t *testing.T) {
	got, err := Decode([]byte(`{"type":"pong","data":{"ok":true}}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Type != "pong" {
		t.Fatalf("Type = %q, want %q", got.Type, "pong")
	}
	if string(got.Data) != `{"ok":true}` {
		t.Fatalf("Data = %s, want %s", got.Data, `{"ok":true}`)
	}
}

func TestDecodeEmptyType(t *testing.T) {
	got, err := Decode([]byte(`{}`))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if got.Type != "" {
		t.Fatalf("Type = %q, want empty", got.Type)
	}
}

func TestDecodeInvalidJSON(t *testing.T) {
	_, err := Decode([]byte(`not json`))
	if err == nil {
		t.Fatal("Decode did not return error for invalid JSON")
	}
}

func TestDecodeNonObjectJSON(t *testing.T) {
	_, err := Decode([]byte(`"just a string"`))
	if err == nil {
		t.Fatal("Decode did not return error for non-object JSON")
	}
}

func TestEncodeOmitsEmptyData(t *testing.T) {
	b, err := Encode(Event{Type: "pong"})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal encoded: %v", err)
	}
	if _, ok := m["data"]; ok {
		t.Fatalf("expected 'data' key to be omitted when empty, got %s", b)
	}
}
