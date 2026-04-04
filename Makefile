# plumego — root Makefile
# Minimal targets. Most work happens via codex --yolo or go toolchain directly.

.PHONY: help milestone gates fmt vet test test-race

# Default: show help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# ── Milestone Runner ──────────────────────────────────────────────────────────

## Run a milestone spec through codex --yolo.
## Usage: make milestone M=active/M-001
##   M must be a path relative to tasks/milestones/, without the .md extension.
milestone: ## Run a milestone spec: make milestone M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make milestone M=active/M-001"; \
	  exit 1; \
	fi
	@SPEC=tasks/milestones/$(M).md; \
	if [ ! -f "$$SPEC" ]; then \
	  echo "Error: $$SPEC not found."; \
	  echo "Available specs:"; \
	  ls tasks/milestones/active/*.md 2>/dev/null | sed 's|tasks/milestones/||; s|\.md||'; \
	  exit 1; \
	fi
	@echo "Launching codex --yolo on $$SPEC ..."
	codex --yolo "$(shell cat tasks/milestones/$(M).md)"

# ── Quality Gates (run locally, mirrors CI) ───────────────────────────────────

gates: ## Run all required quality gates (mirrors CI)
	go run ./internal/checks/dependency-rules
	go run ./internal/checks/agent-workflow
	go run ./internal/checks/module-manifests
	go run ./internal/checks/reference-layout
	go vet ./...
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
	  echo "Unformatted files:"; echo "$$UNFORMATTED"; exit 1; \
	fi
	go test -race -timeout 60s ./...
	go test -timeout 20s ./...
	@echo "All gates passed."

fmt: ## Format all Go source files in-place
	gofmt -w .

vet: ## Run go vet on all packages
	go vet ./...

test: ## Run tests (standard timeout)
	go test -timeout 20s ./...

test-race: ## Run tests with race detector
	go test -race -timeout 60s ./...
