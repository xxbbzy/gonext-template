# Root ARCHITECTURE.md Rewrite Design

Date: 2026-03-29
Scope: `ARCHITECTURE.md`
Status: Approved for implementation planning

## 1. Background And Goal

The repository already has a root `ARCHITECTURE.md`, but it currently behaves more
like a runtime topology summary than an execution-oriented architecture guide.
Important conventions still need to be inferred from code or cross-read from
`AGENTS.md` and `CONVENTIONS.md`.

Goal:

1. Rewrite the root `ARCHITECTURE.md` as an AI-agent-facing architecture map.
2. Make backend layer responsibilities and dependency direction explicit.
3. Document where extension work usually happens when adding a new module.
4. Cover frontend and runtime integration points enough for cross-stack changes
   without turning the file into a contributor handbook.

## 2. Scope

In scope:

- Rewrite the structure and content of root `ARCHITECTURE.md`.
- Document responsibilities for `handler`, `service`, `repository`, `model`, and `dto`.
- Document allowed dependency direction and architectural boundaries.
- Document frontend/runtime integration points that influence architecture work.
- Add a recommended "new module" checklist mapped to real repository paths.

Out of scope:

- Business logic changes
- API contract changes in `api/openapi.yaml`
- Changes to `AGENTS.md`, `CONVENTIONS.md`, or bilingual docs unless required by a
  factual mismatch discovered during the rewrite
- ADR changes

## 3. Options Considered

### Option A: Runtime-Only Summary Refresh

- Keep the current structure and only clarify topology/middleware/DI wording.
- Pros: smallest change surface.
- Cons: does not solve the main problem that layer boundaries and extension rules
  still live mostly in code and scattered docs.

### Option B (Selected): Layer-Centered Execution Map

- Organize the document around system topology, backend layers, dependency rules,
  integration points, and a module-addition checklist.
- Pros: directly supports AI-agent execution and architectural consistency.
- Cons: denser than the current document and requires careful editing to stay
  compact.

### Option C: Task-Playbook-First Architecture Doc

- Structure the file around "how to add API/module/frontend page/middleware".
- Pros: very action-oriented.
- Cons: duplicates `AGENTS.md` too heavily and weakens the document as a stable
  architecture reference.

## 4. Design Overview

The rewritten `ARCHITECTURE.md` should answer four operational questions in a
single pass:

1. How the system is wired today
2. What each backend layer owns
3. Which dependencies are allowed and which are boundary violations
4. Which files are typically touched when adding a module

The document remains English-only and compact in style, but "compact" should mean
"high-density and scannable", not "missing architectural rules". The rewrite
should prefer short sections, explicit path references, and rule-style bullets
over narrative prose.

## 5. Proposed Document Structure

### 5.1 Purpose And Scope

Open with a short explanation that the file is the AI-facing architecture map for:

- runtime topology
- layer boundaries
- dependency direction
- extension points

It should also redirect readers to:

- `AGENTS.md` for task playbooks
- `CONVENTIONS.md` for coding/testing/error-handling rules
- `docs/adr/` for architectural rationale

### 5.2 System Topology

Provide a concise end-to-end topology section that connects:

- `api/openapi.yaml`
- `frontend/app/`
- `frontend/lib/api-client.ts` and `frontend/lib/api-client.gen.ts`
- `frontend/lib/query-provider.tsx`
- `frontend/stores/auth.ts`
- `backend/cmd/server/main.go`
- middleware
- `handler -> service -> repository -> model`
- database initialization in `backend/internal/config/database.go`

This section should orient agents before they jump into a layer-specific change.

### 5.3 Backend Layer Responsibilities

Make this the core of the document. Each layer section should use the same shape:

- what it owns
- what it may depend on
- what it must not do
- where it lives
- what usually changes when adding a module

Required layers:

- `handler`
- `service`
- `repository`
- `model`
- `dto`

This section should explicitly distinguish transport shape (`dto`) from
persistence shape (`model`), and business orchestration (`service`) from storage
concerns (`repository`).

### 5.4 Dependency Direction And Boundaries

Add a rule-oriented section that states the intended dependency flow directly,
instead of leaving it implicit:

- `handler -> service -> repository -> model`
- DTOs stay on the transport side
- middleware owns cross-cutting HTTP concerns
- `pkg/response` and `pkg/errcode` are shared support packages, not replacement
  layers
