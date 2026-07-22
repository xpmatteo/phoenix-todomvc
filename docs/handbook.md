# Phoenix Handbook

Status: **durable, living document**. Practice learnings from building this
repository — how to make the Phoenix architecture practical. 
The `docs/principles.md` is the theory (distilled from Fowler); this is our own
practical advice, grounded in practical experience. 

## The components of the game

**Specs** live in `specs` and include:
  - business rules
  - screen templates
  - architectural rules

**Evals** are built out of a domain-specific testing language, and live in their 
own folder `evals`.  This makes them implementation-agnostic.  The DSL is specified
in `eval/DSL.md`, and how to build the **Eval runner** is specified in `evals/HARNESS.md`.
We can thus apply the Phoenix principle to the eval runner itself, which is disposable.

The **architectural rules** come from non-functional requirements, stated in the 
`specs/architecture.md` document itself.  Every architectural decision is presented in
ADR-style.

The architectural rules are tested through architectural tests, eg lints.

The application data must usually be durable, so its DB schema must also be.  A regeneration
that destroys data is not acceptable.  Therefore, the **DB schema** lives in its own folder
`schema`, and a method for schema evolution must be specified in the architecture document.

The evals will likely need side channels to inspect the application state, because
the alternative is to execute long strings of user-accessible operations to get the app
to the desired state, which is inconvenient because it's slow, reduces test reliability, 
and might push us to implement user-accessible operations that are only required by
the evals, which in turn increases the app surface and thus reduces security.



## Boundaries and document architecture

**Give every requirement exactly one home, chosen by a mechanical test.**
Ours: "could a black-box scenario fail because of this?" → spec; otherwise →
architecture doc (with named enforcement); evals↔app interface → harness
contract. Without the rule, requirements drift into whichever file is open.
*Episode:* the SPA→SSR switch rewrote one spec section and one contract file;
everything else stood. The rule is in `README.md`. (2026-07-22)

