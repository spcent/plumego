# with-ai Scenario Reference

`reference/with-ai` is a non-canonical scenario reference.

It shows the safe first path for AI service work using only the stable-tier
`x/ai` subpackages: `provider`, `session`, `streaming`, and `tool`.

The root `x/ai` family remains experimental. This scenario reference uses mock/offline
behavior and does not call live providers.

## What It Demonstrates

- mock-backed provider completion
- in-memory AI session state
- allow-listed tool execution
- explicit stream manager ownership
- normal Plumego route registration and `contract` responses

## Routes

- `POST /api/chat`
- `GET /api/ai/status`

## Run

```bash
go run ./reference/with-ai
```
