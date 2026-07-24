# Item

Covers `spec/spec.md` § "Item" (toggle, destroy, hover). The editing interaction
is covered in `editing.md`.

## Unchecking a completed todo reactivates it

GIVEN model:

    [x] buy milk
    [ ] walk the dog

WHEN:

    click toggle of "buy milk"

THEN page:

    v > ""
    [ ] buy milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy milk
    [ ] walk the dog

NOTE: the disappearance of `| [Clear completed]` from the footer line also pins
that the button hides when the last completed todo is reactivated.

## The destroy button removes its todo

GIVEN model:

    #a1 [ ] buy milk
    #b2 [ ] walk the dog

WHEN:

    click destroy of "buy milk"

THEN page:

    v > ""
    #b2 [ ] walk the dog
    -- **1** item left | (All) Active Completed

THEN model:

    #b2 [ ] walk the dog

NOTE: the ids pin that the right item was removed and the survivor kept its
identity — asserted both in the persisted model and on the rendered row's
data-id attribute.

## Destroying the last todo hides main and footer

GIVEN model:

    [ ] buy milk

WHEN:

    click destroy of "buy milk"

THEN page:

    > ""

THEN model:

    (empty)

## The destroy button appears on hover

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    hover "buy milk"

THEN check:

    destroy button of "buy milk" is visible
    destroy button of "walk the dog" is hidden
