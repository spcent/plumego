package jsonx

import "testing"

func TestFieldString(t *testing.T) {
	raw := []byte(`{"id":"evt_1","livemode":true,"obj":{"name":"x"}}`)
	if FieldString(raw, "id") != "evt_1" {
		t.Fatalf("id mismatch")
	}
	if FieldBool(raw, "livemode") != true {
		t.Fatalf("livemode mismatch")
	}
	if PathString(raw, "obj", "name") != "x" {
		t.Fatalf("path mismatch")
	}
}
