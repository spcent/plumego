package commands

import "github.com/spcent/plumego/cmd/plumego/internal/output"

// Context carries shared CLI dependencies and configuration.
type Context struct {
	Out     *output.Formatter
	EnvFile string
}
