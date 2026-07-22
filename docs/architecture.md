# Architecture

Status: **durable artifact**. This document records the non-functional
requirements (NFRs) and the architecture decisions (ADs) derived from them.
Decisions constrain *how* the app is built; the behavioral evals are deliberately
blind to them (see the boundary rule in `README.md`), so every decision names its
own enforcement. Decisions are appended, not silently rewritten; when one is
reversed, mark it superseded and say why.

## Non-functional requirements

- **NFR-1 Operational simplicity — deployment.** The app is a single binary
  produced by a single build step. No separate frontend build, no asset pipeline.
- **NFR-2 Operational simplicity — storage.** No database server to install,
  configure, or operate. Storage is embedded in the app.
- **NFR-3 Performance: modest, by design.** Single-user scale; no performance
  requirement beyond ordinary responsiveness. The *absence* of performance
  pressure is itself a requirement: it licenses the simplest adequate technology
  and forbids complexity justified by speculative load.
- **NFR-4 Changeability of business logic.** The business logic must be easy to
  understand and change in isolation from the technology it happens to be
  attached to.
- **NFR-5 Ease of app evolution.** The app must be able to change — including
  its data shape — without losing data. Schema changes are managed, repeatable,
  and survive regeneration: user data outlives every implementation.

## Architecture decisions

### AD-1 Server-rendered Go application — no SPA (NFR-1)

One binary serves HTML rendered on the server. Client-side JavaScript is allowed
only as progressive enhancement for interactions the spec itself requires
(double-click to edit, Escape to cancel); rendering and filtering happen
server-side. Static assets are embedded (`go:embed`) so the binary is
self-contained.
*Consequences:* routing is URL paths; every core flow must work with JavaScript
disabled.
*Enforcement:* planned architecture test — run the eval suite with JavaScript
disabled; everything except the editing-interaction scenarios must pass.
(2026-07-22)

### AD-2 SQLite, embedded (NFR-2, NFR-3)

Persistence is a SQLite database file opened by the app. WAL mode, so the eval
side channel (AD-5) can access the file while the server runs.
*Consequences:* backup = copy a file; no connection configuration exists.
*Enforcement:* review; the harness itself fails if seed/read can't operate
alongside the running server. (2026-07-22)

### AD-3 Standard library only for HTTP and templating (NFR-3, NFR-4)

`net/http` and `html/template`; no web framework, no router dependency. NFR-3
says we don't need their performance or features; NFR-4 says every dependency is
conceptual mass a reader must load.
*Enforcement:* planned static check on `go.mod` (allowlist of dependencies).
(2026-07-22)

### AD-4 Model packages separated from technology packages (NFR-4)

The business logic — the todo model, its operations, filtering rules — lives in
package(s) that import no technology: no `net/http`, no `database/sql`, no
templating. Technology packages (HTTP handlers, SQLite storage, rendering) depend
on the model, never the reverse.
*Enforcement:* planned static import check (architecture test over the package
import graph). (2026-07-22)

### AD-5 Eval side channel is a local executable, structurally web-unreachable

The eval harness's seed/read operations are executables that open the database
file directly. No network listener of any kind is involved, so no configuration
mistake can ever expose them to the web; access control is the filesystem.
The contract is canonical in `evals/HARNESS.md`.
*Enforcement:* planned static check — no HTTP handler may reach the side-channel
code path; review. (2026-07-22)

### AD-6 Schema migrations with goose, embedded, applied at startup (NFR-5, NFR-1)

Schema changes are managed with `pressly/goose` used as a *library*: plain SQL
migration files with `-- +goose Up` / `-- +goose Down` annotations, embedded into
the binary, applied automatically at startup. No separate migrate step exists
(NFR-1). The SQLite driver is the pure-Go `modernc.org/sqlite`, keeping the single
build free of cgo and a C toolchain.
*Rejected:* Atlas/Liquibase-style declarative systems (conceptual mass, commercial
gravity — against NFR-3's simplicity license); a hand-rolled `PRAGMA user_version`
loop (Captain Matt's call: a real system, not a mechanism).
*Consequences:* the dependency allowlist (AD-3) is exactly: `pressly/goose`,
`modernc.org/sqlite`.
*Enforcement:* the AD-3 allowlist check; the harness fails if the app cannot start
against an existing database. (2026-07-22)

### AD-7 The schema is a durable asset, outside `app/` (NFR-5)

The database file survives regeneration — data is the one thing a phoenix cannot
burn. Therefore the migration chain is provenance for data and lives on the
durable side: `schema/` at the repository root, append-only. A regenerated
implementation may **append** migrations (reviewed like any durable-artifact
change) but never rewrites, reorders, or deletes existing ones, and must apply
exactly this chain at startup.
So that the disposable app packages can embed and import the durable schema, the
repository is a single Go module rooted at the repo root; `evals/runner/` stays
its own module (it must not share the app's dependency graph).
*Consequences:* the persisted storage format is now contractual, resolving the
coverage audit's open question #3 — the eval side channel's seed/read commands
must read and write this schema, and can be reviewed against it.
*Enforcement:* append-only is checked by provenance (git history of `schema/`);
review of any appended migration. (2026-07-22)

## Artifact constraints

Non-behavioral obligations on the generated implementation (moved from the spec):

- `app/` includes a README describing the general implementation and the build
  process.
- The served HTML stays as close to `docs/main-screen-template.html` as possible;
  template comments are not shipped.
*Enforcement:* planned static conformance checks alongside the runner.

## Architecture tests — roadmap

The enforcement column above, gathered: (1) eval suite under disabled JavaScript
(AD-1); (2) dependency allowlist on `go.mod` (AD-3); (3) import-graph check for
model/technology separation (AD-4); (4) no-web-path-to-side-channel check (AD-5);
(5) static conformance: app README exists, served HTML vs template (artifact
constraints). None exist yet; they should live under `evals/` beside the runner.