- Wire files assemble dependencies but do not own business behavior

This section should also mention common anti-patterns to avoid:

- handlers querying GORM directly
- services returning ad hoc HTTP payload semantics
- repositories taking ownership of business policy
- models being used as API response structs by default

### 5.5 Frontend And Runtime Integration Points

Keep this section narrower than the backend section, but include the architecture
anchors that matter for real changes:

- App Router entrypoints under `frontend/app/`
- generated and handwritten API client boundary
- auth persistence in `frontend/stores/auth.ts`
- TanStack Query bootstrap in `frontend/lib/query-provider.tsx`
- DI assembly in `backend/cmd/server/wire.go`, `providers.go`, `wire_gen.go`
- middleware mounting and route registration in `backend/cmd/server/main.go`
- database bootstrap in `backend/internal/config/database.go`

The purpose is to show where cross-stack and runtime wiring work starts, not to
repeat implementation details already covered elsewhere.

### 5.6 Recommended Flow For Adding A New Module

Write this as a strict-looking checklist mapped to real repository paths:

1. Update `api/openapi.yaml` first when API behavior changes.
2. Add or update DTOs in `backend/internal/dto/`.
3. Implement `handler`, `service`, `repository`, and `model` changes in their
   layer packages.
4. Register providers and constructors in `backend/cmd/server/providers.go` and
   `backend/cmd/server/wire.go`.
5. Regenerate `wire_gen.go` via the existing Wire workflow.
6. Register routes in `backend/cmd/server/main.go`.
7. If persistence changes, update development `AutoMigrate` coverage and add SQL
   migrations under `backend/migrations/`.
8. Run `make gen` when the contract changed.
9. Run `make check`.
10. Run `make e2e` for API/runtime behavior changes.

This section should read like a repository-specific architectural workflow, not a
generic framework tutorial.

### 5.7 Documentation And Decision Links

Close with a short reference list to:

- `AGENTS.md`
- `CONVENTIONS.md`
- `docs/README.md`
- relevant ADRs

This keeps the main sections focused while still giving the reader next hops.

## 6. Content Rules For The Rewrite

The final `ARCHITECTURE.md` should follow these writing constraints:

- English-only
- compact but detailed enough for execution
- path-specific, not abstract
- no duplication of full task playbooks from `AGENTS.md`
- no duplication of coding-rule details from `CONVENTIONS.md`
- explicit about boundaries and extension points

Where current repository docs disagree on implementation details, the rewrite
should align to actual code and stable repository guidance rather than copying the
older wording.

## 7. Risks And Mitigations

Risk: The rewritten file becomes too long and starts overlapping heavily with
`AGENTS.md` or `CONVENTIONS.md`.
Mitigation: keep the architecture file focused on structure, dependency flow, and
entry points; link out for process and coding detail.

Risk: The file stays too high-level and fails to help agents make edits.
Mitigation: every layer section includes concrete paths and "usually changed
files" guidance.

Risk: Frontend/runtime integration remains underspecified for cross-stack changes.
Mitigation: include a dedicated integration-points section that connects contract,
frontend client, auth/query providers, main runtime, Wire, middleware, and DB
bootstrap.

## 8. Acceptance Criteria Mapping

| Requested acceptance criterion                      | Design coverage                                                                 |
| --------------------------------------------------- | ------------------------------------------------------------------------------- |
| Project conventions no longer live only in code     | Explicit layer ownership, dependency rules, and extension checklist in root doc |
| Explain handler/service/repository/model/dto roles  | Dedicated backend layer responsibilities section                                |
| Explain dependency direction and boundaries         | Dedicated dependency direction and anti-pattern section                         |
| Explain recommended process for adding a new module | Repository-specific checklist section mapped to concrete files                  |

## 9. Implementation Checklist

- [ ] Rewrite root `ARCHITECTURE.md` around architecture rules rather than only runtime summary
- [ ] Add system topology section covering frontend, contract, backend runtime, and DB
- [ ] Add per-layer sections for `handler`, `service`, `repository`, `model`, and `dto`
- [ ] Add explicit dependency direction and boundary rules
- [ ] Add frontend/runtime integration points section
- [ ] Add concrete "new module" checklist mapped to repository paths
- [ ] Keep the document English-only and compact in style
- [ ] Run `make check`
