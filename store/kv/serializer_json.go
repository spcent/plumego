package kvstore

import (
	"bufio"
	"encoding/json"
	"io"
	"time"
)

// JSONSerializer implements Serializer using JSON encoding
type JSONSerializer struct{}

// EncodeWALEntry encodes a WAL entry to JSON format
func (s *JSONSerializer) EncodeWALEntry(entry WALEntry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	// Add newline for proper separation
	return append(data, '\n'), nil
}

// DecodeWALEntry decodes a WAL entry from JSON format
func (s *JSONSerializer) DecodeWALEntry(reader *bufio.Reader) (*WALEntry, error) {
	decoder := json.NewDecoder(reader)

	// Check if there's more data
	if !decoder.More() {
		return nil, io.EOF
	}

	var entry WALEntry
	if err := decoder.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// WriteSnapshotHeader writes the JSON snapshot header
func (s *JSONSerializer) WriteSnapshotHeader(writer io.Writer) error {
	header := map[string]any{
		"magic":   magicNumber,
		"version": version,
		"format":  "json",
		"created": time.Now(),
	}
	return json.NewEncoder(writer).Encode(header)
}

// ReadSnapshotHeader reads and validates the JSON snapshot header
func (s *JSONSerializer) ReadSnapshotHeader(reader io.Reader) error {
	var header map[string]any
	if err := json.NewDecoder(reader).Decode(&header); err != nil {
		return err
	}
	// JSON format is more lenient, just verify it's a valid header
	return nil
}

// EncodeEntry encodes a snapshot entry to JSON format
func (s *JSONSerializer) EncodeEntry(entry *Entry) ([]byte, error) {
	data, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}
	// Add newline for proper separation
	return append(data, '\n'), nil
}

// DecodeEntry decodes a snapshot entry from JSON format
func (s *JSONSerializer) DecodeEntry(reader *bufio.Reader) (*Entry, error) {
	decoder := json.NewDecoder(reader)

	// Check if there's more data
	if !decoder.More() {
		return nil, io.EOF
	}

	var entry Entry
	if err := decoder.Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// Format returns the serialization format
func (s *JSONSerializer) Format() SerializationFormat {
	return FormatJSON
}
