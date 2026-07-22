# Clear completed

Covers `docs/spec.md` § "Clear completed button". Visibility of the button is
also pinned incidentally throughout the suite by the presence or absence of
`| [Clear completed]` on footer lines.

## Removes all completed todos, keeps active ones

GIVEN model:

    [ ] buy milk
    [x] walk the dog
    [x] feed the cat

WHEN:

    click "Clear completed"

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | (All) Active Completed

THEN model:

    [ ] buy milk

NOTE: the footer line also asserts that the button hides itself once no
completed todos remain.

## Clearing when everything is completed empties the app

GIVEN model:

    [x] buy milk
    [x] walk the dog

WHEN:

    click "Clear completed"

THEN page:

    >

THEN model:

    (empty)
