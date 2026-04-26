# extension-maturity

`extension-maturity` validates `docs/EXTENSION_MATURITY.md` against extension
module manifests and beta-promotion evidence.

Run the drift check:

```bash
go run ./internal/checks/extension-maturity
```

Print deterministic source data for review:

```bash
go run ./internal/checks/extension-maturity -report
```

The check verifies that each declared `x/*` root has a dashboard row with the
current status, risk, and owner from its `module.yaml`. For modules listed in
`specs/extension-beta-evidence.yaml`, it also verifies the beta evidence link
and the expected blocker text for release history, API snapshots, and owner
sign-off.
