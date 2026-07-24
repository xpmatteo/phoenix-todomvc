# Eval DSL — Definition

Status: **durable artifact**. This document defines the scenario language used by all
files in `evals/scenarios/`. The scenarios and this document outlive any implementation
of the app *and any implementation of the test runner*. A fresh agent must be able to
rebuild the runner from this document plus the scenario files alone; if it can't, this
document is incomplete — fix it here, not in the runner.

Changing this document invalidates the runner and potentially every scenario file.
Change it deliberately.

## Concepts

**Model** — the persisted collection of todos, in order. The model is abstract: scenarios
never say *how* todos are persisted. The runner binds the model to a concrete storage
mechanism through a per-implementation **adapter** with two operations:

- `seed(model)` — establish the given todos as the *entire* persisted state (replacing
  whatever was there) before the scenario's first page load
- `read() → model` — return the currently persisted todos

How seed/read reach the storage is defined by the harness contract (`HARNESS.md`);
nothing in the scenarios depends on that mechanism. The adapter is disposable code,
regenerated along with the app it belongs to.

The persisted representation carries `id`, `title`, `completed` per item (see
`spec/spec.md` § Persistence). Ids are part of the contract: they are opaque non-empty
strings, and each rendered todo item carries a `data-id` attribute equal to its
persisted id — on the item's `<li>` in `.todo-list` (shown in the template). Ids are
always generated automatically — the runner mints one when it
seeds a todo, the app mints one when it creates a todo. A scenario never dictates the
stored value. When identity matters across a step, a scenario gives an id a **name** to
refer to it later (see "Todo identity").

**Page** — an ASCII projection of the rendered UI, defined under "Page projection" below.
Only the projected parts are compared; everything else on the page is ignored.

## Scenario files

Scenario files are markdown. A file may open with preamble prose between the `# h1`
and the first `## heading` — it describes the file for human readers and the runner
ignores everything before the first `## ` heading.

Each `## heading` starts a scenario; the heading is the scenario name. A scenario
contains these sections, each introduced by a keyword line and followed by an indented
block (4 spaces):

| Section | Required | Meaning |
|---|---|---|
| `GIVEN model:` | yes | model seeded before the app loads |
| `GIVEN route:` | no | initial URL path, written inline on the keyword line itself, e.g. `GIVEN route: /active` (default: `/`) |
| `WHEN:` | no | user actions, one per line, executed in order |
| `THEN page:` | no* | expected page projection after the last action |
| `THEN model:` | no* | expected model after the last action |
| `THEN check:` | no | extra named checks (see registry below) |

*At least one `THEN` section is required.

Every scenario runs against a fresh app instance with no state other than what
`GIVEN` establishes. Scenarios are independent; order of execution is irrelevant.

A `NOTE:` line inside a scenario opens commentary for humans and agents, which the
runner ignores. Notes are usually wrapped paragraphs, so the comment runs from the
`NOTE:` line through any following non-blank lines and ends at the next blank line.

The grammar inside a scenario is line-oriented and closed. Every non-blank line is one
of: a keyword line (from the table above), an indented block line, or a `NOTE:` line
and its continuations. An unindented line that is none of these — a misspelled keyword,
stray prose — is a parse error, not silently skipped, so a typo cannot quietly drop a
section and let a broken scenario pass.

Todo titles within a scenario must be unique.

## Model notation

One line per todo, in persisted order:

    [ ] title         — active todo
    [x] title         — completed todo
    #name [ ] title   — todo whose id is bound to the label `#name` (see "Todo identity")
    (empty)           — no todos

## Todo identity

Ids are never written literally into a scenario; they are generated automatically. A
`#name` on a model line (or a page line, see "Page projection") is a **symbolic label
for an id**, not a stored value. It exists only so a scenario can say "this is the same
todo I referred to earlier."

The rule is one binding per name, scoped to a single scenario:

- **First occurrence binds.** The first time a `#name` appears, it is bound to that
  todo's real id. If it first appears in `GIVEN model:`, the runner generates an opaque
  id, records the binding, and seeds that id. If it first appears in a `THEN` section
  (a todo the app created during the run), it binds to whatever id the app minted.
- **Later occurrences assert equality.** Every further use of the same `#name` — in
  `THEN model:` or on a `THEN page:` item line — asserts that row's real id equals the
  bound one. This is how a scenario pins that a todo kept its identity across a step,
  even when its title changed (see `editing.md`).

