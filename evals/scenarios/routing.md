# Routing

Covers `docs/spec.md` § "Routing".

The default route (`#/`, all todos, `(All)` selected) is pinned by every other
scenario in this suite, since they all run without a `GIVEN route:`.

NOTE: the spec says the `#!/` route prefix "is also allowed". This suite pins the
`#/` form as required and does not exercise `#!/` — permitting an alternative is
not testable as an obligation. If we ever want `#!/` guaranteed, the spec must
promote it from "allowed" to "required" and a scenario should be added here.

## The Active filter shows only active todos

GIVEN model:

    [ ] buy milk
    [x] walk the dog

GIVEN route: #/active

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | All (Active) Completed | [Clear completed]

NOTE: the counter and the Clear-completed button reflect the whole model, not
the filtered view.

## The Completed filter shows only completed todos

GIVEN model:

    [ ] buy milk
    [x] walk the dog

GIVEN route: #/completed

THEN page:

    v >
    [x] ~walk the dog~
    -- **1** item left | All Active (Completed) | [Clear completed]

## Clicking a filter link filters the list and moves the selection

GIVEN model:

    [ ] buy milk
    [x] walk the dog

WHEN:

    click filter "Active"

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | All (Active) Completed | [Clear completed]

## Completing an item under the Active filter hides it immediately

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

GIVEN route: #/active

WHEN:

    click toggle of "buy milk"

THEN page:

    v >
    [ ] walk the dog
    -- **1** item left | All (Active) Completed | [Clear completed]

THEN model:

    [x] buy milk
    [ ] walk the dog

NOTE: this is the spec's own example — the item is completed (see THEN model)
but no longer displayed, because it no longer matches the filter.

## The active filter is persisted across a reload

GIVEN model:

    [ ] buy milk
    [x] walk the dog

WHEN:

    click filter "Completed"
    reload

THEN page:

    v >
    [x] ~walk the dog~
    -- **1** item left | All Active (Completed) | [Clear completed]
