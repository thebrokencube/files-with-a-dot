---
name: stacked-pr
description: Stacked branch workflows -- propagation, fixup targeting, generated commits, reset+cherry-pick mechanics. Use when working with branch chains, stacked PRs, propagating changes across dependent branches, or resolving conflicts across branch stacks.
user_invocable: false
---

# Stacked PR Workflows

Rules for managing chains of dependent branches (stacked PRs). Covers propagation ordering, fixup targeting, generated commits, and conflict resolution.

## Propagation Workflow

When a branch in the stack is rewritten (rebase, fixup, amend), all descendants must be propagated. **Always propagate in parent-first DAG order.**

> **Override (commit skill, Rebase Workflow):** During propagation, the "ask before rebasing" step does not apply. Propagation is mechanical — propagate each child without pausing for confirmation. Conflict-category rules below still apply.

### Pre-flight

1. Stash or commit working changes on every branch in the stack
2. Verify the parent branch is clean and fully rebased
3. Record each child's unique commits BEFORE rewriting any branch:

```bash
COMMITS_B=$(git rev-list --reverse feature-a..feature-b)
COMMITS_C=$(git rev-list --reverse feature-b..feature-c)
```

Commit lists MUST be recorded before rewriting because branch ranges change meaning after a parent is rebased.

### Propagate

Rewrite the parent branch first, then propagate each child via reset + cherry-pick:

```bash
# Step 1: Rewrite feature-a (rebase, fixup, etc.)
git checkout feature-a
git rebase origin/main   # or whatever the rewrite is

# Step 2: Propagate feature-b onto the new feature-a
git checkout feature-b
git reset --hard feature-a
git cherry-pick $COMMITS_B

# Step 3: Propagate feature-c onto the new feature-b
git checkout feature-c
git reset --hard feature-b
git cherry-pick $COMMITS_C
```

The pattern is always: reset to the new parent tip, then cherry-pick the child's commits.

> **Why not `rebase --onto`?** Both approaches use cherry-pick internally and have identical conflict behavior. But `reset + cherry-pick` records explicit commit SHAs instead of abstract reference points, which eliminates the "wrong old-tip" failure mode and provides a simpler mental model: "these are my commits, put them on this base."

### Green Commits During Propagation

Every commit MUST remain green after propagation. If a rebase introduces a failure in any commit, fix it and autosquash into that commit — NEVER leave a broken commit in the chain or push a fix as a separate commit.

### Post-flight

1. Verify commit count on each branch matches expectations
2. Run project-specific checks (tests, lint) on the tip of each branch
3. Push all branches (see Push Strategy)

If something goes wrong, see Common Failure Modes at the end of this skill.

### Propagation Checklist

- [ ] All branches have clean working trees
- [ ] Commit lists recorded for every child branch (BEFORE rewriting)
- [ ] Parent rewritten first
- [ ] Children propagated in DAG order (root-to-leaf)
- [ ] Generated commits handled (omit from cherry-pick, re-run after)
- [ ] Every commit green on every branch
- [ ] All branches pushed

## Fixup Targeting

Before creating a fixup commit, verify the target commit is on the current branch and is the correct one.

### Pre-check

```bash
# Which commit introduced the file?
git log --oneline --diff-filter=A -- path/to/file.rb

# Which commit last touched these lines?
git log --oneline -1 -- path/to/file.rb

# Verify the target commit modifies the expected file
git show --stat <target-sha> | grep path/to/file.rb

# Confirm target is on the current branch
git log --oneline main..HEAD -- path/to/file.rb
```

### Apply

```bash
git commit --fixup=<target-sha>
GIT_SEQUENCE_EDITOR=: git rebase -i --autosquash <target-sha>~1
```

`GIT_SEQUENCE_EDITOR=:` makes autosquash non-interactive — it accepts the reordered todo list without opening an editor.

### Rules

- **If the target commit is not on the current branch**, switch to the branch that owns it before creating the fixup. NEVER fixup a commit on a different branch from where you are.
- **NEVER fixup a generated commit** (`auto:` prefix). Drop and re-run instead (see Generated Commits).
- After autosquash, propagate all descendant branches.

## Generated Commits

