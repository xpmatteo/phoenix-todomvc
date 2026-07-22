# Deliberately failing scenarios

Used by the runner's own tests to verify failure reporting.

## Page and model expectations that the stub cannot meet

GIVEN model:

    [ ] real thing

THEN page:

    v >
    [ ] imaginary thing
    -- **1** item left | (All) Active Completed

THEN model:

    [x] real thing
