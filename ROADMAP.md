# Roadmap & Status

Living tracker for the Phoenix experiment. Update it as work lands. For *why*
things are the way they are, see `docs/handbook.md`; for *what* the app must do,
see `spec/`.

Last updated: 2026-07-24.

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
| `evals/DSL.md` | revised | dogfood run 1 gap-log folded in (18/18) |
| `evals/HARNESS.md` | revised | § Settling added; readiness pinned; seed/read contract still to tighten |
| `evals/scenarios/` | stable | 8 files, 35 scenarios; full behavioral coverage |
| `evals/runner/` | built, unvalidated E2E | dogfood run 1 done; write-action semantics untested until an app exists |
| `docs/principles.md` | stable | distilled theory |
| `docs/handbook.md` | living | practice learnings |

## Backlog (roughly ordered)

### 1. Finish tightening the eval docs (dogfood run 1 fold-back, remaining threads)
The 18 gaps are folded in (see § Done). Still open:
- **Projection rework** — restate the § Page projection as simple, context-free rules,
  one local rule per line, with the produced HTML cooperating (declaring the markers the
  projection reads) so the projector never guesses.
- **Projection acceptance tests** — pin the DOM→projection mapping with fixed HTML
  fixtures and expected projection strings, so the projector is verified without an app.
- **Prose-style pass** — `handbook.md` and `spec.md` done; still to sweep
  `architecture.md`, `DSL.md`, `HARNESS.md`, `principles.md`, and the READMEs
  against the prose style guide in `CLAUDE.md`.
- **harness.json seed/read contract** — the manifest and the seed/read operations are
  underspecified. HARNESS.md pins only "a JSON array of `{id, title, completed}`,
  replace-all, exit 0 on success". Gaps to nail down, in HARNESS.md:
  - *Wire format.* Field types (`id` a JSON string? `completed` a JSON boolean, not
    `0/1`?), UTF-8, compact-vs-pretty, trailing-newline tolerance, whether unknown
    fields are ignored or rejected, and byte-for-byte round-trip of titles (internal
    whitespace, Unicode, quotes, newlines).
  - *Decision, reopened on purpose:* pin `id` as a JSON string on the wire? (#17b left
    it unpinned as a runner-decode detail; the seed/read wire is where both sides must
    agree, so revisit.)
  - *Meaning of `exit 0` from `seed`.* Must mean "fully committed and visible to the
    next request the server serves" — the durability handshake the runner relies on
    when it seeds then immediately navigates. Currently unstated.
  - *Error signaling.* Non-zero exit on failure, with a diagnostic on stderr.
  - *"Persisted order"* defined as the todos' display/insertion order, so seed→read
    round-trips order as well as contents.
  - *Concurrency assumption.* The runner never seeds/reads while the app is handling a
    request for that scenario — state it so adapters don't build locking they don't need.
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
- Dogfood run 1 gap-log (18 findings) resolved and folded into DSL.md / HARNESS.md /
  template. Notable: ids redefined as symbolic labels (§ Todo identity); § Settling
  eventual-consistency contract; projection reads declared class markers vs. rendered
  visibility by rule; always-on structural invariants made unconditional.
- Prose-style pass over `handbook.md` and `spec.md`; added a "no grand flourishes"
  rule to the style guide (state findings plainly, no closing epigrams).
