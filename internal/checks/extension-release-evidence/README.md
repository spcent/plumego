# extension-release-evidence

`extension-release-evidence` compares exported extension APIs across two git
refs.

It extracts temporary source trees for the selected refs, generates
deterministic exported API snapshots with `extension-api-snapshot`, and compares
the snapshots. It reports whether the public API is unchanged, but it does not
update `specs/extension-beta-evidence.yaml` or promote modules.

Example:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/rest/... \
  -base v1.0.0-rc.1 \
  -head HEAD \
  -out-dir /tmp/plumego-x-rest-release-evidence
```

For local smoke testing, comparing a ref with itself should report unchanged:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/rest/... \
  -base HEAD \
  -head HEAD
```

The refs used for a real promotion should already contain the
`extension-api-snapshot` tool, because snapshots are generated inside the
temporary source tree for each ref.
