# JSON API Example

Simple REST-like JSON API demonstrating:

- POST and GET routes
- Request/response JSON encoding and decoding
- Path parameter extraction
- Error handling with `contract` error builder
- Thread-safe in-memory storage
- Multiple HTTP status codes (201 Created, 404 Not Found, etc.)

## Run

```bash
go build -o json-api .
./json-api
```

Then in another terminal:

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"alice","email":"alice@example.com"}'
# {"id":1,"name":"alice","email":"alice@example.com"}

# List users
curl http://localhost:8080/users
# [{"id":1,"name":"alice","email":"alice@example.com"}]

# Get one user
curl http://localhost:8080/users/1
# {"id":1,"name":"alice","email":"alice@example.com"}

# Get non-existent user (404)
curl http://localhost:8080/users/999
# {"type":"not_found","message":"User not found"}
```

## Key APIs

- **`app.Post(path, handler)`** — Register a POST route
- **`app.Get(path, handler)`** — Register a GET route
- **`json.NewDecoder(r.Body).Decode(&v)`** — Decode JSON request
- **`json.NewEncoder(w).Encode(v)`** — Encode JSON response (via `contract.WriteResponse`)
- **`contract.NewErrorBuilder().Type(...).Message(...).Build()`** — Create structured errors
- **`r.PathValue("name")`** — Extract path parameter (Go 1.22+)

## Limitations

This is a teaching example. For production:

1. Use proper validation and error reporting.
2. Add middleware for logging and request/response middleware.
3. Implement persistence (database) instead of in-memory storage.
4. Add authentication and authorization.
5. See [`reference/standard-service`](../../reference/standard-service) for a production-style layout.

## Next Steps

- Add middleware with `app.Use()`
- Add route groups with `app.Group()`
- See [`reference/with-rest`](../../reference/with-rest) for CRUD conventions
