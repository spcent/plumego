# x/ops

`x/ops` owns protected operational HTTP surfaces and explicit diagnostics.

Notable subpackages:

- `x/ops/healthhttp`: health, readiness, build-info, runtime, and history handlers

Stable `health` remains responsible for manager state and check primitives, not for HTTP endpoint ownership.
