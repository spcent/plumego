# Development Guide

This guide covers setting up a development environment for Markdown Cloud Vault.

## Prerequisites

- **Go 1.22+**: Backend development
- **Node.js 18+**: Frontend development
- **npm 9+** or **pnpm 8+**: Package management
- **Git**: Version control
- **Make**: Build automation (optional)

## Getting Started

### Clone Repository

```bash
git clone https://github.com/example/cloud-vault.git
cd cloud-vault
```

### Install Dependencies

**Backend**:
```bash
# Install Go dependencies
go mod download

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/air-verse/air@latest  # Hot reload
```

**Frontend**:
```bash
cd web
npm install
# or
pnpm install
```

### Development Environment

**Environment Variables**:
Create `.env` file in project root:
```bash
# Server
SERVER_ADDR=":8080"
SERVER_HOST="localhost"

# Database
DATABASE_PATH="./data/dev.db"

# Storage
STORAGE_PROVIDER="local"
STORAGE_LOCAL_ROOT="./data/objects"

# Authentication
AUTH_ENABLED=false

# AI (optional)
AI_ENABLED=false

# Logging
LOGGING_LEVEL="debug"
LOGGING_FORMAT="text"
```

### Run Development Server

**Backend**:
```bash
# Run server
go run cmd/server/main.go

# Or with hot reload
air
```

**Frontend**:
```bash
cd web
npm run dev
# or
pnpm dev
```

The frontend dev server runs on `http://localhost:5173` and proxies API requests to `http://localhost:8080`.

## Project Structure

```
cloud-vault/
├── cmd/
│   ├── server/          # Server entry point
│   └── desktop/         # Desktop app entry point
├── internal/
│   ├── ai/              # AI features
│   ├── app/             # Application logic
│   ├── auth/            # Authentication
│   ├── backup/          # Backup functionality
│   ├── collection/      # Collections
│   ├── config/          # Configuration
│   ├── database/        # Database layer
│   ├── diagnostics/     # Diagnostic tools
│   ├── document/        # Document management
│   ├── importer/        # Import functionality
│   ├── logging/         # Logging utilities
│   ├── organize/        # Organization features
│   ├── search/          # Search functionality
│   ├── storage/         # Storage backends
│   ├── system/          # System information
│   ├── tag/             # Tag management
│   ├── update/          # Update checking
│   └── web/             # Web server
├── web/
│   ├── src/             # Frontend source
│   ├── public/          # Static assets
│   └── package.json     # Frontend dependencies
├── docs/                # Documentation
├── scripts/             # Build and utility scripts
├── config.example.toml  # Example configuration
├── Makefile             # Build automation
└── README.md            # Project overview
```

## Development Workflow

### Code Style

**Go**:
```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run tests
go test ./...
```

**TypeScript/React**:
```bash
cd web
npm run lint
npm run format
```

### Database Migrations

Create a new migration:
```bash
./scripts/create-migration.sh add_new_table
```

Edit migration file in `internal/database/migrations/`:
```go
func migrate_add_new_table(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS new_table (
            id INTEGER PRIMARY KEY,
            name TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
    return err
}
```

### API Development

**Add new endpoint**:

1. Define handler in appropriate package:
```go
// internal/document/handler.go
func (h *Handler) GetDocument(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    
    doc, err := h.service.GetDocument(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(doc)
}
```

2. Register route:
```go
// internal/web/routes.go
r.Get("/api/v1/documents/{id}", h.document.GetDocument)
```

3. Add tests:
```go
// internal/document/handler_test.go
func TestGetDocument(t *testing.T) {
    // Test implementation
}
```

### Frontend Development

**Add new page**:

1. Create page component:
```tsx
// web/src/pages/NewPage.tsx
export default function NewPage() {
  return (
    <div className="container">
      <h1>New Page</h1>
    </div>
  )
}
```

2. Add route:
```tsx
// web/src/App.tsx
import NewPage from './pages/NewPage'

// In routes:
<Route path="/new" element={<NewPage />} />
```

3. Add navigation:
```tsx
// web/src/components/Navigation.tsx
<Link to="/new">New Page</Link>
```

## Testing

### Run Tests

**Backend**:
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/document/...

