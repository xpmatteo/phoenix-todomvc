# Runner build — gap log

Defects/ambiguities found in the durable documents (`evals/DSL.md`, `evals/HARNESS.md`,
`evals/scenarios/*.md`, `docs/main-screen-template.{html,css}`) while building the runner
from them alone. Each entry: what was needed, where the docs fell short, what was chosen.

## 1. Who assigns ids omitted in `GIVEN model:` — DSL and HARNESS disagree

- **Needed:** ids for seeded todos whose model line carries no `#id`.
- **Docs:** DSL.md § Model notation says "when omitted, the **adapter** assigns a unique
  opaque id". HARNESS.md § seed says the seed payload is `{id, title, completed}` objects
  and "ids are stored **verbatim**" — the seed channel has no way to express an absent id,
  so the adapter never gets the chance to assign one.
- **Chose:** the runner generates a unique opaque id (`gen-xxxxxxxx`) for each id-less
  line and passes it to `seed`. Satisfies HARNESS.md literally; DSL.md's "the adapter
  assigns" should probably be reworded to "the runner assigns".

## 2. NOTE lines actually span multiple lines

- **Needed:** to skip commentary.
- **Docs:** DSL.md § Scenario files says "**Lines starting with** `NOTE:` … are ignored",
  but real notes are wrapped paragraphs whose continuation lines do not start with
  `NOTE:` (e.g. `item.md` "The destroy button removes its todo", `mark-all.md` last
  scenario, `routing.md`).
- **Chose:** `NOTE:` opens a comment that swallows following non-blank lines until a
  blank line.

## 3. Prose outside sections (file preambles)

- **Needed:** to parse files like `routing.md`, which carry unindented prose between the
  `# h1` and the first `## scenario` ("Covers docs/spec.md …", "The default route …").
- **Docs:** DSL.md defines only `## heading` + keyword sections; it says nothing about
  preamble text.
- **Chose:** everything before the first `## ` heading is ignored. *Inside* a scenario,
  an unindented line that is neither a keyword nor a `NOTE:` is a parse error (strict, so
  typos in keyword lines can't be silently skipped).

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

## 5. Input-row rendering when toggle and text are both present

- **Docs:** DSL.md § Page projection shows `>`, `> buy mil`, `v >`, `(v) >` but never a
  combined example.
- **Chose:** compositional: `v > buy mil` / `(v) > buy mil`; a `>` with empty value has
  no trailing space. Trailing whitespace is trimmed from all actual projection lines,
  because markdown scenario files cannot reliably carry trailing spaces in expected lines.

## 6. What "styled as completed" means concretely

- **Docs:** DSL.md says `~…~` "reflects the completed styling", deliberately separate
  from the checkbox. Neither DSL.md nor HARNESS.md says which rendered signal to read.
  The template CSS styles completion as `li.completed label { text-decoration:
  line-through; … }`.
- **Chose:** computed `text-decoration-line` containing `line-through` on the item's
  label (a genuinely *rendered* signal, in the spirit of HARNESS.md's "rendered
  visibility"), rather than the presence of the `completed` class.

## 7. Which element carries `data-id`

- **Docs:** DSL.md says "each rendered todo item carries a `data-id` attribute";
  the template HTML shows no `data-id` anywhere, so the exact element is unspecified.
- **Chose:** the `<li>` in `.todo-list` ("the rendered row's data-id", per the note in
  `item.md`). A row without the attribute renders without a `#id` prefix and fails the
  always-on integrity check.

## 8. "Selected" filter signal

- **Docs:** projection puts the selected filter in parentheses; the only durable marker
  is the template's `class="selected"` (rendered as a border color).
- **Chose:** the `selected` class on the `<a>`. (DOM-presence rather than
  rendered-style — the one place this projection trusts a class name directly, same as
  every TodoMVC implementation does.)

## 9. Editing-mode line when the app buggily also shows the normal controls

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

## 11. `click "Clear completed"` scope

- **Docs:** the action table lists `click "Clear completed"` as its own verb; it could
  be misread as a generic `click "<text>"`.
- **Chose:** closed set taken literally — only the exact string `click "Clear completed"`
  parses; any other `click "…"` is a parse error.

## 12. `click toggle-all` target

- **Docs:** "click the mark-all-as-complete checkbox" — but the template CSS renders the
  `.toggle-all` checkbox as an invisible 1×1, opacity-0 control; users operate its
  `<label>` (the chevron).
- **Chose:** click the label associated with `.toggle-all` (via `for=` or sibling), which
  is what a real user does and toggles the checkbox natively.

## 13. Exact title matching vs. real-world whitespace

- **Docs:** `"title"` matches "by its exact current label text"; server-rendered HTML
  routinely pads text nodes with layout whitespace.
- **Chose:** `label.textContent.trim()` for both matching and projection. Titles with
  *internal* runs of whitespace are preserved exactly. Same trim applied to the counter
  text and filter link names.

## 14. Matching a todo that is in editing mode

- **Docs:** `"title"` matches a **displayed** todo by label text, but in editing mode the
  label is hidden (only the row is displayed) — needed for
  `focus is in the edit field of "title"` right after `dblclick`.
- **Chose:** the row must be displayed; the label's textContent is used for matching even
  while hidden.

## 15. "Fresh browser context" mechanism

- **Docs:** HARNESS.md step 1 demands no client-side state survive between scenarios but
  doesn't say how isolated the context must be.
- **Chose:** a fresh headless Chrome process with a fresh temporary user-data-dir per
  scenario (strongest isolation; ~0.5s/scenario overhead).

## 16. Readiness poll target

- **Docs:** HARNESS.md: poll `url` until HTTP 200. Unstated: method, redirects.
- **Chose:** GET on the manifest `url` exactly, following Go's default redirect policy;
  any 200 counts.

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

## 18. `reload` semantics

- **Docs:** DSL.md: "reload the page"; HARNESS.md: "reloads the current URL". After
  `click filter` navigated to `/completed`, "current URL" is taken as the browser's
  current location (so the filter persists — which `routing.md` "The active filter is
  persisted across a reload" indeed expects).
- **Chose:** browser reload of the current location (CDP Page.reload).
