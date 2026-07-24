# New todo

Covers `spec/spec.md` § "New todo" and § "No todos".

## The input is focused on load, and an empty app hides main and footer

GIVEN model:

    (empty)

THEN page:

    > ""

THEN check:

    focus is on the new-todo input

NOTE: the page projection being a single line asserts that both the main section
and the footer are absent (spec § "No todos").

## Pressing Enter creates the todo, appends it to the list, and clears the input

GIVEN model:

    [ ] buy milk

WHEN:

    type "walk the dog"
    press Enter

THEN page:

    v > ""
    [ ] buy milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy milk
    [ ] walk the dog

## The title is trimmed before the todo is created

GIVEN model:

    (empty)

WHEN:

    type "   buy milk   "
    press Enter

THEN model:

    [ ] buy milk

## A blank title does not create a todo

GIVEN model:

    [ ] buy milk

WHEN:

    type "      "
    press Enter

THEN model:

    [ ] buy milk

NOTE: no page assertion here — the spec does not say whether the input is cleared
after a rejected blank entry, so we only pin what the spec pins: no todo appears.
