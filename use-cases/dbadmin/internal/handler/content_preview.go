package handler

import (
	"encoding/hex"
	"fmt"
	"unicode/utf8"
)

// Default limits for content previewing.
const (
	// DefaultMaxTextPreviewBytes is the maximum size of a text/string value
	// returned in a row result. Values longer than this are truncated with a marker.
	DefaultMaxTextPreviewBytes = 4096 // 4 KiB

	// DefaultMaxBlobPreviewBytes is the hex preview length for BLOB values.
	DefaultMaxBlobPreviewBytes = 64

	// TruncationMarker is appended to truncated text values.
	TruncationMarker = "… [truncated]"
)

// PreviewLimits controls how large values are previewed in query results.
type PreviewLimits struct {
	MaxTextBytes int // truncate text strings longer than this
	MaxBlobBytes int // hex preview length for BLOB values
}

// DefaultPreviewLimits returns the default preview limits.
func DefaultPreviewLimits() PreviewLimits {
	return PreviewLimits{
		MaxTextBytes: DefaultMaxTextPreviewBytes,
		MaxBlobBytes: DefaultMaxBlobPreviewBytes,
	}
}

// PreviewResult is the outcome of previewing a single value.
type PreviewResult struct {
	Value     any
	Truncated bool
}

// PreviewValue applies preview limits to a single scanned value.
// Returns the (possibly truncated) value and whether truncation occurred.
func (p PreviewLimits) PreviewValue(v any) PreviewResult {
	if v == nil {
		return PreviewResult{Value: nil}
	}

	switch val := v.(type) {
	case []byte:
		if len(val) == 0 {
			return PreviewResult{Value: ""}
		}
		if utf8.Valid(val) {
			return p.previewString(string(val))
		}
		// Binary BLOB: show hex preview with size marker
		previewLen := p.MaxBlobBytes
		if previewLen <= 0 {
			previewLen = DefaultMaxBlobPreviewBytes
		}
		end := previewLen
		if end > len(val) {
			end = len(val)
		}
		preview := hex.EncodeToString(val[:end])
		marker := ""
		if len(val) > previewLen {
			marker = "…"
		}
		return PreviewResult{
			Value:     fmt.Sprintf("<BLOB %d bytes|%s%s>", len(val), preview, marker),
			Truncated: len(val) > previewLen,
		}

	case string:
		return p.previewString(val)

	default:
		return PreviewResult{Value: val}
	}
}

func (p PreviewLimits) previewString(s string) PreviewResult {
	maxLen := p.MaxTextBytes
	if maxLen <= 0 {
		maxLen = DefaultMaxTextPreviewBytes
	}
	if len(s) <= maxLen {
		return PreviewResult{Value: s}
	}
	// Truncate at byte boundary, ensure we don't split a multi-byte rune.
	truncated := s[:maxLen]
	for len(truncated) > 0 && !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return PreviewResult{
		Value:     truncated + TruncationMarker,
		Truncated: true,
	}
}

// ApplyPreviewToRow applies preview limits to every value in a row map.
// Returns the row with values potentially truncated and a bool indicating
// whether any value was truncated.
func (p PreviewLimits) ApplyPreviewToRow(row map[string]any) (map[string]any, bool) {
	anyTruncated := false
	for k, v := range row {
		result := p.PreviewValue(v)
		row[k] = result.Value
		if result.Truncated {
			anyTruncated = true
		}
	}
	return row, anyTruncated
}

// ApplyPreviewToRows applies preview limits to all rows.
// Returns whether any value in any row was truncated.
func (p PreviewLimits) ApplyPreviewToRows(rows []map[string]any) bool {
	anyTruncated := false
	for _, row := range rows {
		if _, truncated := p.ApplyPreviewToRow(row); truncated {
			anyTruncated = true
		}
	}
	return anyTruncated
}

// TruncateJSONValue recursively truncates large string and binary values in
// a JSON-like data structure (maps, slices, strings, []byte). Maps and slices
// are modified in place. Non-string/binary values are returned as-is.
// This is used by MongoDB, Redis, and Elasticsearch handlers where the
// result is already a map[string]any or []any structure.
func (p PreviewLimits) TruncateJSONValue(v any) (any, bool) {
	switch val := v.(type) {
	case string:
		result := p.PreviewValue(val)
		return result.Value, result.Truncated

	case []byte:
		result := p.PreviewValue(val)
		return result.Value, result.Truncated

	case map[string]any:
		anyTruncated := false
		for k, child := range val {
			truncated, truncatedFlag := p.TruncateJSONValue(child)
			val[k] = truncated
			if truncatedFlag {
				anyTruncated = true
			}
		}
		return val, anyTruncated

	case []any:
		anyTruncated := false
		for i, child := range val {
			truncated, truncatedFlag := p.TruncateJSONValue(child)
			val[i] = truncated
			if truncatedFlag {
				anyTruncated = true
			}
		}
		return val, anyTruncated

	default:
		return v, false
	}
}
