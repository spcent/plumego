#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

required_paths=(
  "README.md"
  "README_CN.md"
  "docs/CANONICAL_STYLE_GUIDE.md"
  "docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md"
  "specs/agent-entrypoints.yaml"
  "specs/ownership.yaml"
  "specs/dependency-rules.yaml"
  "env.example"
  "reference/standard-service/README.md"
  "reference/standard-service/main.go"
  "reference/standard-service/internal/app/app.go"
  "reference/standard-service/internal/app/routes.go"
)

for path in "${required_paths[@]}"; do
  if [[ ! -e "$path" ]]; then
    echo "missing canonical doc asset: $path" >&2
    exit 1
  fi
done

echo "canonical docs drift check: PASS"
