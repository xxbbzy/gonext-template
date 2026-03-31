#!/bin/bash

# Regression coverage for scripts/new-module.sh scaffold output.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/gonext-new-module-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT

fail() {
  echo "❌ $1" >&2
  exit 1
}

assert_file_exists() {
  local file="$1"
  if [ ! -f "$file" ]; then
    fail "expected file to exist: $file"
  fi
}

assert_file_not_exists() {
  local file="$1"
  if [ -e "$file" ]; then
    fail "did not expect file to exist: $file"
  fi
}

assert_contains() {
  local file="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$file"; then
    fail "expected '$needle' in $file"
  fi
}

assert_not_contains() {
  local file="$1"
  local needle="$2"
  if grep -Fq "$needle" "$file"; then
    fail "did not expect '$needle' in $file"
  fi
}

mkdir -p \
  "$TMP_DIR/scripts" \
  "$TMP_DIR/backend/internal/handler" \
  "$TMP_DIR/backend/internal/service" \
  "$TMP_DIR/backend/internal/repository" \
  "$TMP_DIR/backend/internal/model" \
  "$TMP_DIR/backend/internal/dto"

cp "$ROOT_DIR/scripts/new-module.sh" "$TMP_DIR/scripts/new-module.sh"
chmod +x "$TMP_DIR/scripts/new-module.sh"

(
  cd "$TMP_DIR"
  ./scripts/new-module.sh product
) >"$TMP_DIR/output.txt"

(
  cd "$TMP_DIR"
  ./scripts/new-module.sh key
) >"$TMP_DIR/key-output.txt"

if (
  cd "$TMP_DIR"
  ./scripts/new-module.sh type
) >"$TMP_DIR/type-output.txt" 2>&1; then
  fail "expected Go keyword module name to be rejected"
fi

for generated_file in \
  "$TMP_DIR/backend/internal/model/product.go" \
  "$TMP_DIR/backend/internal/dto/product.go" \
  "$TMP_DIR/backend/internal/repository/product.go" \
  "$TMP_DIR/backend/internal/service/product.go" \
  "$TMP_DIR/backend/internal/handler/product.go" \
  "$TMP_DIR/backend/internal/repository/product_test.go" \
  "$TMP_DIR/backend/internal/service/product_test.go" \
  "$TMP_DIR/backend/internal/handler/product_test.go"; do
  assert_file_exists "$generated_file"
done

assert_contains "$TMP_DIR/backend/internal/model/product.go" "type Product struct {"
assert_contains "$TMP_DIR/backend/internal/model/product.go" "return \"products\""

assert_contains "$TMP_DIR/backend/internal/dto/product.go" "type ListProductsQuery struct {"

assert_contains "$TMP_DIR/backend/internal/repository/product.go" "func (r *ProductRepository) SoftDelete(id uint) error {"
assert_not_contains "$TMP_DIR/backend/internal/repository/product.go" "backend/pkg/response"

assert_contains "$TMP_DIR/backend/internal/service/product.go" "func (s *ProductService) List(offset, limit int) ([]dto.ProductResponse, int64, error) {"
assert_contains "$TMP_DIR/backend/internal/service/product.go" "errcode.ErrNotFoundMsg"
assert_not_contains "$TMP_DIR/backend/internal/service/product.go" "\"github.com/gin-gonic/gin\""

assert_contains "$TMP_DIR/backend/internal/handler/product.go" "response.PagedSuccess(c, products, total, p.Page, p.PageSize)"
assert_contains "$TMP_DIR/backend/internal/handler/product.go" "response.Error(c, appErr.HTTPStatus, appErr.Code, appErr.Message)"
assert_not_contains "$TMP_DIR/backend/internal/handler/product.go" ".JSON("

assert_contains "$TMP_DIR/backend/internal/service/product_test.go" "func assertProductAppError(t *testing.T, err error, wantCode, wantStatus int) {"
assert_not_contains "$TMP_DIR/backend/internal/service/product_test.go" "func assertAppError(t *testing.T, err error, wantCode, wantStatus int) {"
assert_contains "$TMP_DIR/backend/internal/handler/product_test.go" "func decodeProductPayload(t *testing.T, body []byte) map[string]any {"

assert_contains "$TMP_DIR/backend/internal/model/key.go" "return \"keys\""
assert_not_contains "$TMP_DIR/backend/internal/model/key.go" "return \"keies\""
assert_contains "$TMP_DIR/backend/internal/dto/key.go" "type ListKeysQuery struct {"
assert_contains "$TMP_DIR/backend/internal/handler/key.go" "keys := r.Group(\"/keys\")"
assert_not_contains "$TMP_DIR/backend/internal/handler/key.go" "keies := r.Group(\"/keies\")"

assert_contains "$TMP_DIR/type-output.txt" "module_name must not be a Go keyword after normalization."
assert_file_not_exists "$TMP_DIR/backend/internal/model/type.go"
assert_file_not_exists "$TMP_DIR/backend/internal/repository/type.go"

assert_contains "$TMP_DIR/output.txt" "Review api/openapi.yaml before finalizing routes or response shapes"
assert_contains "$TMP_DIR/output.txt" "backend/cmd/server/providers.go"
assert_contains "$TMP_DIR/output.txt" "backend/cmd/server/main.go"
assert_contains "$TMP_DIR/output.txt" "AutoMigrate"
assert_contains "$TMP_DIR/output.txt" "make gen-types"
assert_contains "$TMP_DIR/output.txt" "make gen"
assert_contains "$TMP_DIR/output.txt" "make check"
assert_contains "$TMP_DIR/output.txt" "make e2e"

echo "Module scaffold regression test passed."
