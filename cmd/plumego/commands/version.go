package commands

import (
	"flag"
	"fmt"
	"io"
	"runtime"
)

type VersionCmd struct {
	Version   string
	GitCommit string
	BuildDate string
}

func (c *VersionCmd) Name() string  { return "version" }
func (c *VersionCmd) Short() string { return "Show version information" }

func (c *VersionCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}
	if len(positionals) > 0 {
		return ctx.Out.Error(fmt.Sprintf("unexpected arguments: %v", positionals), 1)
	}

	return ctx.Out.Success("Plumego CLI", map[string]any{
		"version":    c.Version,
		"git_commit": c.GitCommit,
		"build_date": c.BuildDate,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
	})
}
