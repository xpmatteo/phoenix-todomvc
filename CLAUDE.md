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
