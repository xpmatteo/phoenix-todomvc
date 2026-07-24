# Phoenix Handbook

Status: **durable, living document**. What we learned building this repository,
and how to make the Phoenix architecture practical. `docs/principles.md` holds
the theory, distilled from Fowler. This holds our own advice, grounded in
experience.

## The components of the game

**Specs** live in `spec/` and include:
  - business rules
  - screen templates
  - architectural rules

**Evals** are written in a domain-specific testing language and live in `evals/`.
The DSL keeps them implementation-agnostic. `evals/DSL.md` specifies the DSL;
`evals/HARNESS.md` specifies how to build the **eval runner**. The runner is
disposable, so the Phoenix principle applies to it too.

**Architectural rules** come from non-functional requirements. They live in
`spec/architecture.md`, each stated as an ADR. Architectural tests enforce them,
for example lints.

Application data is usually durable, so its DB schema must be too. A regeneration
that destroys data is not acceptable. So the **DB schema** lives in its own folder
`schema`, and the architecture document specifies how the schema evolves.

Evals will likely need side channels to inspect application state. The
alternative is to drive the app through long strings of user-accessible
operations to reach a given state, which is slow and flaky. Worse, it pushes us
to add operations only the evals need, which grows the app surface and weakens
security.



## Boundaries and document architecture

**Give every requirement exactly one home, chosen by a mechanical test.**
Ours: could a black-box scenario fail because of this? If yes, it goes in the
spec. If no, it goes in the architecture doc, with named enforcement. The
evals-app interface goes in the harness contract. Without the rule, requirements
drift into whichever file is open. The SPA-to-SSR switch rewrote one spec
section and one contract file; everything else stood. The rule is in
`README.md`. (2026-07-22)

**Define eval seams on abstractions that survive architecture swaps.**
We defined the eval "model" as the persisted todos, not localStorage contents.
When the architecture changed from SPA to server-rendered SQLite, all 35
scenarios survived untouched; we rewrote only the harness contract. (2026-07-22)

**Beware assumptions inherited from example material.**
localStorage leaked into our first harness contract from TodoMVC's heritage, not
from any decision we made. It survived until Captain Matt said "there is NO
localStorage!". Inherited examples carry hidden architecture. (2026-07-22)

**Align contracts before generating.**
We stopped the runner build one tool-call before launch because the harness
assumptions were wrong. Realigning cost one rewritten markdown file. Had the
runner existed, the same fix would have been a code migration. (2026-07-22)

## Writing the spec

**Writing evals forces the spec to commit.** Every eval-writing session
surfaced spec gaps. Id semantics were missing, so we made ids contractual via
`data-id`. Focus-after-save was unspecified, so we kept focus out of the
projection. Counter-behavior-under-filter was unstated, so we pinned it to the
whole model. The eval suite finds underspecification; budget for the spec
changing while evals are written. (2026-07-22)

**Permission is untestable as obligation.** The spec said the `#!/` route form
"is also allowed". An eval can pin what MUST happen, never what MAY. Every
"may/allowed/preferably" in a spec is either advice, which goes unevaluated, or
a decision still waiting to be made. (2026-07-22)

## Writing evals

**Scenarios in a small DSL beat raw test code, but the DSL needs the same
lock-down discipline.** The scenario files stay readable and technology-free.
The price is a meta-spec (`DSL.md`) whose acceptance test is that a fresh agent
can rebuild the runner from it alone. If DSL semantics live only in the runner's
code, "the implementation remembers" has just moved up one level. (2026-07-22)

**Closed vocabularies keep evals deterministic.** Free-prose actions like "the
user types ABC" would need an LLM to interpret, making them probabilistic where
they must be exact. A fixed verb table and check registry, extended
deliberately, read like prose but parse like code. (2026-07-22)

**Project what the user sees, not what the DOM contains.** The mark-all toggle
rendered as `[ ] Mark all as complete`, looking exactly like a todo item, when
the real UI shows a chevron inside the input row. A screenshot caught it: the
projection was wrong, not the app. Project the rendering the user reads, not the
raw DOM. (2026-07-22)

