package executil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const DefaultOutputLimit = 128 * 1024

type Options struct {
	Name        string
	Args        []string
	Dir         string
	Env         []string
	Timeout     time.Duration
	OutputLimit int
}

type Result struct {
	Stdout          string
	Stderr          string
	StdoutTruncated bool
	StderrTruncated bool
	TimedOut        bool
}

func (r Result) CombinedOutput() string {
	if r.Stdout == "" {
		return r.Stderr
	}
	if r.Stderr == "" {
		return r.Stdout
	}
	return r.Stdout + "\n" + r.Stderr
}

func (r Result) OutputTruncated() bool {
	return r.StdoutTruncated || r.StderrTruncated
}

func Run(ctx context.Context, opts Options) (Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Name == "" {
		return Result{}, fmt.Errorf("command name is required")
	}
	limit := opts.OutputLimit
	if limit <= 0 {
		limit = DefaultOutputLimit
	}

	runCtx := ctx
	cancel := func() {}
	if opts.Timeout > 0 {
		runCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
	}
	defer cancel()

	cmd := exec.CommandContext(runCtx, opts.Name, opts.Args...)
	cmd.Dir = opts.Dir
	if opts.Env != nil {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

	stdout := newLimitedBuffer(limit)
	stderr := newLimitedBuffer(limit)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	result := Result{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		TimedOut:        runCtx.Err() == context.DeadlineExceeded,
	}
	if result.TimedOut && err != nil {
		err = fmt.Errorf("command timed out after %s: %w", opts.Timeout, err)
	}
	return result, err
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated int
}

func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - b.buf.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		_, _ = b.buf.Write(p[:remaining])
	}
	if extra := len(p) - max(remaining, 0); extra > 0 {
		b.truncated += extra
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	out := b.buf.String()
	if b.truncated > 0 {
		out += fmt.Sprintf("\n[plumego: output truncated after %d bytes; %d bytes omitted]\n", b.limit, b.truncated)
	}
	return out
}

func (b *limitedBuffer) Truncated() bool {
	return b.truncated > 0
}
