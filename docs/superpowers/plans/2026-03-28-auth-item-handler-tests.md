# Auth/Item Test Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add and standardize backend tests so AuthService, ItemService, and critical auth/item/health handler endpoints are covered for primary success/failure paths.

**Architecture:** Keep production code unchanged and improve confidence by expanding tests at two boundaries: service-layer business logic tests and lightweight handler route tests. Use in-memory SQLite test fixtures and table-driven test cases for maintainability and consistency. Validate behavior via stable contract assertions: app error code/status at service layer and HTTP status/envelope semantics at handler layer.

**Tech Stack:** Go `testing`, Gin `httptest`, GORM + SQLite in-memory, bcrypt, project `errcode` package, `make check`, `make e2e`.

---

### Task 1: Add AuthService Unit Tests

**Files:**

- Create: `backend/internal/service/auth_test.go`
- Test: `backend/internal/service/auth_test.go`

- [ ] **Step 1: Write failing table-driven tests for Register/Login/Refresh scenarios**

```go
func TestAuthService_Register(t *testing.T) { /* success + duplicate email */ }
func TestAuthService_Login(t *testing.T) { /* success + wrong password + missing email */ }
func TestAuthService_RefreshToken(t *testing.T) { /* success + invalid + expired */ }
```

- [ ] **Step 2: Run targeted tests to verify at least one scenario fails initially (or file missing)**

Run: `cd backend && go test ./internal/service -run TestAuthService -v`
Expected: FAIL before implementation is complete.

- [ ] **Step 3: Implement helper fixtures/assertions in test file only**

```go
func newAuthServiceForTest(t *testing.T) (*AuthService, *gorm.DB, *pkgjwt.Manager)
func assertAppError(t *testing.T, err error, code int, status int)
```

- [ ] **Step 4: Re-run targeted tests and ensure all AuthService cases pass**

Run: `cd backend && go test ./internal/service -run TestAuthService -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/auth_test.go
git commit -m "test: add auth service unit tests"
```

### Task 2: Rewrite ItemService Tests To Table-Driven

**Files:**

- Modify: `backend/internal/service/item_test.go`
- Test: `backend/internal/service/item_test.go`

- [ ] **Step 1: Replace ad-hoc tests with table-driven groups**

```go
func TestItemService_Create(t *testing.T)
func TestItemService_GetByID(t *testing.T)
func TestItemService_Update(t *testing.T)
func TestItemService_Delete(t *testing.T)
func TestItemService_List(t *testing.T)
```

- [ ] **Step 2: Run targeted item service tests to catch regressions**

Run: `cd backend && go test ./internal/service -run TestItemService -v`
Expected: initial failures while refactor is incomplete, then PASS after fixes.

- [ ] **Step 3: Add stronger assertions for not-found branches using AppError code/status**

```go
assertAppError(t, err, errcode.ErrNotFound, http.StatusNotFound)
```

- [ ] **Step 4: Re-run service package tests**

Run: `cd backend && go test ./internal/service -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/item_test.go
git commit -m "test: rewrite item service tests as table driven"
```

### Task 3: Extend Auth Handler API Tests

**Files:**

- Modify: `backend/internal/handler/auth_test.go`
- Test: `backend/internal/handler/auth_test.go`

- [ ] **Step 1: Add tests for `/healthz`, `/readyz` ready branch, `/auth/login` success/failure, duplicate register**

```go
func TestRegisterHealthRoutesLiveness(t *testing.T)
func TestRegisterHealthRoutesReadinessReady(t *testing.T)
func TestAuthLoginSuccessAndWrongPassword(t *testing.T)
func TestAuthRegisterDuplicateEmailReturnsConflict(t *testing.T)
```

- [ ] **Step 2: Run targeted handler auth tests**

Run: `cd backend && go test ./internal/handler -run 'Test(Auth|RegisterHealth)' -v`
Expected: PASS.

- [ ] **Step 3: Keep assertions at contract level (status + key envelope fields)**

```go
if resp.Code != http.StatusUnauthorized { ... }
if payload["code"] != float64(errcode.ErrInvalidCreds) { ... }
```

- [ ] **Step 4: Re-run full handler package tests**

Run: `cd backend && go test ./internal/handler -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/auth_test.go
git commit -m "test: add critical auth and health handler coverage"
```

### Task 4: Full Verification And Wrap-Up

**Files:**

- Modify: none (verification only unless fixes are needed)

- [ ] **Step 1: Run mandatory repository checks**

Run: `make check`
Expected: exit code 0.

- [ ] **Step 2: Run API smoke verification for runtime flow**

Run: `make e2e`
Expected: register -> login -> CRUD cycle succeeds.

- [ ] **Step 3: If failures occur, patch minimal test/code issues and re-run until green**

Run: repeat failing command(s)
Expected: all checks green.

- [ ] **Step 4: Final commit for any follow-up fixes**

```bash
git add <changed-files>
git commit -m "test: finalize auth item handler coverage"
```

- [ ] **Step 5: Provide completion summary with exact changed files and verification evidence**
