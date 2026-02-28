---
name: stacked-pr
description: Stacked branch workflows -- propagation, fixup targeting, generated commits, rebase-onto mechanics. Use when working with branch chains, stacked PRs, propagating changes across dependent branches, or resolving rebase-onto conflicts across branch stacks.
user-invocable: false
---

# Stacked PR Workflows

Rules for managing chains of dependent branches (stacked PRs). Covers propagation ordering, fixup targeting, generated commits, and conflict resolution.

## Propagation Workflow

When a branch in the stack is rewritten (rebase, fixup, amend), all descendants must be propagated. **Always propagate in parent-first DAG order.**

> **Override (commit skill, Rebase Workflow):** During propagation, the "ask before rebasing" step does not apply. Propagation is mechanical — rebase each child without pausing for confirmation. Conflict-category rules below still apply.

### Pre-flight

1. Stash or commit working changes on every branch in the stack
2. Verify the parent branch is clean and fully rebased
3. Record old tips before rewriting:

```bash
OLD_A=$(git rev-parse feature-a)
OLD_B=$(git rev-parse feature-b)
```

### Propagate

Rewrite the parent branch first, then propagate each child:

```bash
# Step 1: Rewrite feature-a (rebase, fixup, etc.)
git checkout feature-a
git rebase origin/main   # or whatever the rewrite is

# Step 2: Propagate feature-b onto the new feature-a
git rebase --onto feature-a $OLD_A feature-b

# Step 3: Propagate feature-c onto the new feature-b
git rebase --onto feature-b $OLD_B feature-c
```

The pattern is always: `git rebase --onto <new-parent-tip> <old-parent-tip> <child>`

### Green Commits During Propagation

Every commit MUST remain green after propagation. If a rebase introduces a failure in any commit, fix it and autosquash into that commit — NEVER leave a broken commit in the chain or push a fix as a separate commit.

### Post-flight

1. Verify commit count on each branch matches expectations
2. Run project-specific checks (tests, lint) on the tip of each branch
3. Push all branches (see Push Strategy)

If something goes wrong, see Common Failure Modes at the end of this skill.

### Propagation Checklist

- [ ] All branches have clean working trees
- [ ] Old tips recorded for every branch that will move
- [ ] Parent rewritten first
- [ ] Children propagated in DAG order (root-to-leaf)
- [ ] Generated commits handled (drop-and-rerun, not rebased)
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

During a rebase that encounters a generated commit:

1. Drop the generated commit (mark `d` in interactive rebase, or `git rebase --skip` during conflict)
2. Read the commit message to get the command: `git log --format=%s -1 <sha>`
3. Run the command from the commit description
4. Recommit with the same `auto:` message

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
- Push all stack branches together when possible:

```bash
git push --force-with-lease origin feature-a feature-b feature-c
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
| Child has duplicate/ghost commits after propagation | Used wrong old-tip SHA in `rebase --onto` | `git reflog show <branch>` to find correct old tip, re-propagate |
| Generated commit has merge conflicts | Rebased instead of drop-and-rerun | `git rebase --abort`, start over with drop-and-rerun |
| Child rebased before parent | Propagated in wrong order | `git reflog show <branch>` to restore, re-propagate in correct DAG order |
| `--force-with-lease` rejected | Remote updated by someone else | Fetch, inspect remote changes, decide with user |
| Rebase aborted mid-propagation | Conflict too complex to resolve inline | Branch restored to pre-rebase state; old tips still valid. Fix the issue, re-run same `rebase --onto` |
| Stack parent ambiguous | Fork in DAG, unclear which branch is parent | **Ask the user.** NEVER guess from commit dates or branch names |

## Terminology

Quick reference for terms used throughout this skill:

- **Old tip**: A branch's tip *before* a rewrite (rebase, fixup, amend)
- **Generated commit**: Deterministic command output, prefixed `auto:` — description is the command
- **Stack**: Chain of branches where each depends on the previous; parent relationships tracked in PR descriptions, no metadata files
