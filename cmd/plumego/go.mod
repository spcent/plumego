module github.com/spcent/plumego/cmd/plumego

go 1.24.7

// Use local plumego package
replace github.com/spcent/plumego => ../..

require (
	github.com/spcent/plumego v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
)
