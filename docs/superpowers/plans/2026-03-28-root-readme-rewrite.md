# Root README Rewrite Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite the root `README.md` so first-time users can understand project positioning, run the stack locally, and extend modules with confidence.

**Architecture:** Keep `README.md` as the onboarding front page while preserving repository conventions from root AI docs. Use a hybrid structure: clear product positioning first, then runnable local and Docker workflows, then extension and contract-generation guidance. Keep deep explanations linked to existing `docs/` references instead of duplicating them.

**Tech Stack:** Markdown docs, Makefile commands, Docker Compose, OpenAPI contract (`api/openapi.yaml`), Go/Gin/GORM/Wire backend, Next.js frontend.

---

## File Structure And Responsibilities

- Modify: `README.md` — primary rewrite target with new information architecture and runnable onboarding content.
- Modify: `.gitignore` — remove `docs/superpowers` ignore rule so spec/plan docs can be versioned.
- Reference-only: `ARCHITECTURE.md`, `CONVENTIONS.md`, `AGENTS.md`, `docs/README.md`, `Makefile`, `docker-compose.yml` — sources for accurate commands and conventions.

## Task 1: Enable Versioning For Superpowers Docs

**Files:**

- Modify: `.gitignore`
- Verify: `.gitignore`

- [ ] **Step 1: Remove ignored-path rule for superpowers docs**

Edit `.gitignore` and delete:

```text
docs/superpowers
```

- [ ] **Step 2: Verify the ignore rule is gone**

Run: `rg -n '^docs/superpowers$' .gitignore || true`  
Expected: no output

- [ ] **Step 3: Commit the `.gitignore` change**

```bash
git add .gitignore
git commit -m "chore: track superpowers docs in git"
```

## Task 2: Rewrite README Front Matter (Positioning + Audience + Scenarios)

**Files:**

- Modify: `README.md`
- Reference: `ARCHITECTURE.md`, `CONVENTIONS.md`

- [ ] **Step 1: Replace top section with new hero and value proposition**

Insert a concise opening block that states:

```md
AI-native full-stack template for solo builders who work like a team.
```

and explains contract-first + architectural guardrails.

- [ ] **Step 2: Add "Who This Is For" section with concrete scenarios**

Add 3-4 scenario bullets (MVP, internal tools, API-first products, AI-agent-assisted delivery) and one "Not ideal for" boundary line.

- [ ] **Step 3: Validate required headings exist**

Run: `rg -n '^## (Who This Is For|Quick Start|Tech Stack & Architecture Conventions)$' README.md`  
Expected: all three headings are found

- [ ] **Step 4: Commit front-matter rewrite**

```bash
git add README.md
git commit -m "docs: rewrite README positioning and audience sections"
```

## Task 3: Add Runnable Onboarding Flow (Quick Start + Local Dev + Docker)

**Files:**

- Modify: `README.md`
- Reference: `Makefile`, `docker-compose.yml`

- [ ] **Step 1: Add "Quick Start (10 minutes)" with copy-paste commands**

Use commands aligned with Make targets:

```bash
git clone <your-repo-url>
cd gonext-template
make init
make dev
```

Include expected endpoints for success checks (`http://localhost:3000`, `http://localhost:8080`).

- [ ] **Step 2: Add "Local Development Flow" section**

Include daily loop: implement -> regenerate when needed -> `make check` -> `make e2e` for runtime behavior changes.

- [ ] **Step 3: Add "Docker Workflow" section**

Use:

```bash
make docker-build
make docker-up
make docker-down
```

and include service/port mapping for frontend, backend, postgres.

- [ ] **Step 4: Verify commands referenced in README exist in Makefile**

Run:

```bash
for c in init dev check e2e docker-build docker-up docker-down gen swagger new-module; do
  rg -n "^$c:" Makefile >/dev/null || echo "missing: $c"
done
```

Expected: no `missing:` output

- [ ] **Step 5: Commit onboarding workflow sections**

```bash
git add README.md
git commit -m "docs: add README quick start, local flow, and docker workflow"
```

## Task 4: Add Contract/Extension Guidance And Roadmap

**Files:**

- Modify: `README.md`
- Reference: `api/openapi.yaml`, `scripts/new-module.sh`, `AGENTS.md`, `ARCHITECTURE.md`, `CONVENTIONS.md`

- [ ] **Step 1: Add "OpenAPI & Type Generation" section**

Cover:

- contract source of truth: `api/openapi.yaml`
- generation commands: `make gen`, `make swagger`
- generated artifacts: `backend/internal/api/server.gen.go`, `frontend/types/api.ts`

- [ ] **Step 2: Add "How To Add A New Module" section**

Include scaffold command and checklist:

```bash
make new-module name=product
```

Checklist should include layer chain, Wire/providers updates, route registration, migration considerations, and verification commands.

- [ ] **Step 3: Add dual-layer roadmap**

Add:

- Near-term milestones (1-2 months)
- Longer-term direction themes (quarterly)

- [ ] **Step 4: Add "Documentation Map" links**

Link root docs and curated docs index:

- `AGENTS.md`
- `ARCHITECTURE.md`
- `CONVENTIONS.md`
- `docs/README.md`

- [ ] **Step 5: Commit architecture/extension/roadmap sections**

```bash
git add README.md
git commit -m "docs: add README contract, extension guide, and roadmap"
```

## Task 5: Verify Consistency And Quality Gate

**Files:**

- Verify: `README.md`, `.gitignore`

- [ ] **Step 1: Validate README includes all required top-level sections**

Run:

```bash
rg -n '^## ' README.md
```

Expected headings include:

- `Who This Is For`
- `Quick Start`
- `Tech Stack & Architecture Conventions`
- `Local Development Flow`
- `Docker Workflow`
- `OpenAPI & Type Generation`
- `How To Add A New Module`
- `Roadmap`
- `Documentation Map`

- [ ] **Step 2: Verify links and referenced files resolve**

Run:

```bash
for f in AGENTS.md ARCHITECTURE.md CONVENTIONS.md docs/README.md api/openapi.yaml backend/internal/api/server.gen.go frontend/types/api.ts; do
  test -e "$f" || echo "missing: $f"
done
```

Expected: no `missing:` output

- [ ] **Step 3: Run mandatory repository validation**

Run: `make check`  
Expected: exit code `0`

- [ ] **Step 4: (Recommended) run runtime smoke verification**

Run: `make e2e`  
Expected: register -> login -> CRUD smoke flow passes

- [ ] **Step 5: Final commit for any verification-related adjustments**

```bash
git add README.md .gitignore
git commit -m "docs: finalize root README rewrite and verification updates"
```

## Notes For Executors

- Keep README English-first and concise.
- Do not introduce new commands not present in `Makefile`.
- Do not drift from architecture constraints documented in root docs.
- Before claiming completion, use @verification-before-completion.