Commits prefixed `auto:` are deterministic outputs of commands. The commit description **is the literal command** to reproduce the content. This makes regeneration self-documenting — the command to re-run is right in `git log`.

### Convention

The description after `auto(<scope>):` is the exact command to run:

```
auto(codegen): graphqlme
auto(deps): bundle install
auto(migrations): bin/rails db:migrate
```

### Detection

```bash
git log --oneline main..HEAD --grep="^auto"
```

### Drop-and-Rerun Pattern

During propagation, generated commits should be omitted from the cherry-pick list and re-run after:

1. Before cherry-picking, identify generated commits: `git log --oneline main..HEAD --grep="^auto"`
2. Exclude their SHAs from the cherry-pick list
3. After cherry-picking the non-generated commits, read each generated commit's message to get the command: `git log --format=%s -1 <sha>`
4. Run the command from the commit description
5. Recommit with the same `auto:` message

**NEVER manually edit generated commit content.** The generator is the source of truth.

## Conflict Resolution

Three categories with different handling. **These rules apply even during mechanical propagation** — the propagation override only skips the rebase confirmation, not conflict judgment.

### Mechanical Conflicts

Whitespace changes and trivially resolvable merge markers. Does NOT include import reordering — reordering imports can have side effects in some languages (Ruby `require`, Python import-time effects).

**Action**: Resolve and continue.

### Semantic Conflicts

Logic changes from a parent branch affect code in a child branch. The child's assumptions about behavior may no longer hold.

**Action**: Pause and confirm with the user. Describe what changed in the parent and what the child assumes.

### Generated Conflicts

Conflicts in `auto:` commits — lockfiles, codegen output, schema files.

**Action**: Drop the commit and re-run the command from its description. NEVER merge generated content manually.

## Push Strategy

- Follow commit skill's force-push rules (`--force-with-lease`, NEVER force push main)
- **Push parent before children** (same DAG order as propagation)
- **Push one branch at a time**, in DAG order (parent before children). Some repos enforce per-push branch limits, and serial pushes avoid hitting them:

```bash
git push --force-with-lease origin feature-a
git push --force-with-lease origin feature-b
git push --force-with-lease origin feature-c
```

### Rejected Force Push

If `--force-with-lease` is rejected, someone else pushed to the branch. Do not override:

1. `git fetch origin`
2. Inspect what changed: `git log feature-a..origin/feature-a`
3. Decide with the user whether to integrate or overwrite

If parent branches already pushed but a child is rejected, the parent pushes are fine — only the child needs resolution.

### Post-Merge Stack Advancement

When a root PR merges into main, advance the stack:

