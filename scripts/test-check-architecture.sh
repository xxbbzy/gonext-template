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
  PATH="/usr/bin:/bin" ./scripts/check-architecture.sh
) >"$TMP_DIR/output.txt"

assert_contains "$TMP_DIR/output.txt" "Architecture guardrail check passed."

echo "Architecture guardrail regression test passed."
