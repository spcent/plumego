module github.com/spcent/plumego/examples/db-metrics

go 1.24.0

require (
	github.com/mattn/go-sqlite3 v1.14.33
	github.com/spcent/plumego v0.0.0
)

// Use local plumego instead of remote version
replace github.com/spcent/plumego => ../..
