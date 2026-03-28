# Unified Runtime Version Policy Design

- Date: 2026-03-28
- Topic: Unify Go / Node / CI / README version policy
- Status: Draft for review

## 1. Background

Current version requirements are drifting across runtime config, CI, and docs:

- Go has conflicting expressions (`1.25.0` in `backend/go.mod`, `1.25.3` in docs).
- Node minimum version differs between docs (`18+`) and CI (`20`).
- There is no automated guard to block future drift.

This causes onboarding confusion and "works locally but fails in CI" risk.

## 2. Goals

1. Define one consistent runtime policy for local, CI, and documentation.
2. Align Go, Node, lint, and CI declarations to the same policy.
3. Add an automated consistency check to prevent future drift.
4. Keep changes minimal and low-risk (no framework/toolchain migration).

## 3. Non-Goals

1. Replacing existing build/test/lint workflow structure.
2. Introducing new version managers (`.nvmrc`, Volta, asdf) in this change.
3. Upgrading framework major versions unrelated to the policy issue.

## 4. Policy Decision

## Runtime Policy

1. Go policy: `1.25+`
2. Node policy: `20+`

## Canonical Sources

1. Go baseline and recommended toolchain:
   - `backend/go.mod` sets `go 1.25`
   - `backend/go.mod` adds `toolchain go1.25.3` (recommended patch toolchain)
2. Node minimum:
   - Documentation states `Node.js 20+`
   - CI runs Node 20 (`actions/setup-node`)

## 5. Design Approach (Selected)

Selected approach: "Consistency + enforcement guard"

1. Align all version declarations (runtime config, CI, docs).
2. Add a lightweight version consistency script and run it in CI.
3. Fail fast on mismatch.

Rationale:

- Solves current drift immediately.
- Prevents recurrence with low implementation complexity.

## 6. Planned Changes

## 6.1 Runtime and Lint

1. `backend/go.mod`
   - Normalize to `go 1.25`
   - Add `toolchain go1.25.3`
2. `backend/.golangci.yml`
   - Add explicit Go target version (`run.go: "1.25"`) to align lint behavior.

## 6.2 CI

1. `.github/workflows/backend-ci.yml`
   - Change `actions/setup-go` from `go-version-file` to explicit `go-version: "1.25.x"`
   - Keep dependency cache behavior unchanged.
2. Keep frontend/codegen Node CI baseline at `20`.

## 6.3 Documentation

1. `README.md`
   - Environment section: Go `1.25+`, Node `20+`
2. `docs/QUICK_START.md`
   - Align Go/Node requirement wording to same policy.
3. `docs/DEPLOYMENT.md`
   - Align Go version wording to same policy.

## 6.4 Consistency Guard

1. Add script: `scripts/check-versions.sh`
2. Script checks:
   - `backend/go.mod` contains `go 1.25`
   - `backend/go.mod` contains `toolchain go1.25.3` (current recommended patch)
   - `backend-ci.yml` contains `go-version: "1.25.x"`
   - `README.md`, `docs/QUICK_START.md`, `docs/DEPLOYMENT.md` all state Go `1.25+` and Node `20+`
3. Integrate script into CI (prefer `codegen-check.yml` as an extra step/job).

## 7. Data Flow and Failure Behavior

## Data Flow

1. Developer edits runtime/CI/docs.
2. Version check script runs in CI.
3. If mismatch detected, CI fails with targeted error.
4. PR cannot merge until policy is consistent.

## Failure Behavior

1. Script should print exact file and expected value.
2. Non-zero exit on first mismatch (or aggregated mismatches if implemented).
3. Messaging should guide contributor to the canonical source.

## 8. Testing and Verification Plan

Mandatory verification after implementation:

1. `make init`
2. `make test`
3. `make check` (required by repository policy)

CI verification:

1. Backend CI passes with Go `1.25.x`.
2. Frontend/codegen CI passes with Node `20`.
3. Version consistency guard passes on aligned state and fails on intentional mismatch.

## 9. Risks and Mitigations

1. Risk: Contributors on Node 18 fail local setup.
   - Mitigation: clear doc requirements + CI enforcement.
2. Risk: Go patch recommendation drifts over time.
   - Mitigation: update `toolchain` and docs together; script catches partial updates.
3. Risk: guard script becomes too strict.
   - Mitigation: keep checks scoped to policy-critical fields only.

## 10. Acceptance Criteria Mapping

1. Local, CI, and README version requirements are consistent.
2. `make init` / `make test` / CI run successfully under unified policy.
3. Drift prevention exists and blocks inconsistent changes.

## 11. Implementation Notes for Next Phase

This spec intentionally limits scope to version-policy unification and enforcement.
No additional refactors are required for this issue.
