# extension-beta-evidence

`extension-beta-evidence` validates the beta-promotion evidence ledger in
`specs/extension-beta-evidence.yaml`.

It checks that each candidate:

- is declared as an extension root in `specs/repo.yaml`
- matches the owner and status in the target `module.yaml`
- points to an existing evidence document
- lists release refs that resolve to git commits when refs are present
- lists API snapshots that exist under `docs/extension-evidence/snapshots/`
  when snapshots are present
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

The ledger supports three candidate shapes:

- `candidates` for root extension modules
- `subpackage_candidates` for stable-tier subpackages inside a root module
- `surface_candidates` for smaller package or feature surfaces that should not
  imply root-module promotion