**Exact-diff projections cannot partially assert, so keep optional facts out.**
Focus was in the projection first, as a `_` marker. That would have forced every
scenario to pin where focus is, including where the spec is silent. We moved it
to an opt-in check. Two corollaries follow. Absence of a line asserts absence on
the page; an empty app's whole expectation is one line, `>`. Under-specification
is expressed by omitting an assertion, with a NOTE saying why. Those NOTEs are a
queue of pending spec decisions. (2026-07-22)

**Prefer a small escape hatch over a contorted notation.** The `<strong>`
counter projected cleanly as markdown (`**2** items left`). Hover-visibility and
focus did not project at all, so they became named checks. A DSL that covers 90%
legibly plus a blunt instrument beats 100% illegibly. (2026-07-22)

**Audit coverage clause-by-clause; every eval strategy has blind spots.** The
behavioral suite covered the spec's behavior fully but could not touch the
template-conformance and README constraints at all. Those need a second, static
kind of eval. We found the gap only by walking the spec line by line against the
scenarios. (2026-07-22)

**An eval seam is only as strong as the most regenerable thing on it.** The
seed/read adapter is generated with the app, so "storage uses keys id, title,
completed" was enforced by nothing durable. App and adapter could agree on the
wrong format and pass everything. We fixed it not by hardening the harness but by
moving the schema to the durable side (AD-7). When a contract can't be enforced
at a boundary, ask whether something is on the wrong side of it. (2026-07-22)

## Architecture

**Derive decisions from upfront NFRs, and give every decision named
enforcement.** Decisions became derivations: single binary follows from
operational simplicity. Each AD names how it stays true across regenerations,
whether an architecture test, a static check, or review. A durable doc that
nothing enforces is just prose. (2026-07-22)

**Convert architecture claims into observable behavior where possible.**
"Server-rendered with JS as progressive enhancement" is invisible to the evals
until you run the suite with JavaScript disabled, when it becomes a pass/fail
fact (editing scenarios exempted). An architecture claim tests best as a
behavioral eval. (2026-07-22)

**Record rejected options inside the decision.** ADs carry a *Rejected:* line
(Liquibase: excessive conceptual mass; hand-rolled user_version: overruled). 
A regeneration that sees only the winning option will evaluate again the losers. (2026-07-22)

**Data is the phoenix's limit.** The database file survives every regeneration,
so the migration chain is the provenance for that data. It must be durable and
append-only, outside `app/`. The fix for a bad migration is another migration;
editing the history falsifies it. (2026-07-22)

## Process and provenance

**Commit reasons, not diff summaries.** Every commit message here records why:
the decision, its driver, what was rejected. The diff already says what.
(2026-07-22)

**Record decisions where they live.** Untestable-spec findings live as NOTEs in
the affected scenario file. Schema rationale lives in `schema/README.md`. The
boundary rule lives in the README. A decision recorded only in conversation is
lost when the conversation ends. (2026-07-22)

**Distill sources with fidelity rules.** The principles doc separates the
author's claims, quoted and attributed per post, from our own inferences,
confined to one labeled section. A distillation that blurs the two corrupts
every document built on it. (2026-07-22)

## Pending experiments

- **Document-sufficiency test ("dogfood"):** build the runner from `DSL.md` +
  `HARNESS.md` + scenarios + templates ONLY; every ambiguity is logged as a
  defect in the documents (`QUESTIONS.md`), not solved by peeking at the spec.
  The gap log's length is a measurement of the durable artifacts' quality.
- **Architecture tests:** the roadmap in `spec/architecture.md`: JS-disabled
  eval run, dependency allowlist, import-graph check, side-channel isolation,
  static template/README conformance.
- **The first regeneration:** generate `app/` from the durable artifacts. Then
  the real test: delete it and regenerate. Success is boredom.
