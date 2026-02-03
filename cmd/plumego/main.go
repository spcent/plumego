package main

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/cmd/plumego/commands"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

func main() {
	if err := commands.Execute(); err != nil {
		if code, ok := output.ExitCode(err); ok {
			os.Exit(code)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
