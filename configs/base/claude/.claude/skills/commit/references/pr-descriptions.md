# PR Descriptions

House style for PR bodies. Applies to all repos, git and jj. Written for a human reviewer skimming the PR — concise, list-forward, and honest about what the PR does *now*.

## Template vs house style — ask first

Before writing a body, check for a repo PR template (`.github/pull_request_template.md` or `PULL_REQUEST_TEMPLATE.md`, or `.github/PULL_REQUEST_TEMPLATE/`).

- **No template** → use the house style below.
- **Template exists** → **ask the user**: use the repo template, or the house style? Don't silently pick either.
- **Settled per-repo preference wins over the ask.** If a project CLAUDE.md or a memory records a decision (e.g. "zenpayroll → use its `.github/pull_request_template.md`"), follow it without asking.

## House style

**Title** — conventional-commit style, same as a commit subject: `type(scope): short description`.

**Body** — these sections, in order. Drop any that don't earn their place; match fidelity to the change (a one-line fix doesn't need all four).

- `## What` — the point first, then concise **list-forward** bullets (see One thought per bullet). Bold the key concepts/terms. Describe what the PR does **now** — no history narration ("originally bundled with…", "split out of #X", "as discussed"). A reviewer doesn't care how it got here.
- `## How` — **optional, and only when a diagram teaches the core concept** (a data flow, a seam, a state path). Less is more: skip it for mechanical or pure-fix PRs. See Diagrams below.
- `## Notes` — terse bullets: caveats, follow-ups, how it was validated ("green in isolation: rubocop + typecheck + tests"), links.
- `## The stack` — only for stacked PRs. See below.

## One thought per bullet

A bullet holds one thought. **If you reach for a semicolon, a middot (`·`), or an em-dash (`—`) to pack more items into one bullet, that's the signal to split** — each item becomes its own bullet or a sub-bullet under a lead-in. Run-on bullets hide structure; nesting shows it. This applies to `## What` and `## Notes` alike.

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
