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
    if isinstance(data, list):
        data = data[int(item)]
    else:
        data = data[item]
print(data)
PY
}

json_quote() {
  python3 -c 'import json, sys; print(json.dumps(sys.stdin.read()))'
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

http_status() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local origin="${4:-}"
  local args=(-sS -o /tmp/dbadmin-smoke-response -w "%{http_code}" -b "${COOKIE_JAR}" -c "${COOKIE_JAR}" -X "${method}")
  if [[ -n "${origin}" ]]; then
    args+=(-H "Origin: ${origin}")
  fi
  if [[ -n "${body}" ]]; then
    args+=(-H "Content-Type: application/json" --data "${body}")
  fi
  curl "${args[@]}" "${BASE_URL}${path}"
}

expect_status() {
  local expected="$1"
  local method="$2"
  local path="$3"
  local body="${4:-}"
  local origin="${5:-}"
  local status
  status="$(http_status "${method}" "${path}" "${body}" "${origin}")"
  if [[ "${status}" != "${expected}" ]]; then
    echo "unexpected status for ${method} ${path}: got ${status}, want ${expected}" >&2
    cat /tmp/dbadmin-smoke-response >&2 || true
    exit 1
  fi
  echo "ok: ${method} ${path} -> ${status}"
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

sql_exec() {
  local id="$1"
  local db="$2"
  local sql="$3"
  local confirm="${4:-false}"
  local sql_json
  sql_json="$(printf '%s' "${sql}" | json_quote)"
  api POST "/api/conn/${id}/db/${db}/query" "{\"sql\":${sql_json},\"database\":\"${db}\",\"confirmDangerous\":${confirm}}"
}

redis_command() {
  local id="$1"
  local command="$2"
  local command_json
  command_json="$(printf '%s' "${command}" | json_quote)"
  api POST "/api/conn/${id}/redis/0/command" "{\"command\":${command_json}}"
}

mongo_insert() {
  local id="$1"
  local document="$2"
  local doc_json
  doc_json="$(printf '%s' "${document}" | json_quote)"
  api POST "/api/connections/${id}/mongo/documents" "{\"database\":\"testdb\",\"collection\":\"smoke_release\",\"document\":${doc_json}}"
}

mongo_query() {
  local id="$1"
  local filter="$2"
  local filter_json
  filter_json="$(printf '%s' "${filter}" | json_quote)"
  api POST "/api/connections/${id}/mongo/documents/query" "{\"database\":\"testdb\",\"collection\":\"smoke_release\",\"filter\":${filter_json},\"limit\":5}"
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

echo "checking unauthenticated and cross-origin guards"
expect_status 401 GET /api/auth/me
expect_status 403 POST /api/auth/login "{\"username\":\"admin\",\"password\":\"bad\"}" "http://evil.local"
expect_status 401 POST /api/auth/login "{\"username\":\"admin\",\"password\":\"bad\"}"

echo "logging in"
curl -fsS -c "${COOKIE_JAR}" \
  -H "Content-Type: application/json" \
  -X POST \
  --data "{\"username\":\"admin\",\"password\":\"${PASSWORD}\"}" \
  "${BASE_URL}/api/auth/login" >/dev/null

expect_status 403 POST /api/connections '{"name":"blocked"}' "http://evil.local"

mysql_id="$(create_connection '{"name":"smoke-mysql","driver":"mysql","host":"127.0.0.1","port":13306,"database":"testdb","username":"dbadmin","password":"dbadminpass","save_password":true}')"
mysql_readonly_id="$(create_connection '{"name":"smoke-mysql-readonly","driver":"mysql","host":"127.0.0.1","port":13306,"database":"testdb","username":"dbadmin","password":"dbadminpass","save_password":true,"readonly":true}')"
redis_id="$(create_connection '{"name":"smoke-redis","driver":"redis","host":"127.0.0.1","port":16379,"password":"redispass","save_password":true}')"
redis_readonly_id="$(create_connection '{"name":"smoke-redis-readonly","driver":"redis","host":"127.0.0.1","port":16379,"password":"redispass","save_password":true,"readonly":true}')"
mongo_id="$(create_connection '{"name":"smoke-mongo","driver":"mongodb","mongo_uri":"mongodb://dbadmin:dbadminpass@127.0.0.1:37017/testdb?authSource=admin"}')"
mongo_readonly_id="$(create_connection '{"name":"smoke-mongo-readonly","driver":"mongodb","mongo_uri":"mongodb://dbadmin:dbadminpass@127.0.0.1:37017/testdb?authSource=admin","readonly":true}')"
es_id="$(create_connection '{"name":"smoke-es","driver":"elasticsearch","es_nodes":["http://127.0.0.1:19200"]}')"
es_readonly_id="$(create_connection '{"name":"smoke-es-readonly","driver":"elasticsearch","es_nodes":["http://127.0.0.1:19200"],"readonly":true}')"

test_connection "${mysql_id}" mysql
test_connection "${mysql_readonly_id}" mysql-readonly
test_connection "${redis_id}" redis
test_connection "${redis_readonly_id}" redis-readonly
test_connection "${mongo_id}" mongodb
test_connection "${mongo_readonly_id}" mongodb-readonly
test_connection "${es_id}" elasticsearch
test_connection "${es_readonly_id}" elasticsearch-readonly

echo "checking credential redaction"
mongo_uri="$(api GET "/api/connections/${mongo_id}" | json_field data.mongo_uri)"
if [[ "${mongo_uri}" == *dbadminpass* ]]; then
  echo "mongo uri leaked password: ${mongo_uri}" >&2
  exit 1
fi
mysql_password="$(api GET "/api/connections/${mysql_id}" | json_field data.password 2>/dev/null || true)"
if [[ -n "${mysql_password}" ]]; then
  echo "mysql password was returned by API" >&2
  exit 1
fi

echo "checking MySQL query and readonly guard"
sql_exec "${mysql_id}" testdb "CREATE TABLE IF NOT EXISTS smoke_release (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(64))" >/dev/null
sql_exec "${mysql_id}" testdb "INSERT INTO smoke_release (name) VALUES ('mysql-smoke')" >/dev/null
mysql_count="$(sql_exec "${mysql_id}" testdb "SELECT COUNT(*) AS count FROM smoke_release WHERE name = 'mysql-smoke'" | json_field data.rows.0.count)"
if [[ "${mysql_count}" -lt 1 ]]; then
  echo "mysql smoke query returned count ${mysql_count}" >&2
  exit 1
fi
expect_status 403 POST "/api/conn/${mysql_readonly_id}/db/testdb/query" '{"sql":"INSERT INTO smoke_release (name) VALUES (\"blocked\")","database":"testdb"}'

echo "checking Redis command and readonly/forbidden guards"
redis_command "${redis_id}" "SET smoke:release redis-smoke" >/dev/null
redis_value="$(redis_command "${redis_id}" "GET smoke:release" | json_field data.result)"
if [[ "${redis_value}" != "redis-smoke" ]]; then
  echo "redis GET returned ${redis_value}" >&2
  exit 1
fi
expect_status 403 POST "/api/conn/${redis_readonly_id}/redis/0/command" '{"command":"SET smoke:blocked blocked"}'
expect_status 403 POST "/api/conn/${redis_id}/redis/0/command" '{"command":"KEYS *"}'

echo "checking MongoDB document operations and readonly guard"
mongo_inserted_id="$(mongo_insert "${mongo_id}" '{"name":"mongo-smoke"}' | json_field data.inserted_id)"
if [[ -z "${mongo_inserted_id}" ]]; then
  echo "mongo insert did not return inserted_id" >&2
  exit 1
fi
mongo_total="$(mongo_query "${mongo_id}" '{"name":"mongo-smoke"}' | json_field data.total)"
if [[ "${mongo_total}" -lt 1 ]]; then
  echo "mongo query returned total ${mongo_total}" >&2
  exit 1
fi
expect_status 403 POST "/api/connections/${mongo_readonly_id}/mongo/documents" '{"database":"testdb","collection":"smoke_release","document":"{\"name\":\"blocked\"}"}'

echo "checking Elasticsearch import/search and readonly guard"
api POST "/api/connections/${es_id}/es/import" '{"index":"smoke-release","documents":[{"_id":"1","name":"es-smoke"}],"confirm":true}' >/dev/null
es_hits=0
for _ in $(seq 1 10); do
  es_hits="$(api POST "/api/connections/${es_id}/es/search" '{"index":"smoke-release","dsl":{"query":{"match":{"name":"es-smoke"}},"size":5}}' | json_field data.total.value)"
  if [[ "${es_hits}" -ge 1 ]]; then
    break
  fi
  sleep 1
done
if [[ "${es_hits}" -lt 1 ]]; then
  echo "elasticsearch search returned hits ${es_hits}" >&2
  exit 1
fi
expect_status 403 POST "/api/connections/${es_readonly_id}/es/import" '{"index":"smoke-release","documents":[{"_id":"blocked","name":"blocked"}],"confirm":true}'

echo "checking history and audit endpoints"
api GET "/api/conn/${mysql_id}/history" >/dev/null
audit_count="$(api GET /api/audit/events | python3 -c 'import json,sys; print(len(json.load(sys.stdin).get("data", [])))')"
if [[ "${audit_count}" -lt 1 ]]; then
  echo "audit events were not recorded" >&2
  exit 1
fi

echo "release smoke passed"
