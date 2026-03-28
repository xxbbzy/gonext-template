# JWT Unit Tests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add unit tests for `backend/pkg/jwt` to cover access/refresh token generation and token parsing in valid, expired, and invalid scenarios.

**Architecture:** Keep production JWT code unchanged and validate behavior through public APIs only (`NewManager`, `GenerateAccessToken`, `GenerateRefreshToken`, `ParseToken`). Use black-box-first table-driven tests in one focused `_test.go` file with minimal helpers for readability.

**Tech Stack:** Go, `testing` (stdlib), `github.com/golang-jwt/jwt/v5`, Makefile validation pipeline

---

## Scope Check

This plan covers one subsystem (`backend/pkg/jwt` tests) and is implementable in a single execution cycle without further decomposition.

## File Structure and Responsibilities

- Create: `backend/pkg/jwt/jwt_test.go`
  - All unit tests for token generation/parsing behavior.
  - Local test helpers only; no shared test utilities required.
- Modify: none expected for production code.
- Verify: `backend/pkg/jwt/jwt.go` behavior only through public methods.

## Spec Alignment

Spec reference: `docs/superpowers/specs/2026-03-28-jwt-unit-tests-design.md`

Plan must satisfy all acceptance points:

- Access token generation covered
- Refresh token generation covered
- Valid token parsing covered
- Expired token scenario covered (first valid then expired)
- Invalid token scenario covered (malformed + wrong signature)
- `backend/pkg/jwt` tests pass
- `make check` passes

### Task 1: Create Test Skeleton and Constructor Error Coverage

**Files:**

- Create: `backend/pkg/jwt/jwt_test.go`
- Test: `backend/pkg/jwt/jwt_test.go`

- [ ] **Step 1: Write failing constructor validation test**

Add:

```go
func TestNewManager_InvalidDuration(t *testing.T) {
	_, err := NewManager("secret", "bad", "24h")
	if err == nil {
		t.Fatal("NewManager() expected error for invalid access duration")
	}

	_, err = NewManager("secret", "15m", "bad")
	if err == nil {
		t.Fatal("NewManager() expected error for invalid refresh duration")
	}
}
```

- [ ] **Step 2: Run targeted test to verify behavior**

Run:

```bash
cd backend && go test ./pkg/jwt -run TestNewManager_InvalidDuration -v
```

Expected: PASS (or FAIL only if test file path/package mismatch to be fixed immediately).

- [ ] **Step 3: Add minimal test helpers for later tasks**

Add helpers:

```go
func newTestManager(t *testing.T, secret, accessTTL, refreshTTL string) *Manager {
	t.Helper()
	m, err := NewManager(secret, accessTTL, refreshTTL)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}
	return m
}

func mustGenerateAccessToken(t *testing.T, m *Manager, userID uint, role string) string {
	t.Helper()
	token, err := m.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}
	return token
}

func mustGenerateRefreshToken(t *testing.T, m *Manager, userID uint, role string) string {
	t.Helper()
	token, err := m.GenerateRefreshToken(userID, role)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}
	return token
}
```

- [ ] **Step 4: Commit Task 1**

Run:

```bash
git add backend/pkg/jwt/jwt_test.go
git commit -m "test(jwt): add test scaffold and manager duration validation"
```

### Task 2: Add Token Generation Success Tests (Black-Box Table-Driven)

**Files:**

- Modify: `backend/pkg/jwt/jwt_test.go`
- Test: `backend/pkg/jwt/jwt_test.go`

- [ ] **Step 1: Write generation tests with shared table**

Add:

```go
func TestManager_GenerateTokens(t *testing.T) {
	m := newTestManager(t, "test-secret", "15m", "24h")

	tests := []struct {
		name string
		gen  func(*testing.T, *Manager, uint, string) string
	}{
		{
			name: "generate access token",
			gen:  mustGenerateAccessToken,
		},
		{
			name: "generate refresh token",
			gen:  mustGenerateRefreshToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.gen(t, m, 42, "admin")
			if token == "" {
				t.Fatal("token should not be empty")
			}

			claims, err := m.ParseToken(token)
			if err != nil {
				t.Fatalf("ParseToken() error = %v", err)
			}
			if claims.UserID != 42 {
				t.Fatalf("claims.UserID = %d, want 42", claims.UserID)
			}
			if claims.Role != "admin" {
				t.Fatalf("claims.Role = %q, want %q", claims.Role, "admin")
			}
		})
	}
}
```

- [ ] **Step 2: Run only generation test first**

Run:

```bash
cd backend && go test ./pkg/jwt -run TestManager_GenerateTokens -v
```

