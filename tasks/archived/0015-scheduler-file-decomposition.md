# Card 0015

Priority: P2
State: done
Primary Module: x/scheduler
Owned Files:
  - x/scheduler/scheduler.go

Depends On: —

Goal:
`x/scheduler/scheduler.go` is a 1777-line single-file god object containing code with 4 distinct
responsibilities:

- **Public API layer** (Start/Stop/AddCron/Schedule/Delay/Cancel/Pause/Resume/List/QueryJobs etc., ~20 methods)
- **Execution engine** (runLoop/dispatch/worker/execute/handleFailure/scheduleNext, ~500 lines)
- **Queue and scheduling** (scheduleHeap, pushScheduleLocked, popDue, nextDue, ~150 lines)
- **Dependency graph** (wouldCreateDependencyCycleLocked/depPathExistsLocked/batch ops, ~200 lines)
- **Query helpers** (QueryJobs/sortJobStatuses/jobQueryMatcher/buildJobStatusLess, ~150 lines)

A single file cannot be read or modified locally; any change requires grepping the entire file to
locate context.

Split into multiple files by responsibility (**within the same package, without changing any type or
function signatures**):

| New File | Contents | Approximate Lines |
|----------|----------|-------------------|
| `scheduler.go` | Scheduler struct, Option, New(), Start/Stop | ~200 lines |
| `scheduler_api.go` | AddCron, Schedule, Delay, TriggerNow, Cancel, Pause, Resume and other public write methods | ~350 lines |
| `scheduler_query.go` | List, Status, QueryJobs, PruneTerminal, BatchByGroup/Tags | ~250 lines |
| `scheduler_executor.go` | runLoop, dispatch, worker, execute, handleFailure, scheduleNext | ~500 lines |
| `scheduler_queue.go` | scheduleHeap, pushScheduleLocked, popDue, nextDue, internal queue operations | ~150 lines |
| `scheduler_deps.go` | wouldCreateDependencyCycleLocked, depPathExistsLocked | ~100 lines |

Scope:
- **Move code to new files only, without changing any logic, signatures, or exported names**
- Remove the migrated portions from scheduler.go; the final scheduler.go retains only the struct
  definition, Option, and New()
- Keep all functions, types, and constants in the same package `scheduler`; no import changes needed
- Check and preserve the positions of all init-level variables and compile-time assertions

Non-goals:
- Do not change any public API (function signatures, type names, package name)
- Do not split into sub-packages
- Do not optimize algorithms or change behavior
- Do not change metrics.go, types.go, job.go, or other already-independent files

Files:
  - x/scheduler/scheduler.go (reduce to struct + Option + New)
  - x/scheduler/scheduler_api.go (new)
  - x/scheduler/scheduler_query.go (new)
  - x/scheduler/scheduler_executor.go (new)
  - x/scheduler/scheduler_queue.go (new)
  - x/scheduler/scheduler_deps.go (new)

Tests:
  - go build ./x/scheduler/...
  - go test -race -timeout 60s ./x/scheduler/...

Docs Sync: —

Done Definition:
- `wc -l x/scheduler/scheduler.go` ≤ 250 lines
- `go build ./x/scheduler/...` passes
- `go test -race ./x/scheduler/...` passes with no test changes (logic untouched)

Outcome:
- Split into 5 files: `scheduler.go` (229 lines, struct/options/New/Start/Stop), `scheduler_api.go` (public write methods: AddJob, RemoveJob, etc.), `scheduler_query.go` (read/query methods: GetJob, ListJobs, etc.), `scheduler_executor.go` (runLoop, dispatch, execute internals), `scheduler_deps.go` (dependency cycle detection helpers)
- All existing tests pass unchanged
