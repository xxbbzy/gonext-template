# Template Residue Cleanup and Documentation Consistency Design

Date: 2026-03-29
Scope: `README.md`, `frontend/README.md`, `docs/QUICK_START.md`, `docs/DEPLOYMENT.md`, and documentation guardrails
Status: Approved for implementation planning

## 1. Background And Goal

The repository currently has several "template residue and inconsistency" signals:

- `frontend/README.md` still contains default `create-next-app` template content.
- Root `README.md` claims Wire is optional/manual by default, while backend runtime actually uses Wire-generated DI initialization.
- Documentation text around local dev, Docker commands, and runtime facts may drift from executable sources (`Makefile`, `.air.toml`, `docker-compose.yml`).

Goal:

1. Remove default template residue.
2. Make README/docs claims match current code and executable commands.
3. Add a lightweight anti-drift mechanism so this inconsistency is less likely to return.

## 2. Scope

In scope:

- Rewrite `frontend/README.md` in Chinese with project-specific, executable guidance.
- Correct root `README.md` statements to match actual implementation and commands.
- Align `docs/QUICK_START.md` and `docs/DEPLOYMENT.md` with current `Makefile` and `docker-compose.yml`.
- Fix stale statements discovered during alignment if they are factually outdated.
- Add documentation consistency guardrails to contributor workflow docs.

Out of scope:

- Backend/frontend business logic changes
- API contract changes in `api/openapi.yaml`
- Refactoring build or deployment pipeline behavior itself

## 3. Options Considered

### Option A: Minimal Document Patch

- Touch only explicitly listed mismatches.
- Pros: fast, low change surface.
- Cons: high chance of future drift; no process guard.

### Option B: Single Source of Truth Alignment

- Align all target docs with executable sources.
- Pros: reliable consistency for current state.
- Cons: still lacks explicit anti-drift workflow rule.

### Option C (Selected): Alignment + Anti-Drift Guardrail

- Do Option B plus explicit doc synchronization rules in contributor docs.
- Pros: solves current mismatch and reduces recurrence risk.
- Cons: slightly broader docs change set.

## 4. Design Overview

This design establishes "documentation follows executable truth":

1. Command truth source: `Makefile`
2. Backend hot-reload truth source: `backend/.air.toml` + `make dev-backend`
3. Container topology truth source: `docker-compose.yml`
4. DI truth source: `backend/cmd/server/main.go`, `wire.go`, `wire_gen.go`

All modified docs should reference these facts consistently and avoid speculative wording.

## 5. Component-Level Changes

### 5.1 `frontend/README.md` (full rewrite, Chinese)

Replace default Next.js scaffold text with project-specific content:

- Project purpose and relation to root workflow
- Start commands (prefer `make dev-frontend`; include direct `npm` fallback)
- API base URL and local integration notes (`NEXT_PUBLIC_API_URL`)
- Script reference (`dev`, `build`, `lint`, `typecheck`, `test`)
- Key frontend directories (`app/`, `lib/`, `stores/`, `types/`)

Acceptance intent: no default `create-next-app` residue remains.

### 5.2 Root `README.md`

Align to implementation facts:

- Replace incorrect Wire statement with "Wire is in active use for compile-time DI".
- Ensure command examples map to current `Makefile` targets.
- Ensure Docker and runtime claims match actual compose behavior.
- Keep concise and executable orientation.

### 5.3 `docs/QUICK_START.md`

Align quick-start behavior with current targets and current effects:

- `make dev` decomposition
- `make check` description includes lint + typecheck + test + build
- command snippets and comments use existing make targets only

### 5.4 `docs/DEPLOYMENT.md`

Align deployment/runtime facts with current compose and repository state:

- backend/frontend/postgres service and port mapping
- `make docker-up` / `make docker-down` behavior
- remove or correct stale version-drift claims if outdated

### 5.5 Anti-Drift Rule Section (in `docs/DEVELOPMENT.md`)

Add a concise "Documentation Consistency Checklist" section:

- When `Makefile` commands change, sync `README.md`, `docs/QUICK_START.md`, `docs/DEPLOYMENT.md`.
- When `docker-compose.yml` changes, sync deployment-related docs.
- When DI strategy changes, sync README + architecture/ADR references.
- Do not commit scaffold-default README content in module directories.

## 6. Data Flow

1. Contributor updates runtime/command/container definitions.
2. Contributor applies required doc synchronization checklist.
3. Reviewers validate docs against executable truth sources.
4. Repository documentation remains aligned with actual behavior.

## 7. Error Handling Strategy

This is a docs-only change, so "error handling" means process-level conflict handling:

- If a doc statement cannot be traced to executable source, remove or qualify it.
- If two docs conflict, prefer executable source (`Makefile`, compose, runtime entry) then propagate correction.
- If historical notes are uncertain, avoid speculative claims and keep factual present-state wording.

## 8. Testing And Verification Plan

Document consistency checks:

1. Search for default scaffold residue in frontend README (for example `create-next-app` text).
2. Search for contradictory Wire wording (for example "manual DI by default").
3. Cross-check README/docs command names against `Makefile`.
4. Cross-check deployment/port claims against `docker-compose.yml`.

Mandatory repository verification after doc edits:

5. Run `make check` and require exit code `0`.

## 9. Risks And Mitigations

Risk: Updated docs become too verbose and lose quick-start utility.
Mitigation: keep root README concise; move detailed explanation to `docs/`.

Risk: Partial synchronization leaves hidden contradictions.
Mitigation: use explicit one-pass cross-check among README, QUICK_START, DEPLOYMENT, Makefile, compose.

Risk: Future contributors reintroduce template residue.
Mitigation: anti-drift checklist in `docs/DEVELOPMENT.md`.

## 10. Acceptance Criteria Mapping

| Requested acceptance criterion           | Design coverage                                                                |
| ---------------------------------------- | ------------------------------------------------------------------------------ |
| No default Next.js README residue        | `frontend/README.md` full rewrite with project-specific Chinese content        |
| README claims map to real implementation | Wire/runtime/commands tied to executable truth sources and cross-check process |

## 11. Implementation Checklist

- [ ] Rewrite `frontend/README.md` in Chinese, remove scaffold defaults
- [ ] Update root `README.md` Wire/dev/docker wording to factual current state
- [ ] Align `docs/QUICK_START.md` with `Makefile` actual behavior
- [ ] Align `docs/DEPLOYMENT.md` with `docker-compose.yml` and current runtime facts
- [ ] Add documentation consistency checklist to `docs/DEVELOPMENT.md`
- [ ] Run `make check`
