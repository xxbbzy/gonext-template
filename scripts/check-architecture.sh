#!/bin/bash

# Repository-owned architecture guardrails for documented backend boundaries.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FAILURES=0
SEARCH_TOOL=""

if command -v rg >/dev/null 2>&1; then
  SEARCH_TOOL="rg"
elif command -v grep >/dev/null 2>&1; then
  SEARCH_TOOL="grep"
else
  echo "scripts/check-architecture.sh requires \"rg\" or \"grep\" but neither is installed." >&2
  exit 1
fi

search_matches() {
  local pattern="$1"
  shift

  if [[ "$SEARCH_TOOL" == "rg" ]]; then
    rg -n --with-filename --color never "$pattern" "$@"
  else
    grep -n -H -E "$pattern" "$@"
  fi
}

search_matches_in_file() {
  local pattern="$1"
  local file="$2"

  if [[ "$SEARCH_TOOL" == "rg" ]]; then
    rg -n --color never "$pattern" "$file"
  else
    grep -n -E "$pattern" "$file"
  fi
}

collect_go_files() {
  local search_root="$1"
  local output=""
  local status=0

  if [ ! -d "$ROOT_DIR/$search_root" ]; then
    return
  fi

  set +e
  if [[ "$SEARCH_TOOL" == "rg" ]]; then
    output="$(cd "$ROOT_DIR" && rg --files "$search_root" -g '*.go' -g '!**/*_test.go' -g '!*_test.go' 2>&1)"
  else
    output="$(cd "$ROOT_DIR" && find "$search_root" -type f -name '*.go' ! -name '*_test.go' -print 2>&1)"
  fi
  status=$?
  set -e

  if [ "$status" -eq 1 ] && [[ "$SEARCH_TOOL" == "rg" ]]; then
    return
  fi

  if [ "$status" -ne 0 ]; then
    echo "guardrail file collection failed for $search_root" >&2
    echo "$output" >&2
    exit 1
  fi

  printf '%s\n' "$output"
}

check_pattern() {
  local title="$1"
  local advice="$2"
  local pattern="$3"
  local search_root="$4"
  local output=""
  local status=0
  local -a files=()

  while IFS= read -r file; do
    [ -n "$file" ] && files+=("$file")
  done < <(collect_go_files "$search_root")

  if [ "${#files[@]}" -eq 0 ]; then
    return
  fi

  set +e
  output="$(
    cd "$ROOT_DIR"
    search_matches "$pattern" "${files[@]}"
  )"
  status=$?
  set -e

  if [ "$status" -eq 1 ]; then
    return
  fi

  if [ "$status" -ne 0 ]; then
    echo "guardrail check failed to execute: $title" >&2
    echo "$output" >&2
    exit 1
  fi

  FAILURES=$((FAILURES + 1))
  echo "❌ $title"
  echo "$advice"
  echo "$output"
  echo ""
}

is_allowed_raw_json_match() {
  local file="$1"
  local line="$2"
  local marker_line=1

  if [ "$line" -gt 1 ]; then
    marker_line=$((line - 1))
  fi

  if sed -n "${marker_line}p" "$file" | grep -Fq "guardrail:allow raw-json"; then
    return 0
  fi

  return 1
}

check_raw_json_pattern() {
  local title="$1"
  local advice="$2"
  local search_root="$3"
  local pattern='\.JSON[[:space:]]*\('
  local -a files=()
  local -a violations=()
  local file=""
  local match=""
  local line=""
  local text=""

  while IFS= read -r file; do
    [ -n "$file" ] && files+=("$file")
  done < <(collect_go_files "$search_root")

  if [ "${#files[@]}" -eq 0 ]; then
    return
  fi

  for file in "${files[@]}"; do
    while IFS= read -r match; do
      [ -z "$match" ] && continue
      line="${match%%:*}"
      text="${match#*:}"
      if is_allowed_raw_json_match "$ROOT_DIR/$file" "$line"; then
        continue
      fi
      violations+=("${file}:${line}:${text}")
    done < <(cd "$ROOT_DIR" && search_matches_in_file "$pattern" "$file" || true)
  done

  if [ "${#violations[@]}" -eq 0 ]; then
    return
  fi

  FAILURES=$((FAILURES + 1))
  echo "❌ $title"
  echo "$advice"
  printf '%s\n' "${violations[@]}"
  echo ""
}

check_pattern \
  "handlers must not import repositories directly" \
  "Use the documented handler -> service -> repository flow instead of reaching into repository packages from handlers." \
  '"github.com/xxbbzy/gonext-template/backend/internal/repository"' \
  backend/internal/handler

check_pattern \
  "handlers must not import GORM packages" \
  "Move persistence access behind a repository and keep handler code focused on HTTP binding and response helpers." \
  '"gorm.io/' \
  backend/internal/handler

check_raw_json_pattern \
  "handlers must not emit raw c.JSON(...) responses" \
  "Use backend/pkg/response helpers so the shared response envelope stays consistent." \
  backend/internal/handler

check_pattern \
  "services must not import Gin packages" \
  "Keep HTTP concerns in handlers and middleware; services should return DTOs and application errors instead." \
  '"github.com/gin-gonic/gin"' \
  backend/internal/service

check_pattern \
  "repositories must not import response helpers" \
  "Response-envelope helpers belong to HTTP layers only; repositories should stay focused on data access." \
  '"github.com/xxbbzy/gonext-template/backend/pkg/response"' \
  backend/internal/repository

if [ "$FAILURES" -gt 0 ]; then
  echo "Architecture guardrail check failed with $FAILURES violation group(s)." >&2
  exit 1
fi

echo "Architecture guardrail check passed."
