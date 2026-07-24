# Projection acceptance tests

Status: **durable artifact**. These fixtures pin the DOM→projection mapping
defined in `../DSL.md` § Page projection. Each case pairs a fixed HTML fragment,
deduced from `spec/main-screen-template.html`, with the exact projection a
correct projector must produce from it.

## Why they exist

The projector turns a rendered page into the ASCII projection that `THEN page:`
diffs against. It lives in the disposable runner, so it is regenerated with the
runner and can drift from the rules in `DSL.md`. These tests catch that drift
without a running app: feed each fixture to the projector and compare its output
to the expected projection. A projector that passes every case here implements
the mapping correctly.

They stand to the projector as `../scenarios/` stands to the app: executable
spec-by-example. They are durable; the projector is not.

## Format

Each case is one [txtar](https://pkg.go.dev/golang.org/x/tools/txtar) file with
two members:

- `dom.html` — the fixture: the rendered `.todoapp` subtree the projector reads.
- `projection.txt` — the exact projection, line by line.

The text above the first `-- ` marker names the rule the case pins and says why.
Comparison is exact, line by line; a trailing newline is not significant.

## The data-id convention

These fixtures pin the projector's *raw* output. Every item line carries its
row's `data-id` as a leading `#<id>`, because the projector renders it there
(`DSL.md` § Runner obligations). A scenario's `THEN page:` omits ids only because
the comparator later erases them from lines whose expected counterpart has none.
That erasure is comparator behavior, not projection, so it is out of scope here.

So `#a1b2 [ ] buy milk` is the projection of an `<li data-id="a1b2">` with an
unchecked toggle and the label `buy milk`. The id in a fixture is fixed and
arbitrary — here it is chosen, not minted by an app.

## Scope

These pin the DOM→string mapping only. The comparator's id-erasure, the
always-on structural invariants, and the `THEN check:` visibility checks all read
more than the projection and are verified elsewhere (`DSL.md` § Runner
obligations; `HARNESS.md`).

Some fixtures are deliberately inconsistent with what a correct app would render
— a `completed` class without a checked toggle, a shown main without a footer.
The projector reads each marker locally, so these are valid inputs and they prove
the markers are independent. Each such fixture says so in its comment.

One gap the fixtures deliberately avoid: `DSL.md` does not define how a `"` inside
the quoted new-todo value is escaped in the projection, so no fixture puts one
there. Pin the rule in `DSL.md` before adding that case.
