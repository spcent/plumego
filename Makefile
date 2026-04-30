# plumego — root Makefile
# Minimal targets. Most work happens via codex --yolo or go toolchain directly.

.PHONY: help milestone check-spec check-plan check-card check-verify new-milestone new-plan new-card new-verify gates fmt vet test test-race setup-hooks

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
	PLAN=tasks/milestones/$${SPEC##*/}; \
	PLAN=$${PLAN%.md}.plan.md; \
	if [ ! -f "$$SPEC" ]; then \
	  echo "Error: $$SPEC not found."; \
	  echo "Available specs:"; \
	  ls tasks/milestones/active/*.md 2>/dev/null | sed 's|tasks/milestones/||; s|\.md||'; \
	  exit 1; \
	fi; \
	if [ ! -f "$$PLAN" ]; then \
	  echo "Error: $$PLAN not found."; \
	  echo "Run: make new-plan M=$(M)"; \
	  echo "Then fill the plan and validate it with: make check-plan M=$(M)"; \
	  exit 1; \
	fi; \
	echo "Validating spec before launch ..."; \
	scripts/check-spec "$$SPEC" || { echo "Fix spec errors above, then re-run."; exit 1; }; \
	echo "Validating plan before launch ..."; \
	scripts/check-spec "$$PLAN" || { echo "Fix plan errors above, then re-run."; exit 1; }; \
	echo "Launching codex --yolo on $$SPEC ..."; \
	codex --yolo "$$(cat "$$SPEC")"

check-spec: ## Validate a milestone spec: make check-spec M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make check-spec M=active/M-001"; \
	  exit 1; \
	fi
	@scripts/check-spec tasks/milestones/$(M).md

check-plan: ## Validate a milestone plan: make check-plan M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make check-plan M=active/M-001"; \
	  exit 1; \
	fi
	@ID=$${M##*/}; \
	PLAN=tasks/milestones/$$ID.plan.md; \
	if [ ! -f "$$PLAN" ]; then \
	  echo "Error: $$PLAN not found."; \
	  echo "Create it with: make new-plan M=$(M)"; \
	  exit 1; \
	fi; \
	scripts/check-spec "$$PLAN"

