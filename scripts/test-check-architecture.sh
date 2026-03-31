#!/bin/bash

# Regression coverage for scripts/check-architecture.sh.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gonext-check-architecture-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "❌ $1" >&2
  exit 1
}

assert_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$file"; then
    fail "expected '$needle' in $file"
  fi
}

mkdir -p \
  "$TMP_DIR/scripts" \
  "$TMP_DIR/backend/internal/handler" \
  "$TMP_DIR/backend/internal/service" \
  "$TMP_DIR/backend/internal/repository"

cp "$ROOT_DIR/scripts/check-architecture.sh" "$TMP_DIR/scripts/check-architecture.sh"
chmod +x "$TMP_DIR/scripts/check-architecture.sh"

cat >"$TMP_DIR/backend/internal/handler/example.go" <<'EOF'
package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/xxbbzy/gonext-template/backend/pkg/response"
)

func Example(c *gin.Context) {
	response.Success(c, nil)
}
EOF

(
  cd "$TMP_DIR"
  ./scripts/check-architecture.sh
) >"$TMP_DIR/output.txt" 2>&1

assert_contains "$TMP_DIR/output.txt" "Architecture guardrail check passed."

cat >"$TMP_DIR/backend/internal/handler/bad_example.go" <<'EOF'
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExampleBad(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "bad"})
}
EOF

if (
  cd "$TMP_DIR"
  ./scripts/check-architecture.sh
) >"$TMP_DIR/bad-output.txt" 2>&1; then
  fail "expected raw c.JSON guardrail violation to fail"
fi

assert_contains "$TMP_DIR/bad-output.txt" "Architecture guardrail check failed"
assert_contains "$TMP_DIR/bad-output.txt" "handlers must not emit raw c.JSON(...) responses"

echo "Architecture guardrail regression test passed."
