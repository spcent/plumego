---
name: Bug report
about: Report a reproducible bug in Plumego
title: "[Bug] "
labels: bug
assignees: ''
---

<!-- Thank you for reporting a bug. Please fill in the sections below.
     A minimal, runnable reproduction dramatically speeds up a fix. -->

## Description

What happened? What did you expect to happen?

## Reproduction

Minimal `main.go` (or a link to a repro repo):

```go
package main

import "fmt"

func main() {
	// minimal repro
	fmt.Println("replace with a real repro")
}
```

Steps:
1. …
2. …

## Environment

- Go version: `go version`
- OS / arch:
- Plumego version: `v1.1.0` / commit `…`
- Using stable roots only, or `x/*` extensions (which)?

## Logs / Evidence

Paste relevant output, panic traces, or race-detector reports here.

## Impact

- [ ] Blocks my service at startup
- [ ] Breaks a request path at runtime
- [ ] Causes a data race / crash
- [ ] Incorrect behaviour (no crash)
- [ ] Performance regression
- [ ] Documentation is wrong or misleading

## Related

- [ ] This affects a stable root (`core` `router` `contract` `middleware` `security` `store` `health` `log` `metrics`)
- [ ] This affects an `x/*` extension (name: `…`)
- [ ] I have a candidate fix in mind