# With verbose output
go test -v ./internal/document/...
```

**Frontend**:
```bash
cd web
npm run test
# or
pnpm test
```

### Write Tests

**Unit test example**:
```go
// internal/document/service_test.go
func TestCreateDocument(t *testing.T) {
    service := NewService(mockRepo)
    
    doc, err := service.CreateDocument(context.Background(), CreateRequest{
        Title:   "Test Document",
        Content: "# Test\n\nContent here",
    })
    
    assert.NoError(t, err)
    assert.NotNil(t, doc)
    assert.Equal(t, "Test Document", doc.Title)
}
```

**Integration test example**:
```go
// internal/document/integration_test.go
func TestDocumentAPI(t *testing.T) {
    app := setupTestApp(t)
    defer app.cleanup()
    
    // Create document
    resp := app.postJSON("/api/v1/documents", map[string]interface{}{
        "title":   "Test",
        "content": "# Test",
    })
    
    assert.Equal(t, http.StatusCreated, resp.Code)
    
    var doc Document
    json.Unmarshal(resp.Body.Bytes(), &doc)
    
    // Get document
    resp = app.get("/api/v1/documents/" + doc.ID)
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

## Building

### Build Server

```bash
# Development build
go build -o bin/cloud-vault cmd/server/main.go

# Production build
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/cloud-vault cmd/server/main.go
```

### Build Frontend

```bash
cd web
npm run build
# or
pnpm build
```

Output in `web/dist/` is embedded into the server binary.

### Build Desktop App

Requires Wails CLI:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Build:
```bash
cd cmd/desktop
wails build
```

### Build All

```bash
make build
```

## Debugging

### Backend Debugging

**Using delve**:
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start server with debugger
dlv debug cmd/server/main.go

# Connect IDE debugger to port 2345
```

**Using logs**:
```go
logger.Debug("Processing document", 
    zap.String("doc_id", docID),
    zap.Int("size", len(content)),
)
```

### Frontend Debugging

**React DevTools**:
- Install browser extension
- Inspect component tree
- View props and state

**Console logging**:
```typescript
console.log('Document:', document)
console.error('Error:', error)
```

**Breakpoints**:
- Use browser DevTools
- Add `debugger` statement in code

## Common Tasks

### Add New Configuration Option

1. Add to config struct:
```go
// internal/config/config.go
type Config struct {
    // ...
    NewOption string `toml:"new_option"`
}
```

2. Add to example config:
```toml
# config.example.toml
new_option = "default_value"
```

3. Use in code:
```go
value := cfg.NewOption
```

### Add Database Table

1. Create migration:
```bash
./scripts/create-migration.sh create_new_table
```

2. Write migration:
```go
func migrate_create_new_table(db *sql.DB) error {
    _, err := db.Exec(`
        CREATE TABLE new_table (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
    `)
    return err
}
```

3. Add repository methods:
```go
// internal/newfeature/repository.go
type Repository struct {
    db *sql.DB
}

func (r *Repository) Create(name string) error {
    _, err := r.db.Exec(
        "INSERT INTO new_table (name) VALUES (?)",
        name,
    )
    return err
}
```

### Add New API Endpoint

1. Create handler:
```go
// internal/newfeature/handler.go
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
    items, err := h.service.List(r.Context())
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(items)
}
```

2. Register route:
```go
// internal/web/routes.go
r.Get("/api/v1/newfeature", h.newfeature.List)
```

3. Add documentation:
```markdown
# API Documentation

## GET /api/v1/newfeature

List all items.

**Response**:
```json
[
  {
    "id": "abc123",
    "name": "Item 1"
  }
]
```
```

## Performance Profiling

### CPU Profiling

```bash
# Start server with profiling
go run cmd/server/main.go --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling

```bash
# Start server with profiling
go run cmd/server/main.go --memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

### Frontend Performance

```bash
cd web
npm run build -- --analyze
```

## Contributing

### Code Review Process

1. Create feature branch:
```bash
git checkout -b feature/new-feature
```

2. Make changes and commit:
```bash
git add .
git commit -m "feat: add new feature"
```

3. Push and create PR:
```bash
git push origin feature/new-feature
```

4. Address review comments

5. Merge after approval

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Example**:
```
feat(search): add advanced search operators

Add support for phrase search, wildcards, and boolean operators.
Includes comprehensive tests and documentation updates.

Closes #123
```

## Troubleshooting

### Build Errors

**Error**: `go.mod file not found`
```bash
go mod init cloud-vault
go mod tidy
```

**Error**: `node_modules not found`
```bash
cd web
npm install
```

### Runtime Errors

**Error**: `database locked`
- Stop other instances
- Check for zombie processes
- Restart application

**Error**: `port already in use`
```bash
# Find process
lsof -i :8080

# Kill process
kill -9 <PID>

# Or use different port
./cloud-vault --addr :8081
```

### Test Failures

**Error**: `test timeout`
```bash
# Increase timeout
go test -timeout 30s ./...
```

**Error**: `database migration failed`
```bash
# Reset test database
rm -f data/test.db

# Run migrations
go run cmd/migrate/main.go
```

## Resources

### Documentation

- [Go Documentation](https://golang.org/doc/)
- [React Documentation](https://react.dev/)
- [Vite Documentation](https://vitejs.dev/)
- [Tailwind CSS](https://tailwindcss.com/)

### Tools

- [GoLand](https://www.jetbrains.com/go/) or [VS Code](https://code.visualstudio.com/) with Go extension
- [Postman](https://www.postman.com/) for API testing
- [DB Browser for SQLite](https://sqlitebrowser.org/) for database inspection

### Community

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and ideas
- Discord: Real-time chat (if available)

## Next Steps

- Read the [Architecture Overview](./architecture.md) (future)
- Review [API Documentation](./api.md) (future)
- Check [Contributing Guidelines](./contributing.md) (future)

---

**Last updated**: 2026-05-30
