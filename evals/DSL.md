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

- `seed(model)` — establish the given todos as the persisted state before the app loads
- `read() → model` — return the currently persisted todos

(For a client-side SPA the adapter reads/writes localStorage; for a server-rendered app
it would talk to the server's store. The adapter is disposable code, regenerated along
with the app it belongs to.)

The persisted representation carries `id`, `title`, `completed` per item (see
`docs/spec.md` § Persistence). Ids are part of the contract: they are opaque non-empty
strings, and each rendered todo item carries a `data-id` attribute equal to its
persisted id. The notation treats ids as optional — a scenario writes them only when
identity matters (see "Model notation").

**Page** — an ASCII projection of the rendered UI, defined under "Page projection" below.
Only the projected parts are compared; everything else on the page is ignored.

## Scenario files

Scenario files are markdown. Each `## heading` starts a scenario; the heading is the
scenario name. A scenario contains these sections, each introduced by a keyword line
and followed by an indented block (4 spaces):

| Section | Required | Meaning |
|---|---|---|
| `GIVEN model:` | yes | model seeded before the app loads |
| `GIVEN route:` | no | initial URL hash, written inline on the keyword line itself, e.g. `GIVEN route: #/active` (default: none) |
| `WHEN:` | no | user actions, one per line, executed in order |
| `THEN page:` | no* | expected page projection after the last action |
| `THEN model:` | no* | expected model after the last action |
| `THEN check:` | no | extra named checks (see registry below) |

*At least one `THEN` section is required.

Every scenario runs against a fresh app instance with no state other than what
`GIVEN` establishes. Scenarios are independent; order of execution is irrelevant.

Lines starting with `NOTE:` inside a scenario are commentary for humans and agents;
the runner ignores them.

Todo titles within a scenario must be unique.

## Model notation

One line per todo, in persisted order:

    [ ] title         — active todo
    [x] title         — completed todo
    #id [ ] title     — todo with an explicit id (id is any #-prefixed opaque token)
    (empty)           — no todos

Ids are optional per line:

- In `GIVEN model:`, an explicit id is seeded verbatim; when omitted, the adapter
  assigns a unique opaque id.
- In `THEN model:`, an explicit id must match the persisted item's id exactly; when
  omitted, the id is not compared (but must still be a non-empty string — the runner
  enforces this invariant on every persisted item, always).

## Page projection

The projection renders, top to bottom, only the following. A part that the app hides
(or does not render) produces **no line at all** — absence of a line asserts absence
on the page.

1. **Input row** — always first: the new-todo input rendered as `>` followed by its
   current value. When the main section is visible, the row is prefixed by the
   mark-all-complete toggle — rendered as a chevron, mirroring the UI, where parens
   mean "on" (the same convention as the selected filter):

       >               — main section hidden; empty input alone
       > buy mil       — input containing text
       v >             — toggle visible, unchecked
       (v) >           — toggle visible, checked (i.e. all todos completed)

2. **Todo items** — one line each, in displayed order, only the items currently
   displayed (a filtered-out item produces no line):

       [ ] title             — active item
       #id [ ] title         — any item line may carry a leading #id; when present,
                               the rendered item's data-id attribute must equal it
       [x] ~title~           — completed item: checkbox checked AND the item styled
                               as completed. The `[x]` reflects the checkbox, the
                               `~…~` reflects the completed styling; a buggy app can
                               produce one without the other and the diff will show it.
       [edit: value]         — item in editing mode, showing the edit field's current
                               value. Rendering this line also asserts that the item's
                               normal controls (checkbox, label, destroy button) are
                               hidden while editing.

3. **Footer** — only when the footer is visible, as a single line:

       -- <counter> | <filters> | [Clear completed]

   - `<counter>`: the counter text verbatim, with the emphasized number rendered
     markdown-style: `**2** items left`. (This projects the `<strong>` wrapper
     required by the spec.)
   - `<filters>`: the filter links in order, the selected one in parentheses:
     `(All) Active Completed`.
   - ` | [Clear completed]` appears only when the Clear-completed button is visible.

Focus is deliberately **not** part of the projection (asserting it on every line would
pin behavior the spec leaves open). Assert focus explicitly via `THEN check:`.

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
| `click toggle-all` | click the mark-all-as-complete checkbox |
| `click "Clear completed"` | click the Clear-completed button |
| `click filter "Active"` | click that filter link (All / Active / Completed) |
| `go to "#/active"` | navigate to that URL hash |
| `reload` | reload the page (persisted state must survive; in-memory state need not) |
| `hover "title"` | move the pointer over that todo |

`"title"` matches a displayed todo by its exact current label text.

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
- Compare `THEN page:` by exact line-by-line diff of the projection; on failure, print
  expected and actual projections side by side. Render actual item lines with their
  data-id, then erase ids from actual lines whose expected counterpart carries none,
  so the diff stays exact while ids remain opt-in per line.
- Compare `THEN model:` by title, completed state, and order; compare ids only where
  the expected line carries one. Independently verify that every persisted item has a
  non-empty string id and that every displayed item's data-id equals the id of the
  persisted item bearing the same title — not merely *some* persisted id, so swapped
  ids between two rows fail. (Well-defined because titles are unique per scenario.)
- Read the model via the adapter only after the app has had the chance to persist
  (the spec requires persistence immediately after every interaction).
- The runner and adapters are disposable. This file and `evals/scenarios/` are not.
