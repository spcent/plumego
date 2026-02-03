package commands

import (
	"flag"
	"runtime"
)

// Version information (set at build time)
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

type VersionCmd struct{}

func (c *VersionCmd) Name() string {
	return "version"
}

func (c *VersionCmd) Short() string {
	return "Show version information"
}

func (c *VersionCmd) Long() string {
	return `Display version information for the plumego CLI.

Examples:
  plumego version
  plumego version --format json`
}

func (c *VersionCmd) Flags() []Flag {
	return nil
}

func (c *VersionCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("version", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}

	versionInfo := map[string]any{
		"version":    Version,
		"git_commit": GitCommit,
		"build_date": BuildDate,
		"go_version": runtime.Version(),
		"platform":   runtime.GOOS + "/" + runtime.GOARCH,
	}

	return ctx.Out.Success("Plumego CLI", versionInfo)
}
