# extension-beta-evidence

`extension-beta-evidence` validates the beta-promotion evidence ledger in
`specs/extension-beta-evidence.yaml`.

It checks that each candidate:

- is declared as an extension root in `specs/repo.yaml`
- matches the owner and status in the target `module.yaml`
- points to an existing evidence document
- has blockers that match the recorded release refs, API snapshots, and owner
  sign-off state

Run it with:

```bash
go run ./internal/checks/extension-beta-evidence
```

The command prints one deterministic report line per candidate. It exits nonzero
only when the ledger is internally inconsistent; current promotion blockers such
as missing release history, API snapshots, or owner sign-off are expected until a
promotion card resolves them.
