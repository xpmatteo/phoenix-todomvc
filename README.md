# TodoMVC — a Phoenix Architecture experiment

This repository explores regenerative ("phoenix") software: the specification and
the evaluations are the durable artifacts; the implementation is disposable and is
regenerated from them when they change. Background: `docs/principles.md`, distilled 
from Chad Fowler's [The Phoenix Architecture blog](https://aicoding.leaflet.pub/).

## Layout: durable vs disposable

| Path | Status | What it is |
|---|---|---|
| `docs/spec.md` | durable | behavioral specification |
| `docs/architecture.md` | durable | non-functional requirements and the architecture decisions derived from them |
| `docs/main-screen-template.html/.css` | durable | the UI contract: DOM vocabulary and styling |
| `docs/principles.md` | durable | distilled Phoenix Architecture principles |
| `evals/DSL.md` | durable | the scenario language definition |
| `evals/HARNESS.md` | durable | the runner ↔ implementation contract |
| `evals/scenarios/` | durable | the executable specification-by-example |
| `schema/` | durable, **append-only** | the database schema's migration chain — provenance for data, which outlives every implementation |
| `evals/runner/` | rebuildable | test runner; rebuildable from DSL.md + HARNESS.md + scenarios + templates alone |
| `app/` | **disposable** | the generated implementation; safe to delete at any time |

Nothing in `app/` may ever be load-bearing: if regenerating it from the durable
artifacts fails or needs hand-editing, that is a defect in the durable artifacts —
fix it there.

## The boundary rule

Every requirement lives in exactly one home, chosen by one test:

> **Could a black-box scenario fail because of this?**

- **Yes** → it belongs in `docs/spec.md` (and should be pinned by a scenario in
  `evals/scenarios/`). The spec contains only behavior observable through the eval
  surface: rendered UI, persisted model, URLs.
- **No** → it belongs in `docs/architecture.md`, as a decision traced to a
  non-functional requirement, with its own enforcement (architecture tests, static
  checks, or review) — the behavioral evals are deliberately blind to it.
- Anything that is *interface between the evals and the app* (how to start, seed,
  read an implementation) is canonical in `evals/HARNESS.md` and only referenced
  elsewhere.

## Rules of the game

- The eval suite must survive reimplementation: scenarios never mention
  implementation technology. If a technology change invalidates a scenario, the
  scenario was specified at the wrong boundary.
- A bug found in a generated app is first encoded as a scenario, then fixed by
  regeneration — the lesson must survive the next fire.
- Never edit scenarios to make a failing implementation pass; decide whether the
  spec or the implementation is wrong, and fix that.
