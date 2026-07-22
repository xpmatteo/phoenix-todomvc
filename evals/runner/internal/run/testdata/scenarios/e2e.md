# End-to-end fixtures for the runner itself

These scenarios run against the read-only stub app in ../stubapp, which just
renders whatever the seed command persisted. They exercise the runner's
orchestration (start/poll/seed/navigate/act/project/read), not TodoMVC
behavior.

## Seeded model is rendered, ids flow end to end

GIVEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

WHEN:

    hover "buy milk"

THEN page:

    v >
    #a1 [ ] buy milk
    [x] ~walk the dog~
    -- **1** item left | (All) Active Completed | [Clear completed]

THEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

THEN check:

    destroy button of "buy milk" is visible
    destroy button of "walk the dog" is hidden

NOTE: the second page line asserts its id, the third deliberately does not,
so this pins the erase-unrequested-ids behavior against a live browser.

## Seeding the empty model empties the page, and typing reaches the input

GIVEN model:

    (empty)

WHEN:

    type "abc"

THEN page:

    > abc

THEN model:

    (empty)

THEN check:

    focus is on the new-todo input

## Routes, filter clicks and reload traverse full page loads

GIVEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

GIVEN route: /active

WHEN:

    click filter "Completed"
    reload

THEN page:

    v >
    #b2 [x] ~walk the dog~
    -- **1** item left | All Active (Completed) | [Clear completed]

## go to navigates relative to the app URL

GIVEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

WHEN:

    go to "/active"

THEN page:

    v >
    #a1 [ ] buy milk
    -- **1** item left | All (Active) Completed | [Clear completed]

## All todos completed renders the checked toggle-all chevron

GIVEN model:

    #a1 [x] buy milk

THEN page:

    (v) >
    #a1 [x] ~buy milk~
    -- **0** items left | (All) Active Completed | [Clear completed]
