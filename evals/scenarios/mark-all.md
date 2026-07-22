# Mark all as complete

Covers `spec/spec.md` § "Mark all as complete".

## Clicking the toggle completes every todo

GIVEN model:

    [ ] buy milk
    [x] walk the dog
    [ ] feed the cat

WHEN:

    click toggle-all

THEN page:

    (v) >
    [x] ~buy milk~
    [x] ~walk the dog~
    [x] ~feed the cat~
    -- **0** items left | (All) Active Completed | [Clear completed]

THEN model:

    [x] buy milk
    [x] walk the dog
    [x] feed the cat

NOTE: mixed starting state on purpose — "toggles all the todos to the same state
as itself", so from unchecked it completes everything, including the already
completed one staying completed.

## Clicking the checked toggle reactivates every todo

GIVEN model:

    [x] buy milk
    [x] walk the dog

WHEN:

    click toggle-all

THEN page:

    v >
    [ ] buy milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy milk
    [ ] walk the dog

## Completing the last active todo checks the toggle

GIVEN model:

    [x] buy milk
    [ ] walk the dog

WHEN:

    click toggle of "walk the dog"

THEN page:

    (v) >
    [x] ~buy milk~
    [x] ~walk the dog~
    -- **0** items left | (All) Active Completed | [Clear completed]

## Reactivating one todo unchecks the toggle

GIVEN model:

    [x] buy milk
    [x] walk the dog

WHEN:

    click toggle of "buy milk"

THEN page:

    v >
    [ ] buy milk
    [x] ~walk the dog~
    -- **1** item left | (All) Active Completed | [Clear completed]

## The toggle does not stay checked after Clear completed

GIVEN model:

    [x] buy milk

WHEN:

    click "Clear completed"
    type "walk the dog"
    press Enter

THEN page:

    v >
    [ ] walk the dog
    -- **1** item left | (All) Active Completed

THEN model:

    [ ] walk the dog

NOTE: guards the spec clause "clear the checked state after the Clear completed
button is clicked" — a stale checked toggle would otherwise show `(v)` when the
next todo is created.
