# Eval Harness Contract

Status: **durable artifact**. This document is the contract every generated app
must satisfy so a single, unchanging test runner can drive it — start it, seed it,
and read its state back. `DSL.md` defines *what* the scenarios mean;
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

The runner reads the page using a fixed DOM vocabulary: element classes, page
structure, and the `data-id` attribute. Two files define it:
`spec/main-screen-template.html` and `spec/main-screen-template.css`. Both are part
of the durable contract.

The page projection reads only declared markers — classes, attributes, values, text,
and presence — never computed style (DSL.md § Page projection). Computed style enters
only where a scenario asks directly whether an element is visible: the `THEN check:`
visibility checks (the destroy button revealed on hover). There, visible means rendered
visibility — an element present in the DOM but hidden does not count.

## What every implementation must provide

A test needs two things the web page cannot give the runner on its own: a way to
put the app into a known starting state *before* the test runs, and a way to see
the resulting state *after* the test has acted. `seed` does the first — it writes
the scenario's `GIVEN model:` into storage. `read` does the second — it reports
what the app persisted, so a `THEN model:` check can compare. `start` and `url`
just get the app running.

Each implementation declares these commands in a manifest at `app/harness.json`:

    {
      "start": "go run ./cmd/todomvc",
      "url": "http://localhost:8080",
      "seed": "go run ./cmd/evalctl seed",
      "read": "go run ./cmd/evalctl read"
    }

The values above are only an example — this one happens to be a Go app. The
contract mandates the four keys and what each command must do; each implementation
supplies its own values, in whatever technology it chooses.

All commands are executed with `app/` as the working directory.

- `start` — builds (if needed) and serves the app at `url` until the process is
  killed. The runner starts it once per run and polls for readiness (timeout: 60s):
  ready means a GET to `url` returns HTTP 200, following redirects (so answering `/`
  with a 302 to a landing page that returns 200 counts). The runner kills the process
  group when the run ends. The poll interval and backoff are runner details.

- `seed` — reads a model from stdin and **replaces the entire persisted state**
  with it. The model is a JSON array of `{id, title, completed}` objects, possibly
  empty; ids are stored verbatim. Must work while the app server is running, and
  must exit 0 only on success.

- `read` — writes the currently persisted model to stdout as the same JSON array
  shape, in persisted order, and exits 0.

### Side-channel requirement

`seed` and `read` are a **local side channel**. They reach the app's stored state
directly — for example by opening its SQLite file, or by connecting to whatever
database it uses — not through the app's own web interface.

The rule this protects: the app must expose no web route, or any other public
endpoint, that can seed or read the model. That way no configuration mistake can
ever make this functionality reachable over the web. To run these commands you need
local access to the host where the app and its storage live — never just an HTTP
request to the app.

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

### Settling: `THEN` is eventually consistent

An action may not finish the instant it is dispatched — a click can trigger a
full-page navigation, and the spec requires the app to persist after every
interaction, which may complete asynchronously. So the runner must not sample the
page or the model once, immediately after the last `WHEN` step, and judge it. It
must treat every `THEN` as **eventually consistent**: re-evaluate the section until
it holds, or until a bounded timeout expires, and only then report the last
observation as the failure.

This places one obligation on every implementation: after an action with no further
input, the app must reach its final observable state — rendered page and persisted
model — within that bound. An app that keeps changing what it shows, or that
persists only after some longer delay, violates the contract.

The bound and the poll interval are runner details, not contract: a rebuilt runner
may choose its own constants (the current runner defaults to a 3s deadline per
`THEN`, overridable with `-wait`). What is durable is the shape — poll to a
deadline, never sample once — and the app-side guarantee that a bounded, no-further-
input window is enough to settle.

## Trust boundary

The adapter is generated together with the app, not independently. So the harness
checks the app at the level of *model semantics*: which todos exist, in what state,
and in what order.

The storage format is also part of the contract — the durable, append-only schema
in `schema/` (AD-7). The harness does not check it directly. Instead, a human
reviews the adapter commands against that schema. And if the adapter and the app
ever disagree, scenarios start to fail — so the problem surfaces on its own.
