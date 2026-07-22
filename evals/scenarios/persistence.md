# Persistence

Covers `docs/spec.md` § "Persistence".

That todos are persisted *immediately after every interaction* is asserted all
over this suite: every `THEN model:` reads the persisted state right after the
last action, without a reload. The scenarios here pin the other half of the
contract: the app renders back what was persisted, and only that.

## Persisted todos survive a reload, states and identities intact

GIVEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

WHEN:

    reload

THEN page:

    v >
    [ ] buy milk
    [x] ~walk the dog~
    -- **1** item left | (All) Active Completed | [Clear completed]

THEN model:

    #a1 [ ] buy milk
    #b2 [x] walk the dog

## A todo created through the UI survives a reload

GIVEN model:

    (empty)

WHEN:

    type "buy milk"
    press Enter
    reload

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | (All) Active Completed

THEN model:

    [ ] buy milk

## Editing mode does not survive a reload

GIVEN model:

    [ ] buy milk

WHEN:

    dblclick "buy milk"
    clear
    type "half-finished edit"
    reload

THEN page:

    v >
    [ ] buy milk
    -- **1** item left | (All) Active Completed

THEN model:

    [ ] buy milk

NOTE: pins "editing mode should not be persisted" — after the reload the item is
back to its normal rendering with its original title; the in-flight edit is gone.
