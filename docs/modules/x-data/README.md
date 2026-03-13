# x/data

`x/data` is the extension boundary for high-level data topology that should not
remain in the stable `store` layer.

Use this module for:

- read-write splitting orchestration
- sharding and routing topology
- tenant-aware data topology composition

Do not use this module for:

- base `store` primitives
- HTTP handlers or middleware
- business repositories