1. `git checkout <next-branch> && git rebase origin/main`
2. This is mechanical — content should not change since the branch was already based on the merged parent
3. If you encounter an `auto:` commit during the rebase, determine whether it needs regeneration. Codegen that depends on the rebased code (e.g., GraphQL types after schema changes) must be dropped and re-run. Codegen that is independent of the change (e.g., a lockfile from a different branch's dep addition) can rebase through cleanly.
4. Push with `--force-with-lease` and repeat for remaining children

### Stack Presentation

When creating PRs for a stack:
- Include a stack overview in each PR description showing the branch order and which PR is current
- Mark child PRs as draft until their parent is approved
- After propagation, update PR descriptions if the diff changed materially

## Decomposition

Ordering principle: foundational changes first, then implementation, then integration, then consumer-facing.

- **Within a branch**: refactor → implement → wire → UI/API
- **Within a stack**: root branch absorbs structural risk; children make the now-easy changes
- **Each branch MUST be independently reviewable and mergeable**
- Stack when total diff > ~400 lines with natural layers; single branch otherwise

Stacks assume linear (rebase) history. NEVER use merge commits within a stack.

## Common Failure Modes

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Fixup landed on wrong branch | Didn't verify target commit ownership | `git cherry-pick <sha>` to correct branch, `git rebase -i` to drop from wrong one |
| Cherry-pick conflict mid-propagation | Semantic conflict between parent and child changes | Resolve conflict, `git cherry-pick --continue`. Recorded commit list still valid for remaining picks |
| Generated commit has merge conflicts | Cherry-picked instead of omit-and-rerun | `git cherry-pick --abort`, omit the generated SHA, re-run command after remaining picks |
| Child propagated before parent | Propagated in wrong order | `git reflog show <branch>` to restore, re-record commit lists, re-propagate in correct DAG order |
| `--force-with-lease` rejected | Remote updated by someone else | Fetch, inspect remote changes, decide with user |
| `--force-with-lease` rejected with "stale info" | Branch was recreated via `checkout -B` during propagation | `git fetch origin <branch>` then retry. If remote ref is gone (new branch), use `--force` |
| Forgot to record commits before rewriting | Branch ranges invalid after parent rebase | `git reflog show <child>` to find pre-rewrite tip, then `git rev-list --reverse <old-parent-tip>..<old-child-tip>` to recover the commit list |
| Two commits merged into one after rebase | Used `git commit --amend` at a **conflict** stop instead of `git rebase --continue` | See Edit vs Conflict Stops below. Recover: `git rebase --abort`, re-run rebase |
| Stack parent ambiguous | Fork in DAG, unclear which branch is parent | **Ask the user.** NEVER guess from commit dates or branch names |

## Edit vs Conflict Stops During Rebase

**CRITICAL**: `edit` stops and conflict stops during `git rebase -i` require different workflows. Mixing them up silently merges commits.

### Edit stop (you requested `edit` in the todo list)

The commit HAS been applied. You're amending it.

```bash
# Make changes to files
git add <files>
git commit --amend --no-edit
git rebase --continue
```

### Conflict stop (rebase hit a merge conflict)

The commit has NOT been created yet. `git rebase --continue` will create it.

```bash
# Resolve conflicts
git add <files>
git rebase --continue        # creates the commit
```

**NEVER `git commit --amend` at a conflict stop.** It amends the *previous* commit, silently merging two commits into one. This is extremely hard to notice and painful to recover from.

### How to tell which stop you're at

- **Edit stop**: message says "Stopped at <sha>... You can amend the commit now"
- **Conflict stop**: message says "CONFLICT" and files have merge markers

## Retroactive Edits Across a Stack

When you need to apply a change retroactively across multiple commits (e.g. renaming a pattern introduced in commit 8 and used in commits 10, 14, 17), a single `rebase -i` with targeted `edit` stops is the fastest approach:

```bash
# Mark specific commits as edit using sed on the todo list
GIT_SEQUENCE_EDITOR="sed -i '' -e '/^pick <sha1>/s/^pick/edit/' -e '/^pick <sha2>/s/^pick/edit/'" git rebase -i <base>
```

At each stop: make the change, `git add`, `git commit --amend --no-edit`, `git rebase --continue`.

After the rebase completes, update child branch pointers (`git branch -f <child> <new-tip>`) and propagate descendants.

## Folio Integration

When a `folio.yml` exists, use it as the source of truth for branch topology instead of parsing PR descriptions or relying on mental models.

### Topology Discovery

```bash
folio dag --branches --json --folio <path-to-folio.yml>
```

Returns a `BranchTopology` JSON structure with nested roots → children. Each node has `id` (target name), `branch`, `base`, and `pr` fields. This replaces manual branch-parent tracking.

### Propagation Order

The JSON tree encodes propagation order implicitly: pre-order traversal (root-to-leaf) gives the correct order. Walk each root's children depth-first — parent branches always appear before their children.

### Post-Propagation

After propagation completes, check for stale composition targets:

```bash
folio stale --json --folio <path-to-folio.yml>
```

Stale entries include a `branch` field for direct mapping. If stale targets exist, suggest `/folio compose <target>` for recomposition.

### Fallback

When no `folio.yml` exists (or targets lack `branch` fields), fall back to manual branch-parent tracking: branch names in PR descriptions, user confirmation for ambiguous parent relationships.

## Terminology

Quick reference for terms used throughout this skill:

- **Commit list**: The SHAs of a child branch's unique commits (`git rev-list --reverse parent..child`), recorded before rewriting
- **Generated commit**: Deterministic command output, prefixed `auto:` — description is the command
- **Stack**: Chain of branches where each depends on the previous; parent relationships tracked in folio.yml `branch`/`blocked_by` fields, or PR descriptions when no folio.yml
