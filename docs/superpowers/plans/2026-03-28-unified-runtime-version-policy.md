# Unified Runtime Version Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Go/Node version requirements consistent across runtime config, lint, CI, and docs, and add CI enforcement to prevent future drift.

**Architecture:** Keep existing toolchain and workflow topology unchanged, only normalize version declarations and add one guard script plus CI hook. Go policy is expressed as `go 1.25` with a recommended `toolchain go1.25.3`; Node policy is `20+` in docs and `20` in CI workflow declarations.

**Tech Stack:** Go modules, golangci-lint, GitHub Actions, Bash, Make, Markdown docs

---

## Scope Check

This work is a single subsystem (runtime version policy governance) and can be implemented in one plan without decomposition.

## File Structure and Responsibilities

- Modify: `backend/go.mod`
  - Canonical Go baseline and recommended toolchain patch.
- Modify: `backend/.golangci.yml`
  - Lint runtime target alignment with Go 1.25 policy.
- Modify: `.github/workflows/backend-ci.yml`
  - Backend CI Go setup policy (`1.25.x`).
- Modify: `.github/workflows/codegen-check.yml`
  - Enforce version consistency guard in PR CI.
- Modify: `README.md`
  - Root environment requirement statement (Go/Node).
- Modify: `docs/QUICK_START.md`
  - Quick-start environment requirement statement.
- Modify: `docs/DEPLOYMENT.md`
  - Deployment/runtime Go requirement statement.
- Create: `scripts/check-versions.sh`
  - Single drift guard script validating required version declarations.

### Task 1: Align Go Baseline in Runtime, Lint, and Backend CI

**Files:**

- Modify: `backend/go.mod`
- Modify: `backend/.golangci.yml`
- Modify: `.github/workflows/backend-ci.yml`
- Test: local command checks in shell

- [ ] **Step 1: Capture current mismatch evidence (pre-change)**

Run:

```bash
rg -n "^(go |toolchain )" backend/go.mod
rg -n "^run:|^  go:" backend/.golangci.yml
rg -n "setup-go|go-version|go-version-file" .github/workflows/backend-ci.yml
```

Expected: `go 1.25.0`, no `toolchain`, no explicit `run.go`, and backend CI still tied to `go-version-file`.

- [ ] **Step 2: Apply minimal Go policy edits**

Update `backend/go.mod` to:

```go
go 1.25
toolchain go1.25.3
```

Update `backend/.golangci.yml` under `run:` to include:

```yaml
run:
  timeout: 5m
  go: "1.25"
```

Update `.github/workflows/backend-ci.yml` each `setup-go` block to:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: "1.25.x"
    cache-dependency-path: backend/go.sum
```

- [ ] **Step 3: Verify Go policy is applied exactly**

Run:

```bash
rg -n "^(go 1\\.25|toolchain go1\\.25\\.3)$" backend/go.mod
rg -n '^  go: "1\\.25"$' backend/.golangci.yml
rg -n 'go-version: "1\\.25\\.x"' .github/workflows/backend-ci.yml
```

Expected: all matches found, no remaining `go-version-file` in backend CI.

- [ ] **Step 4: Commit Task 1**

Run:

```bash
git add backend/go.mod backend/.golangci.yml .github/workflows/backend-ci.yml
git commit -m "chore: align go runtime lint and backend ci versions"
```

### Task 2: Align Node/Go Version Statements in Docs

**Files:**

- Modify: `README.md`
- Modify: `docs/QUICK_START.md`
- Modify: `docs/DEPLOYMENT.md`
- Test: text assertions via ripgrep

- [ ] **Step 1: Confirm current doc drift (pre-change)**

Run:

```bash
rg -n "Node\\.js|Go|1\\.25\\.3|18\\+" README.md docs/QUICK_START.md docs/DEPLOYMENT.md
```

Expected: at least one `Node.js 18+` and/or fixed patch wording inconsistent with policy.

- [ ] **Step 2: Update docs to unified policy wording**

Use these target phrases:

- `README.md`: Go `1.25+`, Node.js `20+`
- `docs/QUICK_START.md`: Go `1.25+`, Node.js `20+`
- `docs/DEPLOYMENT.md`: Go `1.25+` (and Node `20+` if file includes Node requirements)

Avoid introducing extra policy sources; docs should mirror config/CI.

- [ ] **Step 3: Verify docs are consistent**

Run:

```bash
rg -n "Go.*1\\.25\\+|Node\\.js.*20\\+" README.md docs/QUICK_START.md docs/DEPLOYMENT.md
rg -n "Node\\.js 18\\+|1\\.25\\.3" README.md docs/QUICK_START.md docs/DEPLOYMENT.md
```

Expected: first command shows expected new statements; second command returns no hits for old values.

- [ ] **Step 4: Commit Task 2**

Run:

```bash
git add README.md docs/QUICK_START.md docs/DEPLOYMENT.md
git commit -m "docs: unify go and node version requirements"
```

### Task 3: Add Version Drift Guard Script

**Files:**

- Create: `scripts/check-versions.sh`
- Test: direct script execution (positive + negative path)

- [ ] **Step 1: Create the guard script with explicit checks**

Create `scripts/check-versions.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

