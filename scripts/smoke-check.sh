#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
TIMEOUT="${TIMEOUT:-5}"

echo "=== API Smoke Check ==="
echo "Target: $BASE_URL"
echo

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { echo "Missing: $1"; exit 1; }
}

require_cmd curl

echo "1) Health"
status=$(curl -sS -o /tmp/smoke-health.out -m "$TIMEOUT" -w "%{http_code}" "$BASE_URL/health")
if [[ "$status" != "200" ]]; then
  echo "✗ /health failed (HTTP $status)"
  echo "Response:" && cat /tmp/smoke-health.out
  exit 1
fi
echo "✓ /health -> $status"
echo

echo "2) List tasks (default)"
tmp_headers="/tmp/smoke-headers.out"
tmp_body="/tmp/smoke-body.out"
status=$(curl -sS -D "$tmp_headers" -o "$tmp_body" -m "$TIMEOUT" -w "%{http_code}" "$BASE_URL/tasks")
if [[ "$status" != "200" ]]; then
  echo "✗ /tasks failed (HTTP $status)"
  echo "Response:" && cat "$tmp_body"
  exit 1
fi
echo "✓ /tasks -> $status"
for h in X-Total-Count X-Page X-Limit X-Has-Next; do
  val=$(awk -v header="$h:" 'tolower($1)==tolower(header) {print $2}' "$tmp_headers")
  echo "  $h: ${val:-n/a}"
done
echo

echo "3) Create task"
create_file="/tmp/smoke-create.out"
status=$(curl -sS -o "$create_file" -m "$TIMEOUT" -w "%{http_code}" \
  -X POST "$BASE_URL/tasks" \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke test task","description":"Deployment verification"}')
if [[ "$status" != "201" ]]; then
  echo "✗ create failed (HTTP $status)"
  echo "Response:" && cat "$create_file"
  exit 1
fi
TASK_ID=$(grep -o '"id"[[:space:]]*:[[:space:]]*[0-9]\+' "$create_file" | head -n 1 | grep -o '[0-9]\+$')
if [[ -z "${TASK_ID:-}" ]]; then
  echo "✗ could not extract task ID"
  echo "Response:" && cat "$create_file"
  exit 1
fi
echo "✓ created task id=$TASK_ID"
echo

echo "4) Get task by ID"
status=$(curl -sS -o /tmp/smoke-get.out -m "$TIMEOUT" -w "%{http_code}" "$BASE_URL/tasks/$TASK_ID")
if [[ "$status" != "200" ]]; then
  echo "✗ get task failed (HTTP $status)"
  echo "Response:" && cat /tmp/smoke-get.out
  exit 1
fi
echo "✓ /tasks/$TASK_ID -> $status"
echo

echo "5) Update task"
status=$(curl -sS -o /tmp/smoke-update.out -m "$TIMEOUT" -w "%{http_code}" \
  -X PUT "$BASE_URL/tasks/$TASK_ID" \
  -H "Content-Type: application/json" \
  -d '{"title":"Smoke test updated","description":"Updated","completed":true}')
if [[ "$status" != "200" ]]; then
  echo "✗ update failed (HTTP $status)"
  echo "Response:" && cat /tmp/smoke-update.out
  exit 1
fi
echo "✓ /tasks/$TASK_ID -> $status"
echo

echo "6) Delete task"
status=$(curl -sS -o /tmp/smoke-delete.out -m "$TIMEOUT" -w "%{http_code}" -X DELETE "$BASE_URL/tasks/$TASK_ID")
if [[ "$status" != "204" ]]; then
  echo "✗ delete failed (HTTP $status)"
  echo "Response:" && cat /tmp/smoke-delete.out
  exit 1
fi
echo "✓ /tasks/$TASK_ID -> $status"
echo

echo "7) Confirm deleted"
status=$(curl -sS -o /tmp/smoke-get-deleted.out -m "$TIMEOUT" -w "%{http_code}" "$BASE_URL/tasks/$TASK_ID")
if [[ "$status" != "404" ]]; then
  echo "✗ confirm delete failed (expected 404, got $status)"
  echo "Response:" && cat /tmp/smoke-get-deleted.out
  exit 1
fi
echo "✓ /tasks/$TASK_ID -> $status (deleted)"
echo

echo "=== All checks passed ==="
