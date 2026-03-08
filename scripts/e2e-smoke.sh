#!/bin/bash
# E2E Smoke Test: register → login → CRUD → assert
# Usage: bash scripts/e2e-smoke.sh
#
# Starts the backend on a random port with an ephemeral SQLite database,
# runs through the full user journey, and reports pass/fail.

set -euo pipefail

# ===== Configuration =====
PORT=${E2E_PORT:-0}  # 0 = let OS pick a free port
DB_FILE=$(mktemp "${TMPDIR:-/tmp}/e2e_smoke_XXXXXX")
BACKEND_PID=""

cleanup() {
  if [ -n "$BACKEND_PID" ] && kill -0 "$BACKEND_PID" 2>/dev/null; then
    kill "$BACKEND_PID" 2>/dev/null || true
    wait "$BACKEND_PID" 2>/dev/null || true
  fi
  rm -f "$DB_FILE" "$RESP_BODY_FILE" 2>/dev/null || true
}
trap cleanup EXIT

PASS=0
FAIL=0
RESP_BODY_FILE=$(mktemp /tmp/e2e_resp_XXXXXX)

assert_status() {
  local label="$1" expected="$2" actual="$3"
  if [ "$actual" -eq "$expected" ]; then
    echo "  ✅ $label (HTTP $actual)"
    PASS=$((PASS + 1))
  else
    echo "  ❌ $label — expected HTTP $expected, got $actual"
    FAIL=$((FAIL + 1))
  fi
}

# Portable HTTP call: writes body to RESP_BODY_FILE, returns status code.
http_call() {
  curl -s -o "$RESP_BODY_FILE" -w "%{http_code}" "$@"
}

# ===== Start backend =====
echo "==> Starting backend (SQLite: $DB_FILE)..."

# Find a free port
if [ "$PORT" -eq 0 ]; then
  PORT=$(python3 -c 'import socket; s=socket.socket(); s.bind(("",0)); print(s.getsockname()[1]); s.close()')
fi

BASE="http://localhost:$PORT"

export APP_PORT="$PORT"
export DB_DRIVER="sqlite"
export DB_DSN="$DB_FILE"
export JWT_SECRET="e2e-test-secret"
export JWT_ACCESS_EXPIRY="15m"
export JWT_REFRESH_EXPIRY="24h"
export APP_BASE_URL="$BASE"

# Rebuild on every run so the smoke test always exercises current sources.
BIN="bin/server"
echo "==> Building backend..."
cd backend && CGO_ENABLED=1 go build -o ../bin/server ./cmd/server/ && cd ..

$BIN &
BACKEND_PID=$!

# Wait for readiness
echo "==> Waiting for backend readiness..."
for i in $(seq 1 30); do
  if curl -sf "$BASE/healthz" > /dev/null 2>&1; then
    echo "==> Backend ready on port $PORT"
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "❌ Backend failed to start within 30 seconds"
    exit 1
  fi
  sleep 1
done

# ===== Test: Register =====
echo ""
echo "==> Register user..."
REGISTER_STATUS=$(http_call -X POST "$BASE/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"e2euser","email":"e2e@test.com","password":"password123"}')
assert_status "Register" 201 "$REGISTER_STATUS"

# ===== Test: Login =====
echo "==> Login..."
LOGIN_STATUS=$(http_call -X POST "$BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"e2e@test.com","password":"password123"}')
assert_status "Login" 200 "$LOGIN_STATUS"

# Extract access token
TOKEN=$(cat "$RESP_BODY_FILE" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['access_token'])" 2>/dev/null || echo "")
if [ -z "$TOKEN" ]; then
  echo "  ❌ Could not extract access_token from login response"
  FAIL=$((FAIL + 1))
else
  echo "  ✅ Token extracted"
  PASS=$((PASS + 1))
fi

AUTH="Authorization: Bearer $TOKEN"

# ===== Test: Create Item =====
echo "==> Create item..."
CREATE_STATUS=$(http_call -X POST "$BASE/api/v1/items" \
  -H "Content-Type: application/json" \
  -H "$AUTH" \
  -d '{"title":"E2E Item","description":"smoke test"}')
assert_status "Create Item" 201 "$CREATE_STATUS"

ITEM_ID=$(cat "$RESP_BODY_FILE" | python3 -c "import sys,json; print(int(json.load(sys.stdin)['data']['id']))" 2>/dev/null || echo "")

# ===== Test: Get Item =====
echo "==> Get item..."
GET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE/api/v1/items/$ITEM_ID" -H "$AUTH")
assert_status "Get Item" 200 "$GET_STATUS"

# ===== Test: Update Item =====
echo "==> Update item..."
UPDATE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X PUT "$BASE/api/v1/items/$ITEM_ID" \
  -H "Content-Type: application/json" \
  -H "$AUTH" \
  -d '{"title":"Updated E2E Item"}')
assert_status "Update Item" 200 "$UPDATE_STATUS"

# ===== Test: List Items =====
echo "==> List items..."
LIST_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE/api/v1/items" -H "$AUTH")
assert_status "List Items" 200 "$LIST_STATUS"

# ===== Test: Delete Item =====
echo "==> Delete item..."
DELETE_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE/api/v1/items/$ITEM_ID" -H "$AUTH")
assert_status "Delete Item" 200 "$DELETE_STATUS"

# ===== Test: Verify Deleted =====
echo "==> Verify item deleted..."
VERIFY_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "$BASE/api/v1/items/$ITEM_ID" -H "$AUTH")
assert_status "Item Gone (404)" 404 "$VERIFY_STATUS"

# ===== Summary =====
echo ""
echo "════════════════════════════════"
echo "  E2E Results: $PASS passed, $FAIL failed"
echo "════════════════════════════════"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
echo "✅ All E2E smoke tests passed!"
