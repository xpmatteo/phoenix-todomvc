# Editing

Covers `spec/spec.md` § "Editing" and the editing interaction of § "Item".

## Double-clicking a label enters editing mode

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    dblclick "buy milk"

THEN page:

    v > ""
    [edit: buy milk]
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy milk
    [ ] walk the dog

THEN check:

    focus is in the edit field of "buy milk"

NOTE: the `[edit: …]` rendering asserts that the item's normal controls are hidden
while editing (DSL § "Page projection"), and the unchanged THEN model asserts that
editing mode itself is not persisted (spec § "Persistence").

## Enter saves the edit

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    dblclick "buy milk"
    clear
    type "buy oat milk"
    press Enter

THEN page:

    v > ""
    [ ] buy oat milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy oat milk
    [ ] walk the dog

## Editing preserves the item's identity

GIVEN model:

    #a1 [ ] buy milk
    #b2 [ ] walk the dog

WHEN:

    dblclick "buy milk"
    clear
    type "buy oat milk"
    press Enter

THEN model:

    #a1 [ ] buy oat milk
    #b2 [ ] walk the dog

NOTE: guards against implementations that "edit" by destroying and recreating the
item — the title changes but the id must not.

## Blur saves the edit

GIVEN model:

    [ ] buy milk

WHEN:

    dblclick "buy milk"
    clear
    type "buy oat milk"
    blur

THEN model:

    [ ] buy oat milk

## The edited title is trimmed

GIVEN model:

    [ ] buy milk

WHEN:

    dblclick "buy milk"
    clear
    type "   buy oat milk   "
    press Enter

THEN model:

    [ ] buy oat milk

## Saving an empty edit destroys the todo

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    dblclick "buy milk"
    clear
    press Enter

THEN page:

    v > ""
    [ ] walk the dog
    -- **1** item left | (All) Active Completed

THEN model:

    [ ] walk the dog

## Escape cancels the edit and discards changes

GIVEN model:

    [ ] buy milk
    [ ] walk the dog

WHEN:

    dblclick "buy milk"
    clear
    type "something else entirely"
    press Escape

THEN page:

    v > ""
    [ ] buy milk
    [ ] walk the dog
    -- **2** items left | (All) Active Completed

THEN model:

    [ ] buy milk
    [ ] walk the dog
