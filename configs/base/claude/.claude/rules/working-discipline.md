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
  Then exercise the guard through its real caller — tested in isolation it is not evidence the guard holds
  in situ, and ~12k passing assertions once shipped a hole that one real call found immediately.
- **Match CI exactly.** Run CI's own whole-repo command, not a filtered subset — a `*.rb`-only lint misses
  `bin/` scripts.
- **Rename sweeps need three passes:** case-insensitive, line-wrap-aware, and identifier-vs-prose. Any one
  alone produces a false "zero occurrences".
- **Cohesion pass after multi-file edits.** Review the whole artifact against its source, not just the
  changed files — that is what catches stale cross-refs and vocabulary drift.
- **Verify it yourself.** Read the diff, run the check. Never spawn a subagent to re-check correctness —
  Opus 5 self-verifies, so a verification agent buys nothing and costs a full context. Delegate for
  genuinely independent, sizeable tracks; if one agent can do it, use one.
- **A cold read is not a re-check.** Its value is the context the reader lacks, which no amount of
  self-verification produces — you cannot un-know your own premise. Spend one fresh agent when the risk is
  your context rather than your diligence: a premise you never thought to question, prior art you did not
  know to look for, an artifact about to be signed off and built on. Give it one specific question, require
  `file:line` or drop the finding, and confirm each survivor yourself before relaying it. A cold agent that
  is confidently wrong costs more than no agent.
- **A sign-off gate is standing authorisation for one cold read.** Do not ask permission for it — a gate
  the user is about to approve is the moment the author's blind spot is most expensive. One agent, on the
  gated artifact, before it is shown. Everything outside a gate still needs asking.
- **Tell subagents to deliver.** A review subagent idles without sending. Put the SendMessage-to-parent
  instruction in its prompt, and never read silence as "it found nothing".
- **State the output contract in the prompt.** Every delegation names the return shape and a length cap —
  "a table of file:line and finding, nothing else", "under 10 lines". An output style never reaches a
  subagent, and the built-in `Explore` and `Plan` skip the rules files as well, so the prompt is the only
  surface that constrains what comes back. Unbounded returns are what the lead then relays at length.

## Reporting

- **Show the artifact at every gate.** At a sign-off or decision point, paste the actual content or open the
  rendered page. A prose recap gives nothing to react to.
- **Ground every claim in a source.** Never promote a flagged unknown to a stated fact.
- **Never prescribe access the user lacks.** Validation steps must use repo-native observability and
  observable effects, not a production console.
- **Answer confirmations directly.** Yes or no, then stop. No "did you remember…" checklist.

**Why:** each line above was given as feedback more than once. They are collected here rather than
rediscovered per session — see [no-memory-files.md](no-memory-files.md) for why that matters.
