# Card 0600: Log Call-Site Depth Regression Coverage

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: log
Owned Files:
- `log/callsite_helper_test.go`
- `log/glog_test.go`
- `log/logger_semantics_test.go`
Depends On: none

Goal:
Make text logger headers and verbosity/vmodule checks consistently resolve the
caller that invoked the logging API, not an internal wrapper method.

Problem:
The text backend relies on hard-coded caller-depth values across the structured
text logger, the internal `gLogger`, and test-only default wrappers. Before this
card, the behavior was only indirectly covered by format tests, so a one-frame
change could silently move output from the direct caller to an outer wrapper or
testing helper.

Analysis Result:
The initial audit suspected the existing depths were one frame too shallow. A
focused test run showed the opposite: increasing the depth made headers report
the caller of the helper function instead of the logging call site, and broke
existing `vmodule` semantics. The implementation constants are intentionally
unchanged; this card adds coverage that prevents that regression.

Scope:
- Add focused tests that distinguish wrapper files from caller files.
- Cover structured text logging, internal instance logging, default wrapper
  logging, and `vmodule` caller resolution.

Non-goals:
- Do not change the public `StructuredLogger` interface.
- Do not expose glog-style default helpers.
- Do not add dependencies.

Outcome:
- Added `log/callsite_helper_test.go` to create distinguishable call-site
  frames.
- Added regression tests for structured text logger call-site output.
- Added regression tests for internal text logger and default-wrapper
  call-site output.
- Added regression tests for instance and default `vmodule` caller resolution.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Text log headers identify the direct logging caller for structured, instance,
  and default-wrapper logging paths.
- `vmodule` checks use the direct logging caller file for instance and default
  wrapper paths.
- The listed validation commands pass.
