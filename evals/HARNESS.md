# Eval Harness Contract

Status: **durable artifact**. Defines the boundary between the test runner and each
generated implementation of the app. `DSL.md` defines *what* the scenarios mean;
this document defines *how* a runner gets any implementation running, seeded, and
read back. A fresh agent must be able to build the runner from this document plus
`DSL.md`, the scenario files, and the two template files — nothing else.

## Roles

- **Runner** (`evals/runner/`) — a Go program. Parses the scenario files per
  `DSL.md`, drives a real browser over the Chrome DevTools Protocol, and uses the
  adapter files described below. It contains no knowledge of any particular
  implementation and is never modified when the app is regenerated. It is
  disposable in principle (rebuildable from these documents) but stable in
  practice.

- **Adapter** — shipped *by each implementation* inside `app/`, regenerated along
  with it. The only place that knows how the app is built, served, and persisted.

The DOM vocabulary the runner needs (element classes, structure, the `data-id`
attribute) is defined by `docs/main-screen-template.html` and
`docs/main-screen-template.css`, which are part of the durable contract. Visibility
of an element means rendered visibility (computed style), not mere DOM presence.

## What every implementation must provide

### 1. `app/harness.json`

    {
      "start": "go run ./cmd/todomvc",
      "url": "http://localhost:8080"
    }

- `start` — a command, executed with `app/` as working directory, that builds (if
  needed) and serves the app at `url` until the process is killed. The runner polls
  `url` until it responds with HTTP 200 (timeout: 60s), and kills the process group
  when the run ends.
- `url` — where the running app is reachable.

### 2. `app/eval-adapter/seed.js`

A single JS **function expression** taking the model and establishing it as the
persisted state:

    (model) => { localStorage.setItem('todos', JSON.stringify(model)) }

- `model` is an array of `{id, title, completed}` objects (ids seeded verbatim).
- Executed by the runner in a page on the app's origin. May return a Promise; the
  runner awaits it.

### 3. `app/eval-adapter/read.js`

A single JS function expression returning the currently persisted model (or a
Promise of it), in persisted order, as the same array shape:

    () => JSON.parse(localStorage.getItem('todos') ?? '[]')

A server-side implementation would instead `fetch()` its own store through an
endpoint it provides for this purpose. Either way the runner only ever sees
`[{id, title, completed}]`.

## Execution semantics, per scenario

1. Create a fresh browser context — no cookies, no storage, nothing survives from
   any previous scenario.
2. Navigate to `url`, evaluate `seed.js` with the `GIVEN model:` (skip evaluation
   for `(empty)`).
3. Navigate to `url` + the `GIVEN route:` hash (or plain `url` if none) with a full
   page (re)load. The scenario's observations start from this load.
4. Execute the `WHEN:` steps in order. The `reload` action reloads the current URL;
   the browser context (and therefore client-side storage) is preserved.
5. Evaluate the `THEN` sections: `THEN page:` by projecting the DOM per `DSL.md`;
   `THEN model:` by evaluating `read.js`; `THEN check:` per the check registry.

## Trust boundary

`seed.js`/`read.js` are generated together with the app, so the harness verifies
the app against the *model semantics* (what todos exist, in what state and order),
not against any particular storage representation. Whether the raw persisted format
itself is part of the contract is deliberately out of scope here — see the coverage
audit; if it ever becomes contractual, the enforcement must live outside `app/`.
