# Working Discipline

Standing corrections. Each is here because it recurred.

## Scoping

- **One question at a time.** Present one claim-first question, wait, then the next. Never a numbered list
  of three. Applies to formal alignment, informal clarification, and scope-narrowing alike.
- **Smallest change that delivers the ask.** Correctness is not the bar; reviewability is. A PR must answer
  "why these exact changes, together?" — if it cannot, split it.
- **Split before sweeping.** When scope grows mid-task, stop and propose the split rather than folding in
  "while I'm here" fixes. Track adjacent improvements as follow-ups.
- **Edit only what was flagged.** Do not restructure surrounding content that nobody asked about.
- **Weigh intent and leverage, not usage count.** Whether to build a primitive is not decided by how many
  callers exist today.

## Executing

- **Root cause, not tolerance.** No skip-lists, no special-case exceptions to make validation pass. Fix the
  cause or redefine the domain.
- **Read the siblings.** Before adding a page, component, or service beside existing ones, read two of them
  and match their guard and conditional logic.
- **Screenshots are ground truth.** When given one, read it. Do not dispatch agents to rediscover what is
  already visible.

## Verifying

- **Every commit, not the tip.** Loop the fast checks over the whole revset before pushing a restructured
  stack. The recurring defect is a commit whose test asserts a value a *later* commit introduces — it
  survives every tip-only check, and a bisect lands on a broken tree.
- **Mutation-test the guard.** After writing a test for a mechanism, delete the mechanism and confirm the
  test fails. Restore in the same step, never a later one. A green test is not evidence it guards anything.
- **Match CI exactly.** Run CI's own whole-repo command, not a filtered subset — a `*.rb`-only lint misses
  `bin/` scripts.
- **Rename sweeps need three passes:** case-insensitive, line-wrap-aware, and identifier-vs-prose. Any one
  alone produces a false "zero occurrences".
- **Cohesion pass after multi-file edits.** Review the whole artifact against its source, not just the
  changed files — that is what catches stale cross-refs and vocabulary drift.
- **Verify light.** One or two cold subagents plus a self-check, not a fan-out workflow.
- **Tell subagents to deliver.** A review subagent idles without sending. Put the SendMessage-to-parent
  instruction in its prompt, and never read silence as "it found nothing".

## Reporting

- **Show the artifact at every gate.** At a sign-off or decision point, paste the actual content or open the
  rendered page. A prose recap gives nothing to react to.
- **Ground every claim in a source.** Never promote a flagged unknown to a stated fact.
- **Never prescribe access the user lacks.** Validation steps must use repo-native observability and
  observable effects, not a production console.
- **Answer confirmations directly.** Yes or no, then stop. No "did you remember…" checklist.

**Why:** each line above was given as feedback more than once. They are collected here rather than
rediscovered per session — see [no-memory-files.md](no-memory-files.md) for why that matters.