Expected: PASS and both subtests execute.

- [ ] **Step 3: Commit Task 2**

Run:

```bash
git add backend/pkg/jwt/jwt_test.go
git commit -m "test(jwt): cover access and refresh token generation"
```

### Task 3: Add Parse Flow Tests (Valid, Expired, Malformed, Wrong Signature)

**Files:**

- Modify: `backend/pkg/jwt/jwt_test.go`
- Test: `backend/pkg/jwt/jwt_test.go`

- [ ] **Step 1: Write parse valid token test**

Add:

```go
func TestManager_ParseToken_Valid(t *testing.T) {
	m := newTestManager(t, "test-secret", "15m", "24h")
	token := mustGenerateAccessToken(t, m, 7, "user")

	claims, err := m.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 7 {
		t.Fatalf("claims.UserID = %d, want 7", claims.UserID)
	}
	if claims.Role != "user" {
		t.Fatalf("claims.Role = %q, want %q", claims.Role, "user")
	}
}
```

- [ ] **Step 2: Write expired token test with pre-expiry success check**

Add:

```go
func TestManager_ParseToken_Expired(t *testing.T) {
	m := newTestManager(t, "test-secret", "20ms", "24h")
	token := mustGenerateAccessToken(t, m, 9, "user")

	if _, err := m.ParseToken(token); err != nil {
		t.Fatalf("ParseToken() before expiry error = %v", err)
	}

	time.Sleep(40 * time.Millisecond)

	_, err := m.ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken() expected error for expired token")
	}
	if !strings.Contains(err.Error(), "token expired") {
		t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "token expired")
	}
}
```

- [ ] **Step 3: Write malformed token and wrong-signature tests**

Add:

```go
func TestManager_ParseToken_Malformed(t *testing.T) {
	m := newTestManager(t, "test-secret", "15m", "24h")
	_, err := m.ParseToken("not-a-jwt")
	if err == nil {
		t.Fatal("ParseToken() expected error for malformed token")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "invalid token")
	}
}

func TestManager_ParseToken_InvalidSignature(t *testing.T) {
	issuer := newTestManager(t, "secret-a", "15m", "24h")
	verifier := newTestManager(t, "secret-b", "15m", "24h")
	token := mustGenerateAccessToken(t, issuer, 11, "admin")

	_, err := verifier.ParseToken(token)
	if err == nil {
		t.Fatal("ParseToken() expected error for invalid signature")
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("ParseToken() error = %q, want contains %q", err.Error(), "invalid token")
	}
}
```

- [ ] **Step 4: Ensure imports compile**

Ensure test file imports include:

```go
import (
	"strings"
	"testing"
	"time"
)
```

- [ ] **Step 5: Run parse-focused tests**

Run:

```bash
cd backend && go test ./pkg/jwt -run TestManager_ParseToken -v
```

Expected: PASS with valid/expired/malformed/invalid-signature coverage.

- [ ] **Step 6: Commit Task 3**

Run:

```bash
git add backend/pkg/jwt/jwt_test.go
git commit -m "test(jwt): cover token parse valid expired and invalid flows"
```

### Task 4: Full Package and Repository Verification (Mandatory)

**Files:**

- Verify only unless fixes are required

- [ ] **Step 1: Run package-level JWT tests**

Run:

```bash
cd backend && go test ./pkg/jwt -v
```

Expected: all JWT tests pass and output shows all target cases.

- [ ] **Step 2: Run mandatory repository check**

Run:

```bash
make check
```

Expected: lint + typecheck + tests + build all green (exit code 0).

- [ ] **Step 3: If `make check` fails, fix minimally and re-run**

Use `@systematic-debugging` for failures and apply smallest scoped fix.

- [ ] **Step 4: Commit verification-driven fixes if any**

Run only when required:

```bash
git add backend/pkg/jwt/jwt_test.go
git commit -m "fix: address verification issues in jwt test coverage"
```

## Final Review Checklist

- [ ] `backend/pkg/jwt/jwt_test.go` exists and compiles
- [ ] Access token generation covered
- [ ] Refresh token generation covered
- [ ] Valid parse covered
- [ ] Expiry flow covered with pre-expiry success check
- [ ] Malformed token covered
- [ ] Wrong-signature token covered
- [ ] `cd backend && go test ./pkg/jwt -v` passes
- [ ] `make check` passes

## Skills for Execution

- `@subagent-driven-development` (recommended orchestration mode)
- `@executing-plans` (inline execution alternative)
- `@verification-before-completion` before final completion claim
- `@test-driven-development` for incremental red/green cycles
