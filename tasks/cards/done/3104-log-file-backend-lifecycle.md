# Card 3104: Log File Backend Lifecycle Tightening

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: log
Owned Files:
- `log/glog.go`
- `log/glog_test.go`
Depends On:
- `tasks/cards/done/3101-log-callsite-depth.md`

Goal:
Make the internal text file backend safer around initialization failures,
closing, and rotation bookkeeping.

Problem:
- `initLogFiles` could return after opening some files without closing the
  partial set on later failures.
- `Close` mutated file state without coordinating with the write mutex, so
  concurrent logging could race with file closure.
- Rotation size accounting only tracked the triggering severity, while the file
  backend fans higher-severity log lines into lower-severity files.

Scope:
- Close partial file state when initialization fails.
- Coordinate `Close` with active writes.
- Track rotation size for every file that receives a fanned-out line.
- Add focused lifecycle/rotation tests.

Non-goals:
- Do not expose file logging or rotation as new stable public API.
- Do not replace the text backend with an external logging library.
- Do not change the `NewLogger` constructor contract.

Outcome:
- Added partial-initialization cleanup for files and size bookkeeping.
- Initialized `currentSize` from opened file sizes and cleared it on close.
- Coordinated `Close` with `writeMu` before clearing file state.
- Added fan-out level detection so rotation accounting follows every file that
  receives a higher-severity log line.
- Added tests for symlink-failure cleanup, fan-out rotation accounting, and
  close/write coordination.

Validation:
- `go test -race -timeout 60s ./log/...`
- `go test -timeout 20s ./log/...`
- `go vet ./log/...`

Done Definition:
- Partial initialization failures do not leave open files in the backend map.
- Closing is safe with concurrent logging.
- Rotation bookkeeping reflects every file that receives fanned-out output.
- The listed validation commands pass.
