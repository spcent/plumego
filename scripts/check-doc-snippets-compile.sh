#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d "${TMPDIR:-/tmp}/plumego-doc-snippets.XXXXXX")"
trap 'rm -rf "$TMP"' EXIT

cat > "$TMP/go.mod" <<EOF
module github.com/spcent/plumego-doc-snippets

go 1.24.0

toolchain go1.24.4

require github.com/spcent/plumego v0.0.0

replace github.com/spcent/plumego => $ROOT
EOF

DOCS=(
	"README.md"
	"README_CN.md"
	"docs/getting-started.md"
	"docs/CANONICAL_STYLE_GUIDE.md"
	"docs/modules/core/README.md"
)

for doc in "${DOCS[@]}"; do
	path="$ROOT/$doc"
	if [[ ! -f "$path" ]]; then
		echo "missing doc: $doc" >&2
		exit 1
	fi

	awk -v tmp="$TMP" -v doc="$doc" '
		function sanitize(value) {
			gsub(/[^A-Za-z0-9]+/, "-", value)
			gsub(/^-+/, "", value)
			gsub(/-+$/, "", value)
			return value
		}
		BEGIN {
			in_go = 0
			snippet_index = 0
			id = sanitize(doc)
		}
		/^```go[[:space:]]*$/ {
			in_go = 1
			snippet = ""
			next
		}
		in_go && /^```[[:space:]]*$/ {
			in_go = 0
			if (snippet ~ /(^|\n)package[[:space:]]+main([[:space:]]|\n)/) {
				snippet_index++
				dir = tmp "/" id "-" snippet_index
				system("mkdir -p " dir)
				file = dir "/main.go"
				printf "%s", snippet > file
				close(file)
			}
			next
		}
		in_go {
			snippet = snippet $0 "\n"
		}
	' "$path"
done

count="$(find "$TMP" -mindepth 2 -maxdepth 2 -name main.go | wc -l | tr -d '[:space:]')"
if [[ "$count" == "0" ]]; then
	echo "no package-main Go documentation snippets found" >&2
	exit 1
fi

(cd "$TMP" && go test ./...)
echo "compiled $count package-main Go documentation snippets"
