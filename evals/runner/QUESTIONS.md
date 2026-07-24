# Runner build — gap log

Defects/ambiguities found in the durable documents (`evals/DSL.md`, `evals/HARNESS.md`,
`evals/scenarios/*.md`, `spec/main-screen-template.{html,css}`) while building the runner
from them alone. Each entry: what was needed, where the docs fell short, what was chosen.

## 1. Who assigns ids omitted in `GIVEN model:` — DSL and HARNESS disagree

- **Needed:** ids for seeded todos whose model line carries no `#id`.
- **Docs:** DSL.md § Model notation said "when omitted, the **adapter** assigns a unique
  opaque id". HARNESS.md § seed says the seed payload is `{id, title, completed}` objects
  and "ids are stored **verbatim**" — the seed channel has no way to express an absent id,
  so the adapter never gets the chance to assign one.
- **Chose:** the runner generates a unique opaque id for each line and passes it to
  `seed`. Satisfies HARNESS.md literally.
- **RESOLVED (DSL.md § Todo identity):** the disagreement was a symptom of a deeper
  confusion — the notation treated `#id` as a *stored value*. It is now defined as a
  *symbolic label* for an auto-generated id. Ids are always generated (runner at seed,
  app on create); a `#name` only binds an id so a scenario can refer to it later. There
  is no "adapter assigns" path left to reconcile.

## 2. NOTE lines actually span multiple lines

- **Needed:** to skip commentary.
- **Docs:** DSL.md § Scenario files says "**Lines starting with** `NOTE:` … are ignored",
  but real notes are wrapped paragraphs whose continuation lines do not start with
  `NOTE:` (e.g. `item.md` "The destroy button removes its todo", `mark-all.md` last
  scenario, `routing.md`).
- **Chose:** `NOTE:` opens a comment that swallows following non-blank lines until a
  blank line.
- **RESOLVED (DSL.md § Scenario files):** the NOTE rule now states the comment spans the
  `NOTE:` line plus continuation lines, ending at the next blank line.

## 3. Prose outside sections (file preambles)

- **Needed:** to parse files like `routing.md`, which carry unindented prose between the
  `# h1` and the first `## scenario` ("Covers spec/spec.md …", "The default route …").
- **Docs:** DSL.md defines only `## heading` + keyword sections; it says nothing about
  preamble text.
