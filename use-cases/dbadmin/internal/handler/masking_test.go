package handler

import (
	"reflect"
	"testing"
)

func TestMaskedColumnSet(t *testing.T) {
	if set := maskedColumnSet(nil); set != nil {
		t.Errorf("maskedColumnSet(nil) = %v, want nil", set)
	}
	if set := maskedColumnSet([]string{}); set != nil {
		t.Errorf("maskedColumnSet([]) = %v, want nil", set)
	}
	set := maskedColumnSet([]string{"Password", "SSN"})
	if !set["password"] || !set["ssn"] {
		t.Errorf("maskedColumnSet did not lowercase keys: %v", set)
	}
}

func TestIsMaskedColumn(t *testing.T) {
	set := maskedColumnSet([]string{"Password", "api_key"})
	cases := []struct {
		col  string
		want bool
	}{
		{"password", true},
		{"PASSWORD", true},
		{"Password", true},
		{"api_key", true},
		{"API_KEY", true},
		{"username", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isMaskedColumn(c.col, set); got != c.want {
			t.Errorf("isMaskedColumn(%q) = %v, want %v", c.col, got, c.want)
		}
	}
	if isMaskedColumn("password", nil) {
		t.Error("isMaskedColumn with nil set should always be false")
	}
}

func TestMaskRow(t *testing.T) {
	cases := []struct {
		name    string
		columns []string
		row     []any
		masked  []string
		want    []any
	}{
		{
			name:    "case-insensitive single column",
			columns: []string{"id", "Password"},
			row:     []any{1, "secret"},
			masked:  []string{"password"},
			want:    []any{1, MaskedValuePlaceholder},
		},
		{
			name:    "multiple masked columns",
			columns: []string{"id", "password", "ssn", "name"},
			row:     []any{1, "secret", "123-45-6789", "Alice"},
			masked:  []string{"password", "ssn"},
			want:    []any{1, MaskedValuePlaceholder, MaskedValuePlaceholder, "Alice"},
		},
		{
			name:    "columns not in masked list pass through untouched",
			columns: []string{"id", "name", "email"},
			row:     []any{1, "Alice", "alice@example.com"},
			masked:  []string{"password"},
			want:    []any{1, "Alice", "alice@example.com"},
		},
		{
			name:    "empty masked list returns row unchanged",
			columns: []string{"id", "password"},
			row:     []any{1, "secret"},
			masked:  nil,
			want:    []any{1, "secret"},
		},
		{
			name:    "masks regardless of original type",
			columns: []string{"id", "credit_card", "balance"},
			row:     []any{1, 1234567890123456, 99.95},
			masked:  []string{"credit_card"},
			want:    []any{1, MaskedValuePlaceholder, 99.95},
		},
		{
			name:    "masks nil values too",
			columns: []string{"id", "api_key"},
			row:     []any{1, nil},
			masked:  []string{"api_key"},
			want:    []any{1, MaskedValuePlaceholder},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := maskRow(c.columns, c.row, maskedColumnSet(c.masked))
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("maskRow() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestMaskRow_doesNotMutateInput(t *testing.T) {
	row := []any{1, "secret"}
	original := append([]any{}, row...)
	_ = maskRow([]string{"id", "password"}, row, maskedColumnSet([]string{"password"}))
	if !reflect.DeepEqual(row, original) {
		t.Errorf("maskRow mutated input row: got %v, want %v", row, original)
	}
}

func TestMaskRowMap(t *testing.T) {
	cases := []struct {
		name   string
		row    map[string]any
		masked []string
		want   map[string]any
	}{
		{
			name:   "case-insensitive masking",
			row:    map[string]any{"id": 1, "Email": "a@example.com"},
			masked: []string{"email"},
			want:   map[string]any{"id": 1, "Email": MaskedValuePlaceholder},
		},
		{
			name:   "multiple masked columns",
			row:    map[string]any{"password": "p", "ssn": "s", "name": "n"},
			masked: []string{"password", "ssn"},
			want:   map[string]any{"password": MaskedValuePlaceholder, "ssn": MaskedValuePlaceholder, "name": "n"},
		},
		{
			name:   "columns not masked pass through",
			row:    map[string]any{"id": 1, "name": "Alice"},
			masked: []string{"password"},
			want:   map[string]any{"id": 1, "name": "Alice"},
		},
		{
			name:   "empty masked list is a no-op",
			row:    map[string]any{"password": "secret"},
			masked: nil,
			want:   map[string]any{"password": "secret"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := maskRowMap(c.row, maskedColumnSet(c.masked))
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("maskRowMap() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestMaskRowMaps(t *testing.T) {
	rows := []map[string]any{
		{"id": 1, "password": "secret1"},
		{"id": 2, "password": "secret2"},
	}
	maskRowMaps(rows, maskedColumnSet([]string{"password"}))
	want := []map[string]any{
		{"id": 1, "password": MaskedValuePlaceholder},
		{"id": 2, "password": MaskedValuePlaceholder},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("maskRowMaps() = %v, want %v", rows, want)
	}
}
