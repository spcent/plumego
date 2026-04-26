# extension-maturity

`extension-maturity` validates `docs/EXTENSION_MATURITY.md` against extension
module manifests, machine-readable maturity signals, and beta-promotion
evidence.

Run the drift check:

```bash
go run ./internal/checks/extension-maturity
```

Print deterministic source data for review:

```bash
go run ./internal/checks/extension-maturity -report
```

The check verifies that each declared `x/*` root has a dashboard row with:

- current status, risk, and owner from its `module.yaml`
- recommended entrypoint, docs signal, and coverage signal from
  `specs/extension-maturity.yaml`
- beta evidence link and expected blocker text from
  `specs/extension-beta-evidence.yaml`, when the module is a beta candidate
