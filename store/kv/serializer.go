package kvstore

import (
	"bufio"
	"io"
)

// SerializationFormat represents the serialization format
type SerializationFormat string

const (
	// FormatBinary uses custom binary encoding for best performance
	FormatBinary SerializationFormat = "binary"
	// FormatJSON uses JSON encoding for readability and compatibility
	FormatJSON SerializationFormat = "json"
)

// Serializer defines the interface for serialization strategies
type Serializer interface {
	// WAL serialization
	EncodeWALEntry(entry WALEntry) ([]byte, error)
	DecodeWALEntry(reader *bufio.Reader) (*WALEntry, error)

	// Snapshot serialization
	WriteSnapshotHeader(writer io.Writer) error
	ReadSnapshotHeader(reader io.Reader) error
	EncodeEntry(entry *Entry) ([]byte, error)
	DecodeEntry(reader *bufio.Reader) (*Entry, error)

	// Format identification
	Format() SerializationFormat
}

// GetSerializer returns the appropriate serializer for the given format
func GetSerializer(format SerializationFormat) Serializer {
	switch format {
	case FormatJSON:
		return &JSONSerializer{}
	case FormatBinary:
		return &BinarySerializer{}
	default:
		// Default to binary for best performance
		return &BinarySerializer{}
	}
}

// DetectFormat attempts to detect the serialization format from file content
// Returns the detected format and resets the reader to the beginning
func DetectFormat(reader io.ReadSeeker) (SerializationFormat, error) {
	// Read first 4 bytes to check magic number
	magic := make([]byte, 4)
	n, err := reader.Read(magic)
	if err != nil && err != io.EOF {
		return FormatBinary, err
	}

	// Reset reader to beginning
	if _, err := reader.Seek(0, 0); err != nil {
		return FormatBinary, err
	}

	// If we couldn't read 4 bytes, assume JSON
	if n < 4 {
		return FormatJSON, nil
	}

	// Check for binary magic number (0x4B565354 = "KVST")
	if magic[0] == 0x4B && magic[1] == 0x56 && magic[2] == 0x53 && magic[3] == 0x54 {
		return FormatBinary, nil
	}

	// Check for JSON (starts with '{')
	if magic[0] == '{' {
		return FormatJSON, nil
	}

	// Default to binary
	return FormatBinary, nil
}

// DetectWALFormat attempts to detect the WAL serialization format
func DetectWALFormat(reader io.ReadSeeker) (SerializationFormat, error) {
	// Read first 4 bytes to check magic number
	magic := make([]byte, 4)
	n, err := reader.Read(magic)
	if err != nil && err != io.EOF {
		return FormatBinary, err
	}

	// Reset reader to beginning
	if _, err := reader.Seek(0, 0); err != nil {
		return FormatBinary, err
	}

	// If we couldn't read 4 bytes, assume JSON
	if n < 4 {
		return FormatJSON, nil
	}

	// Check for WAL binary magic number (0x57414C45 = "WALE")
	if magic[0] == 0x57 && magic[1] == 0x41 && magic[2] == 0x4C && magic[3] == 0x45 {
		return FormatBinary, nil
	}

	// Check for JSON (starts with '{')
	if magic[0] == '{' {
		return FormatJSON, nil
	}

	// Default to binary
	return FormatBinary, nil
}
