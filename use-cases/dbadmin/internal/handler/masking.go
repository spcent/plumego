package handler

import "strings"

// MaskedValuePlaceholder replaces the original value of a masked column
// regardless of its type. This is a display/export-time transformation only —
// the underlying stored data is never modified.
const MaskedValuePlaceholder = "***MASKED***"

// maskedColumnSet builds a case-insensitive lookup set from a list of column
// names. Returns nil when cols is empty, so callers can skip masking entirely
// with a single nil check.
func maskedColumnSet(cols []string) map[string]bool {
	if len(cols) == 0 {
		return nil
	}
	set := make(map[string]bool, len(cols))
	for _, c := range cols {
		set[strings.ToLower(c)] = true
	}
	return set
}

// isMaskedColumn reports whether col is in maskedSet, compared case-insensitively.
func isMaskedColumn(col string, maskedSet map[string]bool) bool {
	if len(maskedSet) == 0 {
		return false
	}
	return maskedSet[strings.ToLower(col)]
}

// maskRow replaces values in masked columns with MaskedValuePlaceholder,
// regardless of original type. Column name comparison is case-insensitive.
// Columns not present in maskedSet pass through untouched. The input slice
// is not mutated; a new slice is returned.
func maskRow(columns []string, row []any, maskedSet map[string]bool) []any {
	if len(maskedSet) == 0 {
		return row
	}
	out := make([]any, len(row))
	copy(out, row)
	for i, col := range columns {
		if i >= len(out) {
			break
		}
		if isMaskedColumn(col, maskedSet) {
			out[i] = MaskedValuePlaceholder
		}
	}
	return out
}

// maskRowMap replaces values for masked columns in a map[string]any row,
// in place, and returns the same map for convenience. Column name comparison
// is case-insensitive. No-op when maskedSet is empty.
func maskRowMap(row map[string]any, maskedSet map[string]bool) map[string]any {
	if len(maskedSet) == 0 {
		return row
	}
	for col := range row {
		if isMaskedColumn(col, maskedSet) {
			row[col] = MaskedValuePlaceholder
		}
	}
	return row
}

// maskRowMaps applies maskRowMap to every row in rows, in place.
func maskRowMaps(rows []map[string]any, maskedSet map[string]bool) {
	if len(maskedSet) == 0 {
		return
	}
	for _, row := range rows {
		maskRowMap(row, maskedSet)
	}
}
