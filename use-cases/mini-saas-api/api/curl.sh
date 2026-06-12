#!/usr/bin/env bash
# End-to-end curl walkthrough for mini-saas-api.
# Start the server first:  go run .   (default address :8090)
# Requires: curl, python3 (for JSON field extraction).
set -euo pipefail

BASE="${BASE:-http://localhost:8090}"

json() { python3 -c "import json,sys; print(json.load(sys.stdin)$1)"; }

echo "== health =="
curl -fsS "$BASE/healthz" | json "['data']['status']"
curl -fsS "$BASE/readyz" | json "['data']['ready']"

STAMP=$(date +%s)
EMAIL="demo-$STAMP@example.com"
SLUG="demo-$STAMP"

echo "== signup ($EMAIL / $SLUG) =="
SIGNUP=$(curl -fsS -X POST "$BASE/api/v1/auth/signup" \
  -H 'Content-Type: application/json' \
  -d "{\"email\":\"$EMAIL\",\"name\":\"Demo\",\"password\":\"Str0ng!Password#2026\",\"workspace_name\":\"Demo Space\",\"workspace_slug\":\"$SLUG\"}")
ACCESS=$(echo "$SIGNUP"  | json "['data']['tokens']['access_token']")
REFRESH=$(echo "$SIGNUP" | json "['data']['tokens']['refresh_token']")
echo "workspace: $(echo "$SIGNUP" | json "['data']['tenant']['slug']")"

echo "== me =="
curl -fsS "$BASE/api/v1/me" -H "Authorization: Bearer $ACCESS" | json "['data']['membership']['role']"

echo "== tenant + usage =="
curl -fsS "$BASE/api/v1/tenant" -H "Authorization: Bearer $ACCESS" | json "['data']['usage']"

echo "== create project (idempotent) =="
PROJECT=$(curl -fsS -X POST "$BASE/api/v1/projects" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -H "Idempotency-Key: demo-key-$STAMP" \
  -d '{"name":"Apollo","description":"first project"}')
PROJECT_ID=$(echo "$PROJECT" | json "['data']['id']")
echo "project: $PROJECT_ID"

echo "== replay the same create (same Idempotency-Key) =="
REPLAY=$(curl -fsS -X POST "$BASE/api/v1/projects" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -H "Idempotency-Key: demo-key-$STAMP" \
  -d '{"name":"Apollo","description":"first project"}')
echo "replayed id matches: $([ "$(echo "$REPLAY" | json "['data']['id']")" = "$PROJECT_ID" ] && echo yes || echo NO)"

echo "== list / update / delete project =="
curl -fsS "$BASE/api/v1/projects" -H "Authorization: Bearer $ACCESS" | json "['data']['total']"
curl -fsS -X PUT "$BASE/api/v1/projects/$PROJECT_ID" \
  -H "Authorization: Bearer $ACCESS" -H 'Content-Type: application/json' \
  -d '{"status":"paused"}' | json "['data']['status']"
curl -fsS -o /dev/null -w "delete: %{http_code}\n" -X DELETE "$BASE/api/v1/projects/$PROJECT_ID" \
  -H "Authorization: Bearer $ACCESS"

echo "== refresh token rotation =="
ROTATED=$(curl -fsS -X POST "$BASE/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}")
echo "new access token issued: $([ -n "$(echo "$ROTATED" | json "['data']['tokens']['access_token']")" ] && echo yes)"
echo "reusing the old refresh token must fail:"
curl -sS -o /dev/null -w "  reuse status: %{http_code} (expect 401)\n" -X POST "$BASE/api/v1/auth/refresh" \
  -H 'Content-Type: application/json' -d "{\"refresh_token\":\"$REFRESH\"}"

echo "== audit trail (admin) =="
curl -fsS "$BASE/api/v1/tenant/audit?limit=5" -H "Authorization: Bearer $ACCESS" | json "['data']['total']"

echo "== metrics =="
curl -fsS "$BASE/metrics" | head -5

echo "ALL OK"