**Define eval seams on abstractions that survive architecture swaps.**
We defined the eval "model" as *the persisted todos*, not localStorage contents.
When the architecture changed from SPA to server-rendered-with-SQLite, all 35
scenarios survived untouched; only the harness contract (the seam's binding)
was rewritten. The payoff arrived the same day the abstraction was chosen.
(2026-07-22)

**Beware assumptions inherited from example material.**
localStorage leaked into our first harness contract from TodoMVC's heritage, not
from any decision. It survived until Captain Matt said "there is NO
localStorage!". Inherited examples smuggle architecture. (2026-07-22)

**Align contracts before generating; let documents take the damage.**
The runner build was stopped one tool-call before launch because the harness
assumptions were wrong. Cost of realignment: one rewritten markdown file. Had
the runner existed, the same realignment would have been a code migration.
(2026-07-22)

## Writing the spec

**Writing evals is spec archaeology — each eval decision forces the spec to
commit.** Every session of eval-writing surfaced spec gaps: id semantics were
entirely missing (made ids contractual via `data-id`), focus-after-save was
unspecified (kept focus out of the projection), counter-behavior-under-filter
was unstated (pinned to whole-model). The eval suite is the tool that finds
underspecification; budget for the spec changing while evals are written.
(2026-07-22)

**Permission is untestable as obligation.** The spec said the `#!/` route form
"is also allowed" — an eval can pin what MUST happen, never what MAY. Every
"may/allowed/preferably" in a spec is either advice (fine, unevaluated) or a
disguised decision waiting to be made. (2026-07-22)

## Writing evals

**Scenarios in a small DSL beat raw test code — but the DSL is itself a durable
artifact needing the same lock-down discipline.** The scenario files stay
readable and technology-free; the price is a meta-spec (`DSL.md`) whose
acceptance test is that a fresh agent can rebuild the runner from it alone. If
DSL semantics live only in the runner's code, "the implementation remembers"
has been recreated one level up. (2026-07-22)

**Closed vocabularies keep evals deterministic.** Free-prose actions ("the user
types ABC") would need an LLM to interpret — probabilistic exactly where rigor
must be rigid. A fixed verb table and check registry, extended deliberately,
read like prose but parse like code. (2026-07-22)

**Project what the user sees, not what the DOM contains.** The mark-all toggle
rendered as `[ ] Mark all as complete` — looking exactly like a todo item, when
the real UI shows a chevron inside the input row. A screenshot caught it; the
projection was wrong, not the app. The DOM is the measurement surface; the
rendering is the communication surface. (2026-07-22)

**Exact-diff projections cannot partially assert — keep optional facts out.**
Focus was in the projection first (a `_` marker), which would have forced every
scenario to pin where focus is, including where the spec is silent. Moved to an
explicit opt-in check. Corollaries: absence of a line asserts absence on the
page (an empty app's whole expectation is one line, `>`); under-specification
is expressed by omitting an assertion, with a NOTE saying why — the NOTEs are a
queue of pending spec decisions. (2026-07-22)

**Prefer a small escape hatch over a contorted notation.** The `<strong>`
counter requirement projected beautifully as markdown (`**2** items left`);
hover-visibility and focus did not project at all and became named checks. A
DSL covering 90% legibly plus a blunt instrument beats 100% illegibly.
(2026-07-22)

**Audit coverage clause-by-clause; every eval strategy has blind spots.** The
behavioral suite covered the spec's behavior fully and could not touch the
template-conformance and README constraints at all — a second, static kind of
eval is required. Found only by walking the spec line by line against the
scenarios. (2026-07-22)

**An eval seam is only as strong as the most regenerable thing on it.** The
seed/read adapter is generated with the app, so "storage uses keys id, title,
completed" was enforced by nothing durable — app and adapter could agree on the
wrong format and pass everything. Resolved not by hardening the harness but by
moving the schema to the durable side (AD-7). When a contract can't be
enforced at a boundary, ask whether something is on the wrong side of it.
(2026-07-22)

## Architecture

**Derive decisions from upfront NFRs, and give every decision named
enforcement.** Decisions became derivations (single binary ← operational
simplicity), and each AD names how it stays true across regenerations: an
architecture test, a static check, or review. A durable doc nothing enforces is
prose. (2026-07-22)

**Convert architecture claims into observable behavior where possible.**
"Server-rendered with JS as progressive enhancement" is invisible to the evals —
until you run the suite with JavaScript disabled, when it becomes a pass/fail
fact (editing scenarios exempted). The strongest architecture test is a
behavioral eval in disguise. (2026-07-22)

**Record rejected options inside the decision.** ADs carry a *Rejected:* line
(Atlas: conceptual mass; hand-rolled user_version: overruled). A regeneration
that only sees the winning option will happily relitigate the losers.
(2026-07-22)

**Data is the phoenix's limit.** The database file survives every regeneration —
so the migration chain is provenance for data and must be durable and
append-only, outside `app/`. The fix for a bad migration is another migration;
editing history is falsifying it. (2026-07-22)

## Process and provenance

**Commit reasons, not diff summaries.** Every commit message here records why —
the decision, its driver, what was rejected. The diff already says what.
(2026-07-22)

**Record decisions where they live.** Untestable-spec findings live as NOTEs in
the affected scenario file; design rationale for the schema lives in
`schema/README.md`; the boundary rule lives in the README. Decisions recorded
only in conversation die with the conversation. (2026-07-22)

**Distill sources with fidelity rules.** The principles doc separates the
author's claims (quoted, attributed per post) from our inferences (confined to
one labeled section). A distillation that blurs the two poisons every document
built on it. (2026-07-22)

## Pending experiments

- **Document-sufficiency test ("dogfood"):** build the runner from `DSL.md` +
  `HARNESS.md` + scenarios + templates ONLY; every ambiguity is logged as a
  defect in the documents (`QUESTIONS.md`), not solved by peeking at the spec.
  The gap log's length is a measurement of the durable artifacts' quality.
- **Architecture tests:** the roadmap in `docs/architecture.md` — JS-disabled
  eval run, dependency allowlist, import-graph check, side-channel isolation,
  static template/README conformance.
- **The first regeneration:** generate `app/` from the durable artifacts; then
  the real test — delete it and regenerate. Success is boredom.
