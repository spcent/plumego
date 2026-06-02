.PHONY: dev build test lint clean package frontend-build frontend-dev docker-build release-smoke help

# Default target
help:
	@echo "Developer Data Workbench - Build & Development"
	@echo ""
	@echo "Usage:"
	@echo "  make dev         - Start development servers (Go + Vite with HMR)"
	@echo "  make build       - Build production binary with embedded frontend"
	@echo "  make test        - Run all tests (Go + frontend)"
	@echo "  make lint        - Run linters (Go + frontend)"
	@echo "  make package     - Create distributable package"
	@echo "  make docker-build - Build release Docker image"
	@echo "  make release-smoke - Run multi-datasource release smoke test"
	@echo "  make clean       - Remove build artifacts"
	@echo ""
	@echo "Development:"
	@echo "  make frontend-dev    - Start Vite dev server only (port 5173)"
	@echo "  make frontend-build  - Build frontend to web/dist"
	@echo ""
	@echo "Production:"
	@echo "  Binary output: bin/dbadmin"
	@echo "  Run: ./bin/dbadmin"
	@echo "  Default: http://127.0.0.1:8080"

# Development mode - runs Go backend and Vite frontend concurrently
dev:
	@echo "Starting development servers..."
	@echo "  Backend:  http://127.0.0.1:8080"
	@echo "  Frontend: http://localhost:5173 (with HMR)"
	@echo ""
	@trap 'kill 0' EXIT; \
	(go run . -addr=127.0.0.1:8080 & \
	cd web && npm run dev)

# Build frontend for production
frontend-build:
	@echo "Building frontend..."
	cd web && npm install && npm run build

# Build production binary with embedded frontend
build: frontend-build
	@echo "Building production binary..."
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)" -o bin/dbadmin .
	@echo "✓ Binary built: bin/dbadmin"
	@echo "  Run: ./bin/dbadmin"
	@echo "  Default: http://127.0.0.1:8080"

# Run all tests
test:
	@echo "Running Go tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Running frontend tests..."
	cd web && npm test
	@echo ""
	@echo "✓ All tests passed"

# Run linters
lint:
	@echo "Running Go linter..."
	go vet ./...
	@echo ""
	@echo "Running frontend linter..."
	cd web && npm run lint
	@echo ""
	@echo "✓ Linting complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/.vite/
	rm -f coverage.out
	@echo "✓ Cleaned"

# Create distributable package
package: build
	@echo "Creating distribution package..."
	@mkdir -p dist
	@cp bin/dbadmin dist/
	@cp README.md dist/
	@cp env.example dist/.env.example
	@cp -r docker-compose.dev.yml dist/ 2>/dev/null || true
	@cp -r docker dist/ 2>/dev/null || true
	@cd dist && tar -czf ../dbadmin-$(shell date +%Y%m%d)-$(shell uname -s)-$(shell uname -m).tar.gz *
	@rm -rf dist
	@echo "✓ Package created: dbadmin-*.tar.gz"

# Build release Docker image. Run from the repository root context because
# dbadmin's go.mod uses replace github.com/spcent/plumego => ../..
docker-build:
	cd ../.. && docker build -f use-cases/dbadmin/Dockerfile -t dbadmin:local .

# Run release smoke test against Docker Compose datasources.
release-smoke:
	bash scripts/release-smoke.sh

# Frontend development server only
frontend-dev:
	cd web && npm run dev
