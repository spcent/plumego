package kvengine

import (
	"bufio"
	"fmt"
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

	return detectSnapshotMagic(magic[:n])
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

	return detectWALMagic(magic[:n])
}

func detectSnapshotFormat(reader *bufio.Reader) (SerializationFormat, error) {
	magic, err := reader.Peek(4)
	if err != nil && err != io.EOF {
		return "", err
	}
	return detectSnapshotMagic(magic)
}

func detectSnapshotMagic(magic []byte) (SerializationFormat, error) {
	if len(magic) < 4 {
		return "", fmt.Errorf("unknown snapshot format: short header")
	}
	if magic[0] == 0x4B && magic[1] == 0x56 && magic[2] == 0x53 && magic[3] == 0x54 {
		return FormatBinary, nil
	}
	if magic[0] == '{' {
		return FormatJSON, nil
	}
	return "", fmt.Errorf("unknown snapshot format: magic 0x%X", magic[:4])
}

func detectWALMagic(magic []byte) (SerializationFormat, error) {
	if len(magic) < 4 {
		return "", fmt.Errorf("unknown WAL format: short header")
	}
	if magic[0] == 0x57 && magic[1] == 0x41 && magic[2] == 0x4C && magic[3] == 0x45 {
		return FormatBinary, nil
	}
	if magic[0] == '{' {
		return FormatJSON, nil
	}
	return "", fmt.Errorf("unknown WAL format: magic 0x%X", magic[:4])
}