check-verify: ## Validate a milestone verify report: make check-verify M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make check-verify M=active/M-001"; \
	  exit 1; \
	fi
	@ID=$${M##*/}; \
	VERIFY=tasks/milestones/$$ID.verify.md; \
	if [ ! -f "$$VERIFY" ]; then \
	  echo "Error: $$VERIFY not found."; \
	  echo "Create it with: make new-verify M=$(M)"; \
	  exit 1; \
	fi; \
	scripts/check-spec "$$VERIFY"

check-card: ## Validate a task card: make check-card C=active/0001-slice-router-work
	@if [ -z "$(C)" ]; then \
	  echo "Error: C is required. Example: make check-card C=active/0001-slice-router-work"; \
	  exit 1; \
	fi
	@scripts/check-spec tasks/cards/$(C).md

new-milestone: ## Scaffold a new milestone spec: make new-milestone N=001 TITLE="My feature"
	@if [ -z "$(N)" ] || [ -z "$(TITLE)" ]; then \
	  echo "Error: N and TITLE are required."; \
	  echo "  Example: make new-milestone N=001 TITLE=\"Add ResourceHandler\""; \
	  exit 1; \
	fi
	@DEST=tasks/milestones/active/M-$(N).md; \
	if [ -f "$$DEST" ]; then \
	  echo "Error: $$DEST already exists."; exit 1; \
	fi; \
	rewrite() { expr="$$1"; file="$$2"; tmp=$$(mktemp); sed "$$expr" "$$file" > "$$tmp" && mv "$$tmp" "$$file"; }; \
	cp tasks/milestones/TEMPLATE.md "$$DEST"; \
	TITLE_ESC=$$(printf '%s\n' "$(TITLE)" | sed 's/[&/]/\\&/g'); \
	rewrite "s/M-XXX: <Title>/M-$(N): $$TITLE_ESC/" "$$DEST"; \
	rewrite "s|milestone/M-XXX-<slug>|milestone/M-$(N)-<slug>|" "$$DEST"; \
	echo "Created: $$DEST"; \
	echo "Next: fill in Goal, Architecture Decisions, Context, Tasks, then:"; \
	echo "  make check-spec M=active/M-$(N)"

new-plan: ## Scaffold a milestone plan: make new-plan M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make new-plan M=active/M-001"; \
	  exit 1; \
	fi
	@SPEC=tasks/milestones/$(M).md; \
	if [ ! -f "$$SPEC" ]; then \
	  echo "Error: $$SPEC not found."; \
	  exit 1; \
	fi; \
	ID=$$(sed -n 's/^# \(M-[0-9][0-9][0-9]\): .*$$/\1/p' "$$SPEC" | head -n1); \
	TITLE=$$(sed -n 's/^# M-[0-9][0-9][0-9]: \(.*\)$$/\1/p' "$$SPEC" | head -n1); \
	if [ -z "$$ID" ] || [ -z "$$TITLE" ]; then \
	  echo "Error: could not derive milestone ID/title from $$SPEC"; \
	  exit 1; \
	fi; \
	DEST=tasks/milestones/$$ID.plan.md; \
	if [ -f "$$DEST" ]; then \
	  echo "Error: $$DEST already exists."; \
	  exit 1; \
	fi; \
	rewrite() { expr="$$1"; file="$$2"; tmp=$$(mktemp); sed "$$expr" "$$file" > "$$tmp" && mv "$$tmp" "$$file"; }; \
	cp tasks/milestones/PLAN_TEMPLATE.md "$$DEST"; \
	TITLE_ESC=$$(printf '%s\n' "$$TITLE" | sed 's/[&/]/\\&/g'); \
	rewrite "s/^# Plan for M-XXX: <Title>/# Plan for $$ID: $$TITLE_ESC/" "$$DEST"; \
	rewrite "s/M-XXX/$$ID/g" "$$DEST"; \
	rewrite "s/<Title>/$$TITLE_ESC/g" "$$DEST"; \
	echo "Created: $$DEST"; \
	echo "Next: fill the plan fields, then:"; \
	echo "  make check-plan M=$(M)"

new-verify: ## Scaffold a milestone verify report: make new-verify M=active/M-001
	@if [ -z "$(M)" ]; then \
	  echo "Error: M is required. Example: make new-verify M=active/M-001"; \
	  exit 1; \
	fi
	@SPEC=tasks/milestones/$(M).md; \
	if [ ! -f "$$SPEC" ]; then \
	  echo "Error: $$SPEC not found."; \
	  exit 1; \
	fi; \
	ID=$$(sed -n 's/^# \(M-[0-9][0-9][0-9]\): .*$$/\1/p' "$$SPEC" | head -n1); \
	TITLE=$$(sed -n 's/^# M-[0-9][0-9][0-9]: \(.*\)$$/\1/p' "$$SPEC" | head -n1); \
	if [ -z "$$ID" ] || [ -z "$$TITLE" ]; then \
	  echo "Error: could not derive milestone ID/title from $$SPEC"; \
	  exit 1; \
	fi; \
	DEST=tasks/milestones/$$ID.verify.md; \
	if [ -f "$$DEST" ]; then \
	  echo "Error: $$DEST already exists."; \
	  exit 1; \
	fi; \
	rewrite() { expr="$$1"; file="$$2"; tmp=$$(mktemp); sed "$$expr" "$$file" > "$$tmp" && mv "$$tmp" "$$file"; }; \
	cp tasks/milestones/VERIFY_TEMPLATE.md "$$DEST"; \
	TITLE_ESC=$$(printf '%s\n' "$$TITLE" | sed 's/[&/]/\\&/g'); \
	rewrite "s/^# Verify M-XXX: <Title>/# Verify $$ID: $$TITLE_ESC/" "$$DEST"; \
	rewrite "s/M-XXX/$$ID/g" "$$DEST"; \
	rewrite "s/<Title>/$$TITLE_ESC/g" "$$DEST"; \
	echo "Created: $$DEST"; \
	echo "Next: fill the verify report, then:"; \
	echo "  make check-verify M=$(M)"

new-card: ## Scaffold a task card: make new-card ID=0001 SLUG=slice-router-work M=M-001 R=fix-bug
	@if [ -z "$(ID)" ] || [ -z "$(SLUG)" ] || [ -z "$(M)" ]; then \
	  echo "Error: ID, SLUG, and M are required."; \
	  echo "  Example: make new-card ID=0001 SLUG=slice-router-work M=M-001 R=fix-bug"; \
	  exit 1; \
	fi
	@DEST=tasks/cards/active/$(ID)-$(SLUG).md; \
	if [ -f "$$DEST" ]; then \
	  echo "Error: $$DEST already exists."; \
	  exit 1; \
	fi; \
	rewrite() { expr="$$1"; file="$$2"; tmp=$$(mktemp); sed "$$expr" "$$file" > "$$tmp" && mv "$$tmp" "$$file"; }; \
	cp tasks/cards/TEMPLATE.md "$$DEST"; \
	rewrite "s/^# Card XXXX/# Card $(ID)/" "$$DEST"; \
	rewrite "s/^Milestone:/Milestone: $(M)/" "$$DEST"; \
	if [ -n "$(R)" ]; then \
	  RECIPE_PATH="specs/change-recipes/$(R).yaml"; \
	  if [ ! -f "$$RECIPE_PATH" ]; then \
	    echo "Error: $$RECIPE_PATH not found."; \
	    echo "Available recipes:"; \
	    ls specs/change-recipes/*.yaml 2>/dev/null | sed 's|specs/change-recipes/||; s|\.yaml$$||'; \
	    rm -f "$$DEST"; \
	    exit 1; \
	  fi; \
	  rewrite "s|^Recipe: .*|Recipe: $$RECIPE_PATH|" "$$DEST"; \
	fi; \
	echo "Created: $$DEST"; \
	echo "Next: fill the card fields, then:"; \
	echo "  make check-card C=active/$(ID)-$(SLUG)"; \
	if [ -z "$(R)" ]; then \
	  echo "Hint: set Recipe: to one of specs/change-recipes/*.yaml or scaffold with R=<recipe-name> next time."; \
	fi

# ── Quality Gates (run locally, mirrors CI) ───────────────────────────────────

gates: ## Run all required quality gates (mirrors CI)
	go run ./internal/checks/dependency-rules
	go run ./internal/checks/agent-workflow
	go run ./internal/checks/module-manifests
	go run ./internal/checks/reference-layout
	go run ./internal/checks/extension-maturity
	go run ./internal/checks/extension-beta-evidence
	go run ./internal/checks/deprecation-inventory
	go vet ./...
	@UNFORMATTED=$$(gofmt -l .); \
	if [ -n "$$UNFORMATTED" ]; then \
	  echo "Unformatted files:"; echo "$$UNFORMATTED"; exit 1; \
	fi
	go test -race -timeout 60s ./...
	go test -timeout 20s ./...
	go test -coverprofile=/tmp/plumego-stable.cover ./core ./router ./middleware/... ./contract ./security/... ./store/... >/tmp/plumego-stable-cover.log
	@TOTAL=$$(go tool cover -func=/tmp/plumego-stable.cover | awk '/^total:/ {gsub("%","",$$3); print $$3}'); \
	echo "Stable-module total coverage: $$TOTAL%"; \
	awk -v total="$$TOTAL" -v min="70.0" 'BEGIN { if (total+0 < min+0) exit 1 }' || { \
	  echo "Coverage gate failed: expected >= 70.0%, got $$TOTAL%"; \
	  tail -n 50 /tmp/plumego-stable-cover.log || true; \
	  exit 1; \
	}
	cd cmd/plumego && go vet ./...
	cd cmd/plumego && go test -race -timeout 60s ./...
	cd cmd/plumego && go test -timeout 20s ./...
	@TMPDIR=$$(mktemp -d); \
	git diff -- website/src/generated > "$$TMPDIR/before"; \
	cd website && pnpm sync; \
	cd ..; \
	git diff -- website/src/generated > "$$TMPDIR/after"; \
	if ! cmp -s "$$TMPDIR/before" "$$TMPDIR/after"; then \
	  echo "website generated files are stale. Run: cd website && pnpm sync"; \
	  exit 1; \
	fi
	cd website && pnpm check
	cd website && pnpm build
	@echo "All gates passed."

fmt: ## Format all Go source files in-place
	gofmt -w .

vet: ## Run go vet on all packages
	go vet ./...

test: ## Run tests (standard timeout)
	go test -timeout 20s ./...

test-race: ## Run tests with race detector
	go test -race -timeout 60s ./...

# ── Git Hooks ─────────────────────────────────────────────────────────────────

setup-hooks: ## Install local git hooks (pre-push quality gates)
	@cp scripts/pre-push .git/hooks/pre-push
	@chmod +x .git/hooks/pre-push
	@echo "Installed: .git/hooks/pre-push"
	@echo "Quality gates now run automatically before every git push."
	@echo "  milestone/* branches: full gate suite"
	@echo "  other branches:       quick check (vet + fmt + tests)"
	@echo "Skip once with: git push --no-verify"
