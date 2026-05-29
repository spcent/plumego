.PHONY: dev web-dev web-build server-build build run clean tidy test test-go test-web test-e2e doctor bench

# Run the frontend Vite dev server (proxies API to :8080)
dev:
	cd web && pnpm dev

# Run the frontend Vite dev server explicitly
web-dev:
	cd web && pnpm dev

# Build the frontend into internal/web/static/
web-build:
	cd web && pnpm install && pnpm build

# Build the Go server binary
server-build:
	go build -o bin/markdown-vault ./cmd/server

# Full build: frontend then server
build: web-build server-build

# Run the server with go run (no binary, uses local storage by default)
run:
	go run ./cmd/server

# Tidy Go module dependencies
tidy:
	go mod tidy

# Remove build artifacts
clean:
	rm -rf bin web/dist internal/web/static/* e2e/test-results e2e/playwright-report
	touch internal/web/static/index.html

# Run all tests
test: test-go

# Run Go tests
test-go:
	go test ./...

# Run frontend type-check and build
test-web:
	cd web && pnpm build

# Run E2E tests (requires: cd e2e && pnpm install && pnpm exec playwright install)
test-e2e:
	cd e2e && pnpm exec playwright test

# Run doctor check against running server
doctor:
	curl -s -X POST http://localhost:8080/api/v1/system/doctor \
		-H 'Content-Type: application/json' \
		-d '{"checks":[]}' | jq .

# Run benchmark
bench:
	go run ./cmd/bench --docs 1000 --queries 20
