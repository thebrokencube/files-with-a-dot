# PR Descriptions

House style for PR bodies. Applies to all repos, git and jj. Written for a human reviewer skimming the PR — concise, list-forward, and honest about what the PR does *now*.

## Template vs house style — ask first

Before writing a body, check for a repo PR template (`.github/pull_request_template.md` or `PULL_REQUEST_TEMPLATE.md`, or `.github/PULL_REQUEST_TEMPLATE/`).

- **No template** → use the house style below.
- **Template exists** → **ask the user**: use the repo template, or the house style? Don't silently pick either.
- **Settled per-repo preference wins over the ask.** If a project CLAUDE.md or a memory records a decision (e.g. "zenpayroll → use its `.github/pull_request_template.md`"), follow it without asking.

## Per-repo shapes (settled — do not re-ask)

| Repo | Body | Title |
|---|---|---|
| `guideline-app/groot` | `## What` · `## Why` · `## How` · `## Notes` · `## Screenshots / Demos` (placeholder for the user to fill). No `TL;DR`. No template file exists — read a recent non-dependabot PR to confirm, since older ones vary. | conventional commit |
| `zenpayroll` | its `.github/pull_request_template.md`: *What is this change doing?* · *Why is this change being made?* · *How did you test this change?* with Happy/Sad-Path checkboxes · *Related documentation:* with `Ticket:` + `Tech Spec:` | conventional commit |
| `guideline-app/app`, `guideline-app/radr` | the repo's `PULL_REQUEST_TEMPLATE.md`, with structured fields wrapped in `[[[` / `]]]` — the wrapper is required and has been missed on four PRs | `[TICKET-123] type(scope): description` — ticket key **in brackets** |

**Never add an `## AI-Assisted` section, to any repo, ever.** Rejected twice in groot and skipped in
zenpayroll's template. Finding it on an older PR is not licence to include it. Where it would have gone,
`## Screenshots / Demos` goes instead. This extends to any AI-attribution footer.

## House style

**Title** — conventional-commit style, same as a commit subject: `type(scope): short description`. The
description says what changes for a reader, not where the bytes live. Don't lead with storage or internal
vocabulary (table and column names, class names, queue names, "unify the write path") when a behavioural
phrasing exists.

**Body** — these sections, in order. Drop any that don't earn their place; match fidelity to the change (a one-line fix doesn't need all four).

- `## What` — the point first, then concise **list-forward** bullets (see One thought per line). Bold the key concepts/terms. Describe what the PR does **now** — no history narration ("originally bundled with…", "split out of #X", "as discussed"). A reviewer doesn't care how it got here.
- `## How` — **optional, and only when a diagram teaches the core concept** (a data flow, a seam, a state path). Less is more: skip it for mechanical or pure-fix PRs. See Diagrams below.
- `## Notes` — terse bullets: caveats, follow-ups, how it was validated ("green in isolation: rubocop + typecheck + tests"), links.
- `## The stack` — only for stacked PRs. See below.

## Depth order

Answer in this order. A reviewer should be able to stop after the first two and still know what landed.

1. **Outcome.** One or two lines: what is true once this merges.
2. **Deploy impact.** Answer "does anything change for a human on deploy?" in the first few lines, including
   when the answer is "nothing". Never leave it to be discovered in `## Notes`.
3. **Vocabulary.** Define every domain noun the body leans on, before its second use. A reviewer who does
   not know what the nouns mean cannot evaluate the change.
4. **Mechanics.** How it works, list-forward.
5. **In-the-weeds detail.** Per-file walkthroughs, enumerated states, migration steps, CSS traps: put them in
   a collapsed block rather than deleting them.

   ```markdown
   <details><summary>Migration steps</summary>

   …

   </details>
   ```

## One thought per line — a joiner is the tell

**If a sentence joins two ideas with an em-dash, a comma, a semicolon, or a middot (`·`), make it a list.**
Reaching for a joiner means you wrote a list as prose. There is no threshold: two ideas is enough.

This applies to every line of the body, not just bullets:

- A run-on bullet splits into sub-bullets under a lead-in.
- A run-on sentence becomes bullets.
- Run-on structure hides the shape.
- Nesting shows it.

Before (one run-on bullet):

- New `foo.md` — the house style: `type(scope):` title; `## What` list-forward; optional `## How` diagram; terse `## Notes`.

After (a lead-in + sub-bullets):

- New `foo.md` — the house style:
  - `type(scope):` title
  - `## What` — list-forward
  - `## How` — optional diagram
  - `## Notes` — terse

## Diagrams

One simple diagram beats three. When you include one:

- Use a mermaid ```mermaid ``` block, `flowchart LR` (or `TD`).
- **Quote every node label**: `A["web (Inertia)"]`. Keep labels plain — avoid a leading `#`, middots (`·`), and `<br/>`; they can stress GitHub's renderer (which sometimes shows "Loading chunk failed" — usually a transient client-side error, but simple syntax avoids feeding it).
- Teach the **concept**, not the file list — the boxes should be the ideas a reviewer needs, not the classes you touched.

## Stacked PRs

Put a `## The stack` list at the **bottom of every PR in the stack** so a reviewer can navigate from any one:

```markdown
## The stack

Ships bottom-up (each PR's base is the one below it):

1. #291 — foundation: data model + spine
2. #293 — surfaces: pages + dev tooling
3. #294 — reviews UI: board interactions
4. #298 — authz: capability model (stack tip)

_This is PR #294._
```

- Numbered bottom → top, one line each: `#NNN — chunk: what it does`. GitHub auto-links the `#NNN`.
- Mark the current PR (`_This is PR #NNN._`).
- Keep an umbrella/overview PR (if any) as the "start here" — high-level What + one diagram + this same stack list as its centerpiece; point reviewers at the stack, not the umbrella diff.

For how to create and re-base stacked PRs (bases, drafts, propagation), see `git-stacking.md` / `jj-stacking.md`.
