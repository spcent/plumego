package storage

// Type identifies a backing store.
type Type string

const (
	// TypeMemory is the in-memory store. No durability.
	TypeMemory Type = "memory"
	// TypeSQLite is the SQLite-backed store. Requires Path.
	TypeSQLite Type = "sqlite"
)
