#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT_DIR"

SEARCH_TOOL=""
if command -v rg >/dev/null 2>&1; then
  SEARCH_TOOL="rg"
elif command -v grep >/dev/null 2>&1; then
  SEARCH_TOOL="grep"
else
  echo "scripts/check-versions.sh requires \"rg\" or \"grep\" but neither is installed." >&2
  exit 1
fi

search_fixed() {
  local pattern="$1"
  local file="$2"

  if [[ "$SEARCH_TOOL" == "rg" ]]; then
    rg -q --fixed-strings "$pattern" "$file"
  else
    grep -F -q "$pattern" "$file"
  fi
}

search_regex() {
  local pattern="$1"
  local file="$2"

  if [[ "$SEARCH_TOOL" == "rg" ]]; then
    rg -q "$pattern" "$file"
  else
    grep -E -q "$pattern" "$file"
  fi
}

fail=0

check_present() {
  local pattern="$1"
  local file="$2"
  local label="$3"

  if ! search_fixed "$pattern" "$file"; then
    echo "Missing ${label} in ${file} (pattern: ${pattern})." >&2
    fail=1
  fi
}

check_present_re() {
  local pattern="$1"
  local file="$2"
  local label="$3"

  if ! search_regex "$pattern" "$file"; then
    echo "Missing ${label} in ${file} (pattern: ${pattern})." >&2
    fail=1
  fi
}

check_absent() {
  local pattern="$1"
  local file="$2"
  local label="$3"

  if search_fixed "$pattern" "$file"; then
    echo "Unexpected ${label} in ${file} (pattern: ${pattern})." >&2
    fail=1
  fi
}

check_present_re '^go 1\.25(\.0)?$' backend/go.mod "Go 1.25 declaration"
check_present_re '^toolchain go1\.25\.3$' backend/go.mod "Go toolchain 1.25.3"

check_present "go-version: \"1.25.x\"" .github/workflows/backend-ci.yml "backend-ci go-version 1.25.x"
check_absent "go-version-file:" .github/workflows/backend-ci.yml "backend-ci go-version-file"

check_present_re "node-version:[[:space:]]*[\"']?20[\"']?" .github/workflows/frontend-ci.yml "frontend-ci node-version 20"
check_present_re "node-version:[[:space:]]*[\"']?20[\"']?" .github/workflows/codegen-check.yml "codegen-check node-version 20"

check_present_re "Go.*1\\.25\\+" README.md "README Go 1.25+"
check_present_re "Node\\.js.*20\\+" README.md "README Node.js 20+"

check_present_re "Go.*1\\.25\\+" docs/QUICK_START.md "docs/QUICK_START Go 1.25+"
check_present_re "Node\\.js.*20\\+" docs/QUICK_START.md "docs/QUICK_START Node.js 20+"

check_present_re "Go.*1\\.25\\+" docs/DEPLOYMENT.md "docs/DEPLOYMENT Go 1.25+"

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi

echo "Version policy check passed."
