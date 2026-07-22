# Eval Harness Contract

Status: **durable artifact**. Defines the boundary between the test runner and each
generated implementation of the app. `DSL.md` defines *what* the scenarios mean;
this document defines *how* a runner gets any implementation running, seeded, and
read back. A fresh agent must be able to build the runner from this document plus
`DSL.md`, the scenario files, and the two template files — nothing else.

## Roles

- **Runner** (`evals/runner/`) — a Go program. Parses the scenario files per
  `DSL.md`, drives a real browser over the Chrome DevTools Protocol, and uses the
  adapter described below. It contains no knowledge of any particular
  implementation and is never modified when the app is regenerated. It is
  disposable in principle (rebuildable from these documents) but stable in
  practice.

- **Adapter** — shipped *by each implementation* inside `app/`, regenerated along
  with it: the `harness.json` manifest and the executables it names. The only
  place that knows how the app is built, served, and persisted.

The DOM vocabulary the runner needs (element classes, structure, the `data-id`
attribute) is defined by `docs/main-screen-template.html` and
`docs/main-screen-template.css`, which are part of the durable contract. Visibility
of an element means rendered visibility (computed style), not mere DOM presence.

## What every implementation must provide

A manifest at `app/harness.json`:

    {
      "start": "go run ./cmd/todomvc",
      "url": "http://localhost:8080",
      "seed": "go run ./cmd/evalctl seed",
      "read": "go run ./cmd/evalctl read"
    }

All commands are executed with `app/` as the working directory.

- `start` — builds (if needed) and serves the app at `url` until the process is
  killed. The runner starts it once per run, polls `url` until it responds with
  HTTP 200 (timeout: 60s), and kills the process group when the run ends.

- `seed` — reads a model from stdin and **replaces the entire persisted state**
  with it. The model is a JSON array of `{id, title, completed}` objects, possibly
  empty; ids are stored verbatim. Must work while the app server is running, and
  must exit 0 only on success.

- `read` — writes the currently persisted model to stdout as the same JSON array
  shape, in persisted order, and exits 0.

### Side-channel requirement

`seed` and `read` are a **local side channel**: they must access the storage
directly (e.g. opening the SQLite database file), never through a network
listener of any kind. The app must not expose any way to reach this
functionality over the web — the point is that no configuration mistake can
ever make it web-accessible. Access control is the filesystem, nothing else.

## Execution semantics, per scenario

1. Create a fresh browser context — no cookies, no client-side state survives
   from any previous scenario.
2. Run the `seed` command with the `GIVEN model:` — **including when it is
   `(empty)`**, which seeds the empty array. State lives server-side, so skipping
   the seed would leak the previous scenario's todos into this one.
3. Navigate to `url` + the `GIVEN route:` path (default `/`) with a full page
   load. The scenario's observations start from this load.
4. Execute the `WHEN:` steps in order. The `reload` action reloads the current
   URL.
5. Evaluate the `THEN` sections: `THEN page:` by projecting the DOM per `DSL.md`;
   `THEN model:` by running the `read` command; `THEN check:` per the check
   registry.

The app process is started once for the whole run, not per scenario; per-scenario
isolation is entirely the `seed` command's replace-everything semantics.

## Trust boundary

The adapter is generated together with the app, so the harness verifies the app
against the *model semantics* (what todos exist, in what state and order), not
against any particular storage schema. Whether the raw persisted format itself is
part of the contract is deliberately out of scope here — see the coverage audit;
if it ever becomes contractual, the enforcement must live outside `app/`.
