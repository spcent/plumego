package handler

import (
	"testing"

	"dbadmin/internal/domain/connection"
)

func TestSqlQuoteString(t *testing.T) {
	cases := []struct {
		in     string
		driver connection.DriverType
		want   string
	}{
		// Plain string — no escaping needed
		{"hello", connection.DriverSQLite, "'hello'"},
		{"hello", connection.DriverMySQL, "'hello'"},

		// Single-quote escaping (both drivers use the same SQL standard rule)
		{"it's", connection.DriverSQLite, "'it''s'"},
		{"it's", connection.DriverMySQL, "'it''s'"},

		// Backslash: SQLite treats \ as a literal — must NOT be doubled
		{`a\nb`, connection.DriverSQLite, `'a\nb'`},
		// Backslash: MySQL default mode treats \ as escape — MUST be doubled
		{`a\nb`, connection.DriverMySQL, `'a\\nb'`},

		// NULL byte (ensure it's passed through safely)
		{"null\x00byte", connection.DriverSQLite, "'null\x00byte'"},

		// Combined: backslash + single-quote in MySQL
		{`O\'Brien`, connection.DriverMySQL, `'O\\''Brien'`},
	}
	for _, c := range cases {
		got := sqlQuoteString(c.in, c.driver)
		if got != c.want {
			t.Errorf("sqlQuoteString(%q, %s)\n  got  %q\n  want %q",
				c.in, c.driver, got, c.want)
		}
	}
}
