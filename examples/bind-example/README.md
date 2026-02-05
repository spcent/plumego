# Bind Example

This example demonstrates the standardized JSON binding + validation middleware.

## Run

```bash
go run ./examples/bind-example
```

The service listens on `:8082`.

## Success

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"supersecret"}'
```

Expected response:

```json
{"email":"user@example.com","ok":true}
```

## Validation error (field-level)

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"not-an-email","password":""}'
```

Expected response (example):

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "validation failed",
    "category": "validation_error",
    "details": {
      "fields": [
        {"field":"email","code":"email","message":"invalid email format"},
        {"field":"password","code":"required","message":"is required"}
      ]
    }
  }
}
```

## JSON parse error

```bash
curl -X POST http://127.0.0.1:8082/v1/users \
  -H 'Content-Type: application/json' \
  -d '{'
```

Expected response (example):

```json
{
  "error": {
    "code": "INVALID_JSON",
    "message": "invalid JSON payload",
    "category": "validation_error"
  }
}
```
