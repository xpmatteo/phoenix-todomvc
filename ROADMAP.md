# Roadmap & Status

Living tracker for the Phoenix experiment. Update it as work lands. For *why*
things are the way they are, see `docs/handbook.md`; for *what* the app must do,
see `spec/`.

Last updated: 2026-07-24.

## Where we are

The durable artifacts exist and the eval runner must still be built.
**No implementation (`app/`) exists yet**

## Status by artifact

| Artifact | State | Notes |
|---|---|---|
| `spec/spec.md` | stable | behavioral spec; purely observable behavior |
| `spec/architecture.md` | stable | 5 NFRs, 7 ADs, each with named enforcement |
| `spec/main-screen-template.{html,css}` | stable | UI contract |
| `schema/` | stable | initial migration only (`001_create_todos.sql`) |
| `evals/DSL.md` | revised | 18/18 gaps folded in; § Page projection reworked to marker-only, context-free rules |
| `evals/HARNESS.md` | revised | § Settling added; readiness pinned; seed/read contract still to tighten |
| `evals/scenarios/` | stable | 8 files, 35 scenarios; full behavioral coverage |
| `evals/projection/` | stable | 19 txtar cases; pin the DOM→projection mapping so a rebuilt projector is verified without an app |
| `evals/runner/` | built, now stale | predates the projection rework (reads quoting/strike/editing the old way); regenerate against current DSL |
| `docs/principles.md` | stable | distilled theory |
| `docs/handbook.md` | living | practice learnings |

## Backlog (roughly ordered)

### 1. Finish tightening the eval docs

- **harness.json seed/read contract** — the manifest and the seed/read operations are
  underspecified. HARNESS.md pins only "a JSON array of `{id, title, completed}`,
  replace-all, exit 0 on success". Gaps to nail down, in HARNESS.md:
  - *Wire format.* Field types (`id` a JSON string? `completed` a JSON boolean, not
    `0/1`?), UTF-8, compact-vs-pretty, trailing-newline tolerance, whether unknown
    fields are ignored or rejected, and byte-for-byte round-trip of titles (internal
    whitespace, Unicode, quotes, newlines).
  - *Meaning of `exit 0` from `seed`.* Must mean "fully committed and visible to the
    next request the server serves" — the durability handshake the runner relies on
    when it seeds then immediately navigates. Currently unstated.
  - *Error signaling.* Non-zero exit on failure, with a diagnostic on stderr.
  - *Concurrency assumption.* The runner never seeds/reads while the app is handling a
    request for that scenario — state it so adapters don't build locking they don't need.
- Then: generate the test runner, logging any significant decision 

### 2. Architecture tests (roadmap in `spec/architecture.md`)

- `go.mod` dependency allowlist (AD-3)
- import-graph check: model imports no technology (AD-4)
- no web path reaches the side-channel code (AD-5)
- static conformance: app README exists; served HTML vs. template

### 3. First regeneration

- Write the regeneration skill/prompt (inputs: `spec/`, `schema/`, and the eval +
  architecture-test suites as the acceptance gate).
- Generate `app/`; run evals to green.
- **The real test:** delete `app/` and regenerate. Success is boredom.

