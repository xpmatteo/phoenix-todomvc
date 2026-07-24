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
- We practise trunk-based development. Commit straight to `main`. Only work in
  a branch or worktree when the user asks for one.
- A pre-commit hook blocks staged credentials. Enable it once per clone with
  `git config core.hooksPath .githooks`; without that it silently does nothing.

# Delegating to subagents

Pick the cheapest model that will do the job well. A subagent on a mechanical
task does not need the model you would choose for the hardest design problem in
the repo. Use your judgement, and prefer a smaller model when the task is
well-specified.

**Never delegate to Fable.** It costs roughly twice Opus per token.

# Prose style

Use simple, practical, no-nonsense language. These docs are read by fresh agents
and humans who need to act on them, often under time pressure. Write for that
reader.

- Use no more words than the point needs. Cut every word that carries no
  meaning. A shorter sentence that says the same thing is always better.
- One idea per sentence. If a sentence has two em-dashes and a semicolon, it is
  probably three sentences.
- Concrete over abstract. Name the actor and the action ("the runner starts the
  app") rather than a relationship ("the boundary between runner and app").
- Explain *why* before *what*. Give the reader the purpose of a thing before the
  mechanical rules for it.
- No grand flourishes. State the finding plainly. Do not end an entry on an
  aphorism that reframes it as a great revealed truth.
