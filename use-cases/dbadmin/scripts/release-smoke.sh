#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_ADDR="${APP_ADDR:-127.0.0.1:18080}"
BASE_URL="http://${APP_ADDR}"
DATA_DIR="${DATA_DIR:-$(mktemp -d)}"
COOKIE_JAR="${COOKIE_JAR:-$(mktemp)}"
PASSWORD="${DBADMIN_PASSWORD:-release-smoke-password}"
KEY="${DBADMIN_ENCRYPTION_KEY:-1111111111111111111111111111111111111111111111111111111111111111}"

cleanup() {
  if [[ -n "${APP_PID:-}" ]]; then
    kill "${APP_PID}" >/dev/null 2>&1 || true
    wait "${APP_PID}" >/dev/null 2>&1 || true
  fi
  rm -f "${COOKIE_JAR}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "missing required command: $1" >&2
    exit 1
  }
}

json_field() {
  python3 - "$1" <<'PY'
import json, sys
path = sys.argv[1].split(".")
data = json.load(sys.stdin)
for item in path:
    data = data[item]
print(data)
PY
}

wait_url() {
  local url="$1"
  local name="$2"
  for _ in $(seq 1 60); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  echo "timed out waiting for ${name}: ${url}" >&2
  exit 1
}

api() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" \
      -H "Content-Type: application/json" \
      -X "${method}" \
      --data "${body}" \
      "${BASE_URL}${path}"
  else
    curl -fsS -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" \
      -X "${method}" \
      "${BASE_URL}${path}"
  fi
}

create_connection() {
  local body="$1"
  api POST /api/connections "${body}" | json_field data.id
}

test_connection() {
  local id="$1"
  local name="$2"
  local ok
  ok="$(api POST "/api/connections/${id}/test" | json_field data.ok)"
  if [[ "${ok}" != "True" && "${ok}" != "true" ]]; then
    echo "connection test failed: ${name}" >&2
    exit 1
  fi
  echo "ok: ${name}"
}

require_cmd curl
require_cmd docker
require_cmd python3

cd "${ROOT_DIR}"
echo "starting datasource stack"
docker compose -f docker-compose.dev.yml up -d --wait

wait_url "http://127.0.0.1:19200/_cluster/health" "elasticsearch"

echo "building frontend"
(cd web && npm ci && npm run build)

echo "starting dbadmin on ${APP_ADDR}"
APP_ADDR="${APP_ADDR}" \
DBADMIN_DATA_DIR="${DATA_DIR}" \
DBADMIN_USER=admin \
DBADMIN_PASSWORD="${PASSWORD}" \
DBADMIN_ENCRYPTION_KEY="${KEY}" \
go run . &
APP_PID="$!"

wait_url "${BASE_URL}/healthz" "dbadmin"

echo "logging in"
curl -fsS -c "${COOKIE_JAR}" \
  -H "Content-Type: application/json" \
  -X POST \
  --data "{\"username\":\"admin\",\"password\":\"${PASSWORD}\"}" \
  "${BASE_URL}/api/auth/login" >/dev/null

mysql_id="$(create_connection '{"name":"smoke-mysql","driver":"mysql","host":"127.0.0.1","port":13306,"database":"testdb","username":"dbadmin","password":"dbadminpass","save_password":true}')"
redis_id="$(create_connection '{"name":"smoke-redis","driver":"redis","host":"127.0.0.1","port":16379,"password":"redispass","save_password":true}')"
mongo_id="$(create_connection '{"name":"smoke-mongo","driver":"mongodb","mongo_uri":"mongodb://dbadmin:dbadminpass@127.0.0.1:37017/testdb?authSource=admin"}')"
es_id="$(create_connection '{"name":"smoke-es","driver":"elasticsearch","es_nodes":["http://127.0.0.1:19200"]}')"

test_connection "${mysql_id}" mysql
test_connection "${redis_id}" redis
test_connection "${mongo_id}" mongodb
test_connection "${es_id}" elasticsearch

echo "release smoke passed"
