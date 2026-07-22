# Schema — durable, append-only

Status: **durable artifact** (see AD-7 in `docs/architecture.md`). The database
outlives every generated implementation; this directory is the provenance of its
shape.

- Migrations are goose-format SQL files, named `NNN_description.sql`, applied in
  order at app startup (AD-6).
- **Append-only**: new migrations are added at the end; existing files are never
  edited, reordered, or deleted. Changing a past migration is falsifying history —
  the fix for a bad migration is another migration.
- A regenerated implementation may propose appending a migration; the append is a
  durable-artifact change and is reviewed as such. It must never fork or restart
  the chain.

## Design notes

`todos.seq` (autoincrement) is the display order: the spec orders todos by
creation, and SQLite rowids may be reused after deletes, so ordering relies on an
explicit monotonic column, never on rowid. `id` is the contract identifier from
`docs/spec.md` (opaque non-empty string, stable for the item's lifetime, exposed
as `data-id`); `seq` is internal and never leaves the database.
