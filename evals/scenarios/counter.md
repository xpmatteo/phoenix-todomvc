# Counter

Covers `docs/spec.md` § "Counter".

## Zero active todos reads "0 items left"

GIVEN model:

    [x] buy milk

THEN page:

    (v) >
    [x] ~buy milk~
    -- **0** items left | (All) Active Completed | [Clear completed]

NOTE: this scenario also pins two neighbors: the mark-all checkbox is checked when
every todo is completed (spec § "Mark all as complete"), and the Clear-completed
button is visible when completed todos exist.

## One active todo reads "1 item left" (singular)

GIVEN model:

    [ ] buy milk

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | (All) Active Completed

## Two active todos read "2 items left" (plural)

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

THEN page:

    v >
    [ ] buy milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

## Only active todos are counted

GIVEN model:

    [ ] buy milk
    [x] walk the dog
    [ ] feed the cat

THEN page:

    v >
    [ ] buy milk
    [x] ~walk the dog~
    [ ] feed the cat
    -- **2** items left | (All) Active Completed | [Clear completed]

## The counter updates when a todo is toggled

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    click toggle of "buy milk"

THEN page:

    v >
    [x] ~buy milk~
    [ ] walk the dog
    -- **1** item left | (All) Active Completed | [Clear completed]

THEN model:

    [x] buy milk
    [ ] walk the dog