Lines without a `#name` still carry a real id; it is simply not named, so nothing later
refers to it. The runner always enforces that every persisted id is a non-empty string,
named or not (see "Runner obligations").

## Page projection

The projection is an ASCII rendering of the page, top to bottom. `THEN page:` compares
it line by line; anything not projected is ignored.

The rules below are pinned by example in `evals/projection/`: fixed HTML fragments,
deduced from the template, each paired with its exact projection. A rebuilt projector
must reproduce every one. That suite verifies this mapping without a running app; keep
it in step when these rules change.

Every line follows from a **local rule on a declared marker** — a class, an attribute,
an input's `.value`, an element's text, or an element's presence. The projection never
infers state from rendered styling: no computed colors, strike-through, or visibility.
Two reasons. Reading the durable CSS would test our own stylesheet instead of the app.
And a marker rule is unambiguous, so the projector never guesses what the app meant.
This puts one duty on the app: **declare state through the markers below**, the way the
template already declares `data-id`, `completed`, and `selected`.

**Shown vs hidden.** The app hides a region by leaving it out of the DOM or giving it
the `hidden` class (durable CSS: `.hidden { display: none }`). A region is **shown**
when its element is present and carries no `hidden` class. The app must not hide a
projected region by any other styling — the projection would still count it as shown.
Genuinely visual facts no marker can express — the destroy button revealed on `:hover`,
document focus — are not projected; assert them with `THEN check:`, which reads computed
visibility (HARNESS.md).

The lines, in order:

**1. Header row** — always first. It reads the new-todo input (`.new-todo`) and the
mark-all toggle (`#toggle-all`):

    > "value"        — toggle hidden: input alone
    v > "value"      — toggle shown, unchecked
    (v) > "value"    — toggle shown, checked (every todo completed)

- `"value"` is the input's `.value`, verbatim, in double quotes. Whitespace inside the
  quotes is significant; an empty value is `""`. It is quoted because it is a literal
  field value the user could submit, compared exactly. (Item titles below are unquoted
  rendered text, matched on trimmed text; the quotes mark that difference.)
- The chevron prefix appears only when the mark-all toggle is shown — `(v)` when
  `#toggle-all` is `checked`, `v` when it is not. Parens mean "on", the same convention
  as the selected filter. The toggle is shown exactly when the main section is (both
  live or vanish together), so the chevron's presence doubles as the "main section is
  shown" signal.

The header is always present — the spec keeps it visible in every state — so this line
never disappears. If the app fails to render it, the runner fails loudly rather than
projecting a placeholder (see "Runner obligations").

**2. Todo items** — one line per shown `<li>` in `ul.todo-list`, in order. A filtered-out
todo is not rendered, so it produces no `<li>` and no line. Scoping to `ul.todo-list`
keeps these rules off any other list or checkbox the UI may grow later. Any item line may
carry a leading `#name`, binding or asserting the row's `data-id` (see "Todo identity").

- **Editing row** — the `<li>` carries the `editing` class:

      [edit: value]

  `value` is the `.edit` field's `.value`. `editing` is the spec's declared marker for
  edit mode, and the durable CSS hides the row's view controls under it. So an editing
  row projects this one line and no view line — read from the class, not from which
  controls happen to be visible.

- **View row** — the `<li>` does not carry `editing`:

      [ ] title        — the `.toggle` checkbox is unchecked
      [x] title        — the `.toggle` checkbox is checked
      [x] ~title~      — checked AND the `<li>` carries the `completed` class

  `[x]` reads the checkbox's `checked`; `~…~` reads the `completed` class. They are two
  independent app outputs, so a buggy app can produce one without the other and the diff
  shows it. `title` is the label's text, trimmed. (The durable CSS renders `completed`
  as strike-through; the projection reads the class, not the rendered line, so the
  assertion stays about the app's output.)

**3. Footer** — one line, only when the footer (`.footer`) is shown:

    -- <counter> | <filters>
    -- <counter> | <filters> | [Clear completed]

- `<counter>`: the `.todo-count` text, with its `<strong>`-wrapped number rendered
  `**n**`: `**2** items left`. (Projecting the `<strong>` pins the wrapper the spec
  requires.)
- `<filters>`: the three filter links in order, the one carrying the `selected` class in
  parentheses: `(All) Active Completed`. Selection is read from the class, the app's
  declared marker, not the rendered border color.