- **Chose:** everything before the first `## ` heading is ignored. *Inside* a scenario,
  an unindented line that is neither a keyword nor a `NOTE:` is a parse error (strict, so
  typos in keyword lines can't be silently skipped).
- **RESOLVED (DSL.md § Scenario files):** preamble prose before the first `## ` is now
  documented as ignored, and the in-scenario grammar is stated as line-oriented and
  closed — an unrecognized unindented line is a parse error.

## 4. No synchronization/timing model anywhere

- **Needed:** how long to wait after a WHEN step before THEN sections may be evaluated,
  and what signals completion. DSL.md § Runner obligations only says to read the model
  "after the app has had the chance to persist"; HARNESS.md gives a timeout for start
  (60s) but none for anything else.
- **Chose:** every action is followed by a settle (100ms + poll `document.readyState`
  until `complete`, so full-page navigations triggered by clicks/Enter work), and every
  THEN section (page, model, check) is polled until it matches or a timeout (default 3s,
  `-wait` flag) expires. Failure output shows the last observation. Keyboard actions
  (`type`, `press`, `clear`) additionally wait up to 1s for *something* to be focused,
  because autofocus/app focus can land after the load event ("type into the currently
  focused element" is otherwise a race on page load).
- **RESOLVED (HARNESS.md § Settling):** the contract now states `THEN` is eventually
  consistent — the runner polls to a bounded deadline, and the app must reach its final
  observable state within that bound after an action with no further input. The exact
  constants (100ms, 1s focus wait, 3s deadline) stay runner details, as chosen here.

## 5. Input-row rendering when toggle and text are both present

- **Docs:** DSL.md § Page projection shows `>`, `> buy mil`, `v >`, `(v) >` but never a
  combined example.
- **Chose:** compositional: `v > buy mil` / `(v) > buy mil`; a `>` with empty value has
  no trailing space. Trailing whitespace is trimmed from all actual projection lines,
  because markdown scenario files cannot reliably carry trailing spaces in expected lines.
- **RESOLVED (DSL.md § Page projection):** the input value is now quoted — `> "buy mil"`,
  empty is `> ""`, combined is `v > "buy mil"` — and compared verbatim. This makes the
  combined case explicit and replaces the fragile trailing-space trim for the input row:
  quoted values compare exactly, only unquoted rendered text (titles, counter, filters)
  is trimmed. All ~29 input-row lines in the scenarios were updated to the quoted form.

## 6. What "styled as completed" means concretely

- **Docs:** DSL.md says `~…~` "reflects the completed styling", deliberately separate
  from the checkbox. Neither DSL.md nor HARNESS.md says which rendered signal to read.
  The template CSS styles completion as `li.completed label { text-decoration:
  line-through; … }`.
- **Chose:** computed `text-decoration-line` containing `line-through` on the item's
  label (a genuinely *rendered* signal, in the spirit of HARNESS.md's "rendered
  visibility"), rather than the presence of the `completed` class.
- **RESOLVED (DSL.md § Page projection) — reversed:** `~…~` now reads the `completed`
  class, not the computed strike-through. The template (line 21) makes setting that
  class the app's responsibility; the strike-through is downstream of the durable CSS,
  which the app does not regenerate, so reading the computed style would re-test our own
  stylesheet rather than the app's output. Independence from `[x]` (the checkbox's
  `checked`) is preserved. The runner must be updated to read the class.

## 7. Which element carries `data-id`

- **Docs:** DSL.md says "each rendered todo item carries a `data-id` attribute";
  the template HTML shows no `data-id` anywhere, so the exact element is unspecified.
- **Chose:** the `<li>` in `.todo-list` ("the rendered row's data-id", per the note in
  `item.md`). A row without the attribute renders without a `#id` prefix and fails the
  always-on integrity check.
- **RESOLVED (template + DSL.md § Concepts):** the template `<li>`s now carry `data-id`
  with an explanatory comment, so the DOM vocabulary is shown where the contract says it
  lives; DSL.md names the carrier as "the item's `<li>` in `.todo-list`".

## 8. "Selected" filter signal

- **Docs:** projection puts the selected filter in parentheses; the only durable marker
  is the template's `class="selected"` (rendered as a border color).
- **Chose:** the `selected` class on the `<a>`. (DOM-presence rather than
  rendered-style — the one place this projection trusts a class name directly, same as
  every TodoMVC implementation does.)
- **RESOLVED (DSL.md § Page projection):** stated — "selected" is read from the
  `selected` class on the filter's `<a>`, the app's declared marker, not the rendered
  border color. Consistent with `~…~` (#6) and `data-id` (#7): the projection reads the
  markers the HTML declares, and reserves computed-style reads for genuine visibility.

## 9. Editing-mode line when the app buggily also shows the normal controls

- **RESOLVED (DSL.md § Page projection) — reversed by the projection rework:** editing is
  now read from the `editing` class on the `<li>`, not from the edit field's rendered
  visibility. A row with `editing` projects the single `[edit: value]` line and no view
  line; a row without it projects the view line. This mirrors #6 (`~…~` reads the
  `completed` class): the projection reads declared markers, never inferred rendered
  styling, so it is a context-free function of the DOM (see the first follow-up below).
  The trade: the projection no longer catches an app that sets `editing` yet force-shows
  its view controls by other styling — it trusts the durable CSS to hide `.view` under
  `editing`. That is deliberate; hiding a projected region by any styling other than
  omission or `.hidden` is now a stated contract violation. The earlier resolution below
  (compositional two-line rendering, "editing judged by rendered visibility") is
  superseded. All item rules remain scoped to `<li>` within `ul.todo-list`, per Matt.
- **Docs:** DSL.md: rendering `[edit: value]` "also asserts that the item's normal
  controls … are hidden while editing", but doesn't say what the projection of a
  violating page looks like (unlike the `[x]`/`~…~` case, where it explains the diff
  will show it).
- **Chose:** compositional projection per row: a normal line if any view control
  (checkbox/label/destroy) is rendered visible, plus an `[edit: …]` line if the edit
  field is visible. A buggy app therefore produces two lines for one row and the diff
  exposes it. "Editing mode" is judged by the edit field's rendered visibility, not the
  `editing` class.

## 10. Footer line when parts are missing

- **Docs:** the footer format `-- <counter> | <filters> | [Clear completed]` assumes
  counter and filters exist (the template marks filters "Remove this if you don't
  implement routing").
- **Chose:** each segment appears only if present/visible: no visible filter links →
  no filters segment; missing/hidden new-todo input projects a sentinel line
  `(no new-todo input)` so the diff fails loudly instead of faking `>`.
- **RESOLVED:** two parts. (a) *Filters no longer optional* — Matt removed the template's
  "remove if you don't implement routing" comment; routing is mandatory, so when the
  footer is visible both counter and filters are always present and only `| [Clear
  completed]` stays conditional. No doc change needed; the format was already right.
  (b) *Missing input* — went with an always-on integrity check (DSL.md § Runner
  obligations) instead of the sentinel: the new-todo input must exist and be visible
  after every scenario. The projection grammar now describes only well-formed pages, and
  the sentinel line is dropped.

## 11. `click "Clear completed"` scope

- **Docs:** the action table lists `click "Clear completed"` as its own verb; it could
  be misread as a generic `click "<text>"`.
- **Chose:** closed set taken literally — only the exact string `click "Clear completed"`
  parses; any other `click "…"` is a parse error.
- **RESOLVED (DSL.md action table):** the row now says it is a fixed verb whose quotes are
  literal, not a generic click-by-text parameter.

## 12. `click toggle-all` target

- **Docs:** "click the mark-all-as-complete checkbox" — but the template CSS renders the
  `.toggle-all` checkbox as an invisible 1×1, opacity-0 control; users operate its
  `<label>` (the chevron).
- **Chose:** click the label associated with `.toggle-all` (via `for=` or sibling), which
  is what a real user does and toggles the checkbox natively.
- **RESOLVED (DSL.md action table):** the verb now targets the visible mark-all control
  (its `<label>`/chevron), noting the checkbox itself is rendered invisible and the click
  must toggle it.

## 13. Exact title matching vs. real-world whitespace

- **Docs:** `"title"` matches "by its exact current label text"; server-rendered HTML
  routinely pads text nodes with layout whitespace.
- **Chose:** `label.textContent.trim()` for both matching and projection. Titles with
  *internal* runs of whitespace are preserved exactly. Same trim applied to the counter
  text and filter link names.
- **RESOLVED (DSL.md action vocabulary):** the `"title"`-matching line no longer says
  "exact" — it now states matching is on trimmed text (leading/trailing ignored, internal
  preserved), applied wherever the projection reads rendered text. Consistent with the
  quoted-vs-unquoted rule from #5.

## 14. Matching a todo that is in editing mode

- **Docs:** `"title"` matches a **displayed** todo by label text, but in editing mode the
  label is hidden (only the row is displayed) — needed for
  `focus is in the edit field of "title"` right after `dblclick`.
- **Chose:** the row must be displayed; the label's textContent is used for matching even
  while hidden.
- **RESOLVED (DSL.md action vocabulary):** stated — identity is keyed on the displayed
  row, using the label text even when the label is hidden (as during editing), so
  `dblclick` then `focus is in the edit field of "title"` resolves.

## 15. "Fresh browser context" mechanism

- **Docs:** HARNESS.md step 1 demands no client-side state survive between scenarios but
  doesn't say how isolated the context must be.
- **Chose:** a fresh headless Chrome process with a fresh temporary user-data-dir per
  scenario (strongest isolation; ~0.5s/scenario overhead).
- **RESOLVED — no doc change:** not a contract question. HARNESS.md already states the
  requirement (no client-side state survives between scenarios); *how strongly* the
  runner isolates is a rebuildable-runner detail, deliberately left out of the durable
  contract.

## 16. Readiness poll target

- **Docs:** HARNESS.md: poll `url` until HTTP 200. Unstated: method, redirects.
- **Chose:** GET on the manifest `url` exactly, following Go's default redirect policy;
  any 200 counts.
- **RESOLVED (HARNESS.md `start`):** readiness is now defined as "a GET to `url` returns
  200, following redirects"; poll interval and backoff stated as runner details.

## 17. Scope of the always-on id invariant

- **Docs:** DSL.md § Model notation says ids must be non-empty "on every persisted item,
  **always**", and § Runner obligations describes the displayed-data-id-vs-persisted-id
  verification alongside the THEN model comparison. Whether these run for scenarios
  *without* a `THEN model:` is implicit.
- **Chose:** after every scenario's THEN sections, the runner reads the model and
  verifies: every persisted id is a non-empty string, and every displayed row's data-id
  equals the id of the persisted item with the same title (a displayed row with no
  persisted counterpart, or no data-id attribute, is also a failure). A JSON id that is
  not a string (e.g. a number) is rejected when decoding the read output.
- **RESOLVED (DSL.md § Runner obligations):** the id integrity check moved out of the
  `THEN model:` comparison bullet into the always-on invariants bullet, now explicitly
  unconditional — it runs after every scenario, including ones with only `THEN check:`,
  and the runner reads the model even when no `THEN model:` asked it to. The JSON id
  *wire type* is deliberately left unpinned: "opaque non-empty string" lives in the DSL
  abstract model; requiring the JSON encoding to be a string is a runner-decode detail,
  not an app obligation.

## 18. `reload` semantics

- **Docs:** DSL.md: "reload the page"; HARNESS.md: "reloads the current URL". After
  `click filter` navigated to `/completed`, "current URL" is taken as the browser's
  current location (so the filter persists — which `routing.md` "The active filter is
  persisted across a reload" indeed expects).
- **Chose:** browser reload of the current location (CDP Page.reload).
- **RESOLVED (DSL.md action table):** the `reload` row now says "reload the browser's
  current location (which may differ from `GIVEN route:` after a navigation)". The old
  confusing parenthetical ("persisted state must survive; in-memory state need not") was
  removed — it described the *application's* state, conflated the verb with its expected
  effect, and is already covered precisely by spec § Persistence ("persist immediately
  after every interaction", "Editing mode should not be persisted", "Reloading the page
  keeps the current filter").

## Follow-ups (raised while addressing the above, to do after the list)

- **DONE — Reworked the projection description toward simple, context-free rules**
  (DSL.md § Page projection). Every projected line now follows from a local rule on a
  declared marker — a class, attribute, `.value`, text, or presence — and the projection
  never infers state from rendered styling. Region visibility is a declared marker too:
  shown means present and without the `hidden` class (app hides by omission or `.hidden`).
  Editing reads the `editing` class (#9, reversed); `~…~` reads `completed` (#6);
  selection reads `selected` (#8); `data-id` (#7) was already declared. Computed-style
  visibility now lives only in `THEN check:` (destroy-on-hover, focus), where no marker
  can express the fact. Projected output strings are unchanged, so no scenario changed.
- **Add acceptance tests for the projection itself.** Pin the DOM→projection mapping
  with fixed HTML fixtures and expected projection strings, so the projector is verified
  independently of any app.
