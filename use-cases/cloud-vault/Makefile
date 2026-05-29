.PHONY: dev web-dev web-build server-build build run clean tidy

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
	rm -rf bin web/dist internal/web/static/*
	touch internal/web/static/index.html
