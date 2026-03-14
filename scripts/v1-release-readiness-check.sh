#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

release_files=(
  "README.md"
  "README_CN.md"
  "AGENTS.md"
  "CLAUDE.md"
  "env.example"
  "reference/standard-service/README.md"
  "cmd/plumego/README.md"
  "cmd/plumego/go.mod"
)

for path in "${release_files[@]}"; do
  if [[ ! -e "$path" ]]; then
    echo "missing release asset: $path" >&2
    exit 1
  fi
done

module_manifests="$(find . -name module.yaml | sort)"
if [[ -z "$module_manifests" ]]; then
  echo "no module manifests found" >&2
  exit 1
fi

while IFS= read -r manifest; do
  [[ -n "$manifest" ]] || continue

  if ! rg -q '^status:' "$manifest"; then
    echo "module manifest missing status: $manifest" >&2
    exit 1
  fi

  if ! rg -q '^summary:' "$manifest"; then
    echo "module manifest missing summary: $manifest" >&2
    exit 1
  fi
done <<<"$module_manifests"

stable_roots=(
  "core/module.yaml"
  "router/module.yaml"
  "contract/module.yaml"
  "middleware/module.yaml"
  "security/module.yaml"
  "store/module.yaml"
  "health/module.yaml"
  "log/module.yaml"
  "metrics/module.yaml"
)

for manifest in "${stable_roots[@]}"; do
  if ! rg -qx 'status: ga' "$manifest"; then
    echo "stable root is not marked ga: $manifest" >&2
    exit 1
  fi
done

bash scripts/check-doc-api-drift.sh

echo "v1 release readiness aggregate check: PASS"