fail() {
  echo "VERSION CHECK FAILED: $1" >&2
  exit 1
}

require_line() {
  local file="$1"
  local pattern="$2"
  local desc="$3"
  if ! rg -n --pcre2 "$pattern" "$file" >/dev/null; then
    fail "$desc (file: $file, pattern: $pattern)"
  fi
}

require_absent() {
  local file="$1"
  local pattern="$2"
  local desc="$3"
  if rg -n --pcre2 "$pattern" "$file" >/dev/null; then
    fail "$desc (file: $file, pattern: $pattern)"
  fi
}

# Go policy
require_line "backend/go.mod" '^go 1\.25$' "backend/go.mod must declare go 1.25"
require_line "backend/go.mod" '^toolchain go1\.25\.3$' "backend/go.mod must declare toolchain go1.25.3"
require_line ".github/workflows/backend-ci.yml" 'go-version: "1\.25\.x"' "backend-ci must use go-version 1.25.x"
require_absent ".github/workflows/backend-ci.yml" 'go-version-file:' "backend-ci must not use go-version-file for this policy"

# Node policy in CI
require_line ".github/workflows/frontend-ci.yml" 'node-version: "20"' "frontend-ci must use node-version 20"
require_line ".github/workflows/codegen-check.yml" 'node-version: "20"' "codegen-check must use node-version 20"

# Doc mirrors
require_line "README.md" 'Go.*1\.25\+' "README must state Go 1.25+"
require_line "README.md" 'Node\.js.*20\+' "README must state Node.js 20+"
require_line "docs/QUICK_START.md" 'Go.*1\.25\+' "QUICK_START must state Go 1.25+"
require_line "docs/QUICK_START.md" 'Node\.js.*20\+' "QUICK_START must state Node.js 20+"
require_line "docs/DEPLOYMENT.md" 'Go.*1\.25\+' "DEPLOYMENT must state Go 1.25+"

echo "Version policy check passed."
```

- [ ] **Step 2: Make script executable and run positive test**

Run:

```bash
chmod +x scripts/check-versions.sh
./scripts/check-versions.sh
```

Expected: exits 0 and prints `Version policy check passed.`

- [ ] **Step 3: Run one negative-path smoke check**

Run:

```bash
cp README.md /tmp/README.md.bak
perl -0pi -e 's/Node\\.js 20\\+/Node.js 18+/g' README.md
if ./scripts/check-versions.sh; then echo "unexpected pass" && exit 1; else echo "negative-path check ok"; fi
mv /tmp/README.md.bak README.md
./scripts/check-versions.sh
```

Expected: middle run fails with `VERSION CHECK FAILED`; final run passes after restore.

- [ ] **Step 4: Commit Task 3**

Run:

```bash
git add scripts/check-versions.sh README.md
git commit -m "chore: add version consistency guard script"
```

### Task 4: Enforce Guard in CI

**Files:**

- Modify: `.github/workflows/codegen-check.yml`
- Test: workflow file static inspection + local script run

- [ ] **Step 1: Add CI step to run version guard**

Insert after dependency installation (before generation or before drift check):

```yaml
- name: Check runtime version policy
  run: ./scripts/check-versions.sh
```

- [ ] **Step 2: Verify workflow now includes enforcement**

Run:

```bash
rg -n "Check runtime version policy|./scripts/check-versions.sh" .github/workflows/codegen-check.yml
```

Expected: new step is present exactly once.

- [ ] **Step 3: Commit Task 4**

Run:

```bash
git add .github/workflows/codegen-check.yml
git commit -m "ci: enforce runtime version policy consistency"
```

### Task 5: Repository-Wide Verification and Final Hygiene

**Files:**

- Verify only; fix files if failures occur

- [ ] **Step 1: Run initialization check**

Run:

```bash
make init
```

Expected: successful initialization under Go 1.25+ and Node 20+.

- [ ] **Step 2: Run test suite check**

Run:

```bash
make test
```

Expected: backend and frontend tests pass.

- [ ] **Step 3: Run mandatory full check**

Run:

```bash
make check
```

Expected: lint + typecheck + tests + build pass (`✅ All checks passed`).

- [ ] **Step 4: If verification fails, debug before changing scope**

Use `@systematic-debugging` to isolate failing stage and apply minimal targeted fixes only.

- [ ] **Step 5: Commit any required verification-driven fixes**

Run (only if changes were needed):

```bash
git add <fixed-files>
git commit -m "fix: resolve verification issues for version policy unification"
```

## Final Review Checklist

- [ ] `backend/go.mod` has `go 1.25` + `toolchain go1.25.3`
- [ ] `backend/.golangci.yml` has `run.go: "1.25"`
- [ ] Backend CI uses `go-version: "1.25.x"`
- [ ] Frontend/codegen workflows use `node-version: "20"`
- [ ] README + QUICK_START + DEPLOYMENT reflect policy without stale values
- [ ] `scripts/check-versions.sh` passes locally
- [ ] `make check` passes
- [ ] Commits are scoped and readable

## Skills for Execution

- `@subagent-driven-development` (recommended for task-by-task execution)
- `@executing-plans` (inline alternative)
- `@verification-before-completion` before final completion claim
