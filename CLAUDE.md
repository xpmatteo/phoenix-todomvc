# Working in this repository

Read `README.md` first: it defines what is durable vs disposable here, and the
boundary rule deciding whether a requirement belongs in `spec/spec.md`,
`spec/architecture.md`, or `evals/HARNESS.md`. Follow it strictly.

Hard rules:

- `app/` is disposable, generated output. Never hand-edit it to fix a problem;
  fix the durable artifacts and regenerate.
- Never weaken or delete a scenario in `evals/scenarios/` to make an
  implementation pass.
- When docs prove ambiguous during a build, record the gap explicitly (and fix
  the doc) rather than silently choosing.
- Commit messages record *why* — the reason for the change, not a diff summary.

# Prose style

Use simple, practical, no-nonsense language. These docs are read by fresh agents
and humans who need to act on them, often under time pressure. Write for that
reader.

- One idea per sentence. If a sentence has two em-dashes and a semicolon, it is
  probably three sentences.
- Concrete over abstract. Name the actor and the action ("the runner starts the
  app") rather than a relationship ("the boundary between runner and app").
- Explain *why* before *what*. Give the reader the purpose of a thing before the
  mechanical rules for it.
- Bold is for the one thing a skimmer must not miss. If four things per page are
  bold, none of them are.
- Don't over-constrain. State the actual invariant, not one accidental way of
  meeting it — "the app exposes no web route to seed" not "never use a network
  listener," which needlessly bans Postgres.
- Prefer the register of the numbered steps in `evals/HARNESS.md`: short,
  ordered, imperative.
