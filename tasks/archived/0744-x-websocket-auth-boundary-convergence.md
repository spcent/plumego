# 0744 - x/websocket auth boundary convergence

Status: done
Priority: P1
State: done
Primary module: `x/websocket`

## Problem

Room password checks and JWT verification are coupled through one
`RoomAuthenticator` interface. The server accepts query tokens by default, and
the helper names do not make anonymous, token-authenticated, and room-authorized
flows explicit.

## Scope

- Split room authorization from token authentication in `ServerConfig`.
- Require explicit `AllowUnauthenticated` and `AllowQueryToken` choices.
- Replace `VerifyJWT` as the server-facing method name with a narrower token
  authentication contract.
- Update in-repository callers and tests in the same change.
- Re-run symbol searches before and after migration.

## Out of Scope

- Full OIDC/JWT policy implementation.
- Origin policy redesign.
- Release evidence changes.

## Validation

- `rg -n --glob '*.go' 'RoomAuthenticator|VerifyJWT|CheckRoomPassword|ServeWSWithAuth' .`
- `go test -timeout 20s ./x/websocket/...`
- `go vet ./x/websocket/...`
- `go build ./...`

## Outcome

- Replaced the combined `RoomAuthenticator` server contract with separate
  `RoomAuthorizer` and `TokenAuthenticator` interfaces.
- Split simple room-password auth from `HS256TokenAuth`.
- Removed the old `ServeWSWithAuth` helper from Go callers and public manifest
  entries.
- Added explicit `AllowUnauthenticated` and `AllowQueryToken` server options;
  query tokens are rejected unless opted in.
- Updated docs, website examples, devserver wiring, and tests.
