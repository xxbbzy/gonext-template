# Root README Rewrite Design

Date: 2026-03-28
Scope: Repository root `README.md`
Status: Approved for implementation planning

## 1. Background And Goal

The current root `README.md` is useful but still reads like a basic scaffold note.
It does not fully serve as the project front page for first-time users who need both:

- fast startup confidence ("can I run this now?")
- architectural confidence ("can this scale with disciplined conventions?")

Goal: rewrite the root README so a new user can understand the template positioning,
run the project locally, understand architectural guardrails, and know how to extend it.

## 2. Target Audience And Positioning

Primary audience:

- Solo developers using AI agents as a "team builder" workflow

Core value proposition (hero message):

- AI-native full-stack template for solo builders who work like a team

Language strategy:

- English-first README

## 3. In Scope / Out Of Scope

In scope:

- Rewrite root `README.md` structure and copy
- Add sections for positioning, scenarios, architecture conventions, local flow, Docker flow, OpenAPI generation, module extension, and roadmap
- Keep command examples directly runnable
- Link to deeper docs for detail (hybrid strategy)

Out of scope:

- Changing backend/frontend implementation
- Changing OpenAPI contract
- Editing `docs/*` except optional link verification if needed
- Changing Makefile targets

## 4. Approach Options And Decision

Three candidate approaches were considered:

1. Landing-first (positioning-heavy)
2. Workflow-first (runbook-heavy)
3. Contract-first (architecture-heavy)

Selected approach:

- Hybrid of landing-first + workflow-first
- Why: keeps first-read impact while preserving practical run commands

## 5. Information Architecture For New README

Planned top-level structure:

1. Hero / Positioning
2. Who This Is For (and "Not Ideal For")
3. Quick Start (10 minutes)
4. Tech Stack + Architecture Conventions
5. Local Development Flow
6. Docker Workflow
7. OpenAPI & Type Generation
8. How To Add A New Module
9. Roadmap
10. Documentation Map

## 6. Section-by-Section Design Requirements

### 6.1 Hero / Positioning

Must include:

- one-line value proposition centered on solo + AI-agent team workflow
- short "why this template" paragraph emphasizing guardrails and predictable scaling

### 6.2 Who This Is For

Must include:

- 3-4 concrete suitable scenarios
- one "not ideal for" statement to set boundary expectations

### 6.3 Quick Start

Must include:

- minimal command path: clone -> `make init` -> `make dev`
- explicit expected URLs and startup checks
- no excessive background explanation in this section

### 6.4 Tech Stack + Architecture Conventions

Must include:

- backend + frontend stack summary
- explicit architecture constraints:
  - OpenAPI contract-first (`api/openapi.yaml`)
  - handler -> service -> repository layering
  - response envelope via `backend/pkg/response`
  - generated files are outputs, not source of truth

### 6.5 Local Development Flow

Must include:

- practical daily loop: edit -> generate if needed -> verify
- mandatory validation command: `make check`
- runtime behavior change verification: `make e2e`

### 6.6 Docker Workflow

Must include:

- `make docker-build`, `make docker-up`, `make docker-down`
- service/port summary for frontend, backend, and postgres

### 6.7 OpenAPI & Type Generation

Must include:

- when to modify `api/openapi.yaml`
- when to run `make gen` and `make swagger`
- generated artifact targets:
  - `backend/internal/api/server.gen.go`
  - `frontend/types/api.ts`

### 6.8 How To Add A New Module

Must include:

- scaffold command `make new-module name=<module>`
- follow-up checklist aligned with architecture:
  - update OpenAPI first if API behavior changes
  - implement layer chain
  - wire providers/constructors
  - register routes
  - update migration/AutoMigrate considerations
  - run verification

### 6.9 Roadmap (Dual-Layer)

Must include:

- Near-term milestones (1-2 months)
- Longer-term direction themes (quarterly horizon)

Roadmap should remain directional, not a hard release commitment.

### 6.10 Documentation Map

Must include pointers to:

- `AGENTS.md`
- `ARCHITECTURE.md`
- `CONVENTIONS.md`
- `docs/README.md`

## 7. Content Style Rules

- English-first, concise, high-scanability
- Prefer checklists/tables over long prose when operational
- Keep examples executable and consistent with current Makefile targets
- Avoid duplicated deep documentation; link out for details

## 8. Acceptance Criteria Mapping

| Required outcome | README section that satisfies it |
| --- | --- |
| Project positioning | Hero / Positioning |
| Suitable scenarios | Who This Is For |
| Stack + architecture conventions | Tech Stack + Architecture Conventions |
| Local development process | Local Development Flow |
| Docker startup method | Docker Workflow |
| OpenAPI / type generation | OpenAPI & Type Generation |
| How to add module | How To Add A New Module |
| Roadmap | Roadmap |
| New user can understand and run | Quick Start + Flow + clear checks |

## 9. Risks And Mitigations

Risk 1: README becomes too long and loses first-read clarity.

- Mitigation: keep deep details linked, keep top-level sections crisp.

Risk 2: command drift from actual automation targets.

- Mitigation: source commands directly from `Makefile`.

Risk 3: architecture claims drift from current runtime wiring.

- Mitigation: align claims with `ARCHITECTURE.md` and `CONVENTIONS.md`.

## 10. Verification Plan For Rewrite Task

After rewriting `README.md`:

1. Verify all commands still exist in `Makefile`.
2. Verify all linked docs paths resolve.
3. Run mandatory validation pipeline:
   - `make check`
4. Optional but recommended for runtime-facing wording confidence:
   - `make e2e`

Expected:

- README is self-sufficient for onboarding and startup
- no contradiction with root architecture/convention docs
- repository checks pass

