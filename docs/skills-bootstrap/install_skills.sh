#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGET_ROOT="${CODEX_HOME:-$HOME/.codex}/skills"

SKILLS=(
  "plumego-task-slicer"
  "plumego-boundary-guard"
  "plumego-test-gate-runner"
  "plumego-style-enforcer"
  "plumego-doc-sync-checker"
  "plumego-release-readiness"
)

mkdir -p "$TARGET_ROOT"

for skill in "${SKILLS[@]}"; do
  src="$SCRIPT_DIR/$skill"
  dst="$TARGET_ROOT/$skill"
  if [[ ! -d "$src" ]]; then
    echo "skip: missing source directory $src" >&2
    continue
  fi

  rm -rf "$dst"
  cp -R "$src" "$dst"
  echo "installed: $skill -> $dst"
done

echo "all done. target root: $TARGET_ROOT"
