.PHONY: build test run lint clean frontend-build frontend-sync e2e

BIN := bin/guardus

build: frontend-sync
	@mkdir -p bin
	go build -o $(BIN) .

run: build
	./$(BIN)

test:
	go test ./... -race -timeout 60s

e2e:
	go test ./internal/app/... -timeout 120s -run TestAppRouteWiring

lint:
	go vet ./...

clean:
	rm -rf bin

# Build the React + Vite SPA in web/app/ into web/static/.
# Requires Node.js; output is embedded by the Go binary via //go:embed.
frontend-build:
	cd web/app && npm ci --no-audit --no-fund && npm run build

# Legacy alias kept for back-compat — the React SPA is now owned by this
# module (see web/app/), so "sync" just means "build".
frontend-sync: frontend-build