- ` | [Clear completed]` appears only when the Clear-completed button (`.clear-completed`)
  is shown.

Focus is deliberately **not** projected — asserting it on every line would pin behavior
the spec leaves open. Assert focus explicitly via `THEN check:`.

## Action vocabulary (`WHEN:`)

Closed set. A scenario may only use these verbs; if a behavior can't be expressed,
extend this table deliberately (and note why) rather than improvising in a scenario.

| Verb | Meaning |
|---|---|
| `type "text"` | type text into the currently focused element |
| `press Enter` / `press Escape` | press the key in the currently focused element |
| `clear` | clear the currently focused input |
| `blur` | remove focus from the currently focused element |
| `click toggle of "title"` | click the checkbox of the todo with that title |
| `click destroy of "title"` | click the destroy button of that todo (hovering first) |
| `dblclick "title"` | double-click the label of that todo |
| `click toggle-all` | click the visible mark-all control the user operates — its `<label>` (the chevron), since the durable CSS renders the `.toggle-all` checkbox itself invisible; this must toggle that checkbox |
| `click "Clear completed"` | click the Clear-completed button — a fixed verb; the quotes are literal, not a generic click-by-text parameter |
| `click filter "Active"` | click that filter link (All / Active / Completed) |
| `go to "/active"` | navigate to that URL path |
| `reload` | reload the browser's current location (which may differ from `GIVEN route:` after a navigation) |
| `hover "title"` | move the pointer over that todo |

`"title"` matches a displayed todo by its current label text, compared on trimmed text:
leading and trailing whitespace (the layout padding HTML rendering adds) is ignored,
internal whitespace is preserved. The same trim applies wherever the projection reads
rendered text — item titles, the counter, filter link names.

Identity is keyed on the row, not the label's visibility. The todo's row (`<li>`) must
be displayed, but its label text is used as the key even when the label itself is hidden
— as it is during editing, where `dblclick "buy milk"` then `focus is in the edit field
of "buy milk"` must still resolve. A todo's identity does not vanish because it is being
edited.

## Check registry (`THEN check:`)

Escape hatch for assertions the projection can't express. Closed set; extend
deliberately.

| Check | Meaning |
|---|---|
| `focus is on the new-todo input` | document focus is on the new-todo input |
| `focus is in the edit field of "title"` | document focus is in that todo's edit field |
| `destroy button of "title" is visible` | that todo's destroy button is visible |
| `destroy button of "title" is hidden` | that todo's destroy button is not visible |

## Runner obligations

- Discover and execute every scenario in every `evals/scenarios/*.md` file.
- Project the page per § Page projection, and satisfy `evals/projection/`: the
  projector's output for each fixture must equal the paired projection. Those fixtures
  pin the projector's raw output, so their item lines carry each row's data-id — the
  form this list's `THEN page:` comparison produces *before* it erases unmatched ids.
- Compare `THEN page:` by exact line-by-line diff of the projection; on failure, print
  expected and actual projections side by side. Render actual item lines with their
  data-id, then erase ids from actual lines whose expected counterpart carries none,
  so the diff stays exact while ids remain opt-in per line.
- Compare `THEN model:` by title, completed state, and order. For a line carrying a
  `#name`, resolve the name to its binding: if the name was already bound (e.g. seeded
  in `GIVEN model:`), the persisted item's id must equal the bound id; if this is the
  name's first occurrence, bind it to the persisted item's id.
- Enforce the always-on structural invariants after **every** scenario, whatever THEN
  sections it carries — a scenario with only `THEN check:` and no `THEN model:` still
  gets them; the runner reads the model for this even when nothing asked it to:
  - The new-todo input exists and is rendered visible. The spec keeps the header visible
    in every state, so its absence is always a bug — the runner fails with a clear
    message rather than projecting a placeholder line.
  - Every persisted item has a non-empty string id, and every displayed row's data-id
    equals the id of the persisted item bearing the same title — not merely *some*
    persisted id, so swapped ids between two rows fail. A displayed row with no data-id,
    or with no persisted counterpart, fails too. (Well-defined because titles are unique
    per scenario.)
- Evaluate every `THEN` section as eventually consistent — poll to a deadline rather
  than sampling once — per HARNESS.md § Settling. This covers reading the model after
  the app has had the chance to persist.
- The runner and adapters are disposable. This file and `evals/scenarios/` are not.
