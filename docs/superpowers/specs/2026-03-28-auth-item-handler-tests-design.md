# Auth/Item Service + Handler Test Coverage Design

Date: 2026-03-28
Scope: `backend/internal/service` and `backend/internal/handler`
Status: Approved for implementation planning

## 1. Background And Goal

This design addresses three test coverage issues:

- Issue 5: add unit tests for `AuthService` register/login/refresh core flows
- Issue 6: strengthen and standardize `ItemService` unit tests for CRUD core flows
- Issue 7: add lightweight API tests for critical handler endpoints (`/healthz`, `/readyz`, `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/items`)

Primary goal: ensure business-critical auth/item behavior is covered at the service layer and basic API contract behavior is verified at the handler layer, while staying aligned with repository conventions.

## 2. Scope

In scope:

- New table-driven `AuthService` tests
- Rewrite `ItemService` tests to table-driven structure with stronger app-error assertions
- Add/extend lightweight handler tests for health/auth/item core routes

Out of scope:

- Refactoring production auth/item code
- Changing OpenAPI contract
- Full-stack e2e flow expansion beyond current `make e2e` behavior

## 3. Recommended Approach

Use a two-layer testing strategy:

1. Service unit tests with real in-memory SQLite repositories (`testutil.NewTestDB`)
2. Lightweight handler/API tests with `gin` + `httptest` without full `main.go` boot

Why:

- Matches repository patterns in `CONVENTIONS.md`
- Keeps tests deterministic and fast
- Verifies behavior at the correct boundary (business rules in service, HTTP contract in handler)

## 4. Test Architecture

### 4.1 Service Layer

- `AuthService`: new test file `backend/internal/service/auth_test.go`
- `ItemService`: rewrite existing `backend/internal/service/item_test.go`
- Use table-driven subtests (`t.Run`) and per-case isolated DB fixtures
- Assert `*errcode.AppError` code + HTTP status on error branches

### 4.2 Handler Layer

- Extend `backend/internal/handler/auth_test.go`
- Reuse existing `backend/internal/handler/item_test.go` for item core path verification
- Keep tests lightweight: route registration + request/response assertions only
- Validate key envelope fields (`code`, `data`, `message`) and status codes

## 5. Detailed Coverage Matrix

### 5.1 Issue 5: AuthService Unit Tests

`Register`:

- success registration
- duplicate email returns `ErrEmailAlreadyExists` (409)

`Login`:

- success login
- wrong password returns `ErrInvalidCredentials` (401)

`RefreshToken`:

- success refresh using valid refresh token
- invalid token returns `ErrTokenInvalidMsg` (401)
- expired token returns `ErrRefreshTokenExpired` (401)

Additional low-cost hardening:

- login with non-existent email returns invalid credentials
- registration stores bcrypt hash (not plaintext)

### 5.2 Issue 6: ItemService Unit Tests (Table-Driven Rewrite)

`Create`:

- create with default status
- create with explicit status

`GetByID`:

- existing item
- non-existent item returns not found

`Update`:

- update mutable fields
- non-existent item returns not found

`Delete`:

- delete existing item then verify fetch not found
- delete non-existent item returns not found

`List` (kept for regression value):

- paginated list returns expected total/length
- keyword/status filtering scenarios

### 5.3 Issue 7: Critical Handler/API Tests

Health:

- `GET /healthz` returns 200 and alive payload
- `GET /readyz` covers both ready=200 and not-ready=503

Auth:

- `POST /api/v1/auth/register` success path (already present, kept)
- register duplicate email returns 409 with conflict error code
- `POST /api/v1/auth/login` success path
- login wrong password returns 401

Items:

- keep existing basic `/api/v1/items` create/get/list/update/delete tests as core path coverage

## 6. Error Handling And Assertions

- Service errors must be asserted as `*errcode.AppError`, checking both:
  - `Code`
  - `HTTPStatus`
- Handler tests assert HTTP status first, then key envelope semantics
- Avoid brittle full-JSON equality snapshots

## 7. Risks And Mitigations

Risk: flaky expired-token test timing.
Mitigation: use tiny but safe durations (for example `1ms`) and buffered sleep before parse.

Risk: accidental cross-test DB contamination.
Mitigation: always create DB per subtest via fixture helper.

Risk: over-coupling handler tests to implementation details.
Mitigation: assert contract-level fields only.

## 8. Verification Plan

Mandatory:

1. `make check` (required by AGENTS.md)

Recommended after API behavior-related handler work:

2. `make e2e`

Expected:

- New/rewritten tests pass reliably
- `make check` exits 0
- Auth/item core service and handler paths are covered according to issue checklist

## 9. Implementation Checklist

- [ ] Add `backend/internal/service/auth_test.go` (table-driven)
- [ ] Rewrite `backend/internal/service/item_test.go` to table-driven style
- [ ] Extend `backend/internal/handler/auth_test.go` for `/healthz` and `/api/v1/auth/login`
- [ ] Add duplicate-email API test for `/api/v1/auth/register`
- [ ] Ensure `/readyz` both branches are covered
- [ ] Keep `/api/v1/items` basic route tests green
- [ ] Run `make check`
- [ ] Run `make e2e`
