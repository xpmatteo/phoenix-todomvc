# Roadmap & Status

Living tracker for the Phoenix experiment. Update it as work lands. For *why*
things are the way they are, see `docs/handbook.md`; for *what* the app must do,
see `spec/`.

Last updated: 2026-07-23.

## Where we are

The durable artifacts exist and the eval runner has been built once from them
alone. **No implementation (`app/`) exists yet** — the first regeneration is the
next milestone.

## Status by artifact

| Artifact | State | Notes |
|---|---|---|
| `spec/spec.md` | stable | behavioral spec; purely observable behavior |
| `spec/architecture.md` | stable | 5 NFRs, 7 ADs, each with named enforcement |
| `spec/main-screen-template.{html,css}` | stable | UI contract |
| `schema/` | stable | initial migration only (`001_create_todos.sql`) |
| `evals/DSL.md` | needs revision | fold in gap-log decisions (see below) |
| `evals/HARNESS.md` | needs revision | add a timing model; resolve id-assignment |
| `evals/scenarios/` | stable | 8 files, 35 scenarios; full behavioral coverage |
| `evals/runner/` | built, unvalidated E2E | dogfood run 1 done; write-action semantics untested until an app exists |
| `docs/principles.md` | stable | distilled theory |
| `docs/handbook.md` | living | practice learnings |

## Backlog (roughly ordered)

### 1. Close dogfood run 1 — fold `evals/runner/QUESTIONS.md` back into the docs
18 gaps found. Decisions still needed from Matt:
- **#1 id assignment** — DSL says the adapter assigns omitted ids; HARNESS's seed
  takes them verbatim. Pick one (leaning: "the runner assigns"). *contradiction*
- **#4 timing model** — neither doc says when a THEN may be evaluated; runner
  polls to a timeout. Promote to contract text. *systemic*
- **#6 "completed styling"** — rendered strikethrough vs. the `completed` class;
  runner chose rendered truth. Confirm.
- **#15 isolation** — fresh Chrome per scenario (~0.5s × 35). Confirm or relax.
- The other ~14 are low-controversy; write them into DSL/HARNESS as-is.
- Then: dogfood run 2 should yield a near-empty gap log (the quality metric).

### 2. Architecture tests (roadmap in `spec/architecture.md`)
- eval suite with JavaScript disabled (enforces AD-1 server-rendering)
- `go.mod` dependency allowlist (AD-3)
- import-graph check: model imports no technology (AD-4)
- no web path reaches the side-channel code (AD-5)
- static conformance: app README exists; served HTML vs. template

### 3. First regeneration
- Write the regeneration skill/prompt (inputs: `spec/`, `schema/`, and the eval +
  architecture-test suites as the acceptance gate).
- Generate `app/`; run evals to green.
- **The real test:** delete `app/` and regenerate. Success is boredom.

## Open decisions (not blocking, but pending)
- Rename `spec/spec.md` → `spec/behavior.md`? (reads well beside `architecture.md`)
- Storage-key enforcement resolved via AD-7 (schema is durable) — no action.

## Done
- Distilled the Phoenix blog → `docs/principles.md`.
- Spec-by-example DSL + 35 scenarios; model = persisted todos (survived SPA→SSR).
- id contract pinned (`data-id`, opaque, stable).
- Architecture split from spec; NFR-driven ADs; boundary rule in README.
- goose migrations; schema as a durable, append-only asset.
- Phoenix handbook started.
- Eval runner built from the durable docs alone (dogfood run 1).
- `spec/` (normative) split from `docs/` (commentary).
