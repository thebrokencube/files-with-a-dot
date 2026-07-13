# jj Stacked PR Workflows

Rules for managing stacks of changes using jj (Jujutsu). Covers building stacks, bookmark management, push semantics, PR creation, review cycles, and post-merge cleanup.

## Key Differences from Git

- **No staging area.** The working copy IS the change — files auto-snapshot.
- **Change IDs are stable.** Rewrites (rebase, squash) create new revisions but the change ID persists. Bookmarks follow automatically.
- **No current branch.** You edit a specific change, not a branch. Bookmarks must be explicitly created.
- **Operation log.** Every jj operation is recorded and reversible via `jj op undo`.

## Building a Stack

```bash
# Start from main
jj new main -m "refactor: extract helper"
# ... edit files (auto-snapshotted) ...

jj new -m "feat: add widget support"
# ... edit more files ...

jj new -m "feat: add widget tests"
```

Each `jj new` creates a child of the current working-copy change. The result is a linear chain descending from main.

Describe in-progress work at any time:
```bash
jj describe -m "type(scope): updated description"
```

Split a change that grew too large:
```bash
jj split    # interactive: choose which hunks go in the first half
```

## Bookmark Management

GitHub requires named branches for PRs. Two approaches:

### Explicit bookmarks (recommended for team-visible stacks)

```bash
# Name each change in the stack
jj bookmark create refactor-helper -r @--      # grandparent
jj bookmark create feat-widget -r @-           # parent
jj bookmark create feat-widget-tests -r @      # current
```

Or from anywhere using change IDs:
```bash
jj bookmark create refactor-helper -r <change-id>
```

### `--change` auto-names (for solo/throwaway work)

```bash
jj git push --change <change-id>
# Auto-creates bookmark named push-<short-change-id>
```

**Recommendation**: Use explicit bookmarks for stacks with multiple PRs. Use `--change` for quick single-change pushes.

### Moving bookmarks

Bookmarks automatically follow rewrites (rebase, squash). But they do NOT auto-advance when you create children. After extending a change:

```bash
jj bookmark set feat-widget -r @    # create-or-update
```

## Push

jj's push behaves like `git push --force-with-lease` by default — verifies remote position before force-pushing.

```bash
# Push all bookmarks
jj git push --all

# Push specific bookmarks
jj git push --bookmark refactor-helper --bookmark feat-widget

# Push all bookmarks on changes between trunk and working copy
jj git push --revisions 'trunk()..@'

# Dry run first
jj git push --all --dry-run
```

## PR Creation

### Colocated repos (`.jj` + `.git` both present)

`gh` CLI works normally:
```bash
gh pr create --head refactor-helper --base main --title "refactor: extract helper"
gh pr create --head feat-widget --base refactor-helper --title "feat: add widget support"
```

### Non-colocated repos (`.jj` only)

Set `GIT_DIR` for `gh` compatibility:
```bash
export GIT_DIR=$(jj git root)
```

Use direnv to automate: `echo 'export GIT_DIR=$(jj git root)' > .envrc`

**Always set `--base` to the parent branch** — `gh pr create` defaults to `main`, which is wrong for child branches.

For the PR body itself (house style vs a repo template, and the `## The stack` list every stacked PR should carry), see `pr-descriptions.md`.

## Addressing Review Feedback

### Edit the target change directly

```bash
jj edit <change-id>
# ... make changes (auto-snapshotted) ...
# All descendants automatically rebase
```

### Or squash fixes into a target

```bash
jj new <target-change>     # create child of the target
# ... make fixes ...
jj squash                   # squash into parent (the target)
```

For fixing a specific target from anywhere:
```bash
jj squash --into <target-change-id>
```

After edits, push the updated stack:
```bash
jj git push --all    # auto-force-pushes changed bookmarks
```

## Post-Merge Cleanup

When GitHub squash-merges a PR, it creates a new commit unrelated to jj's change ID. jj cannot auto-detect this. Manual cleanup required:

```bash
# 1. Fetch updated main
jj git fetch

# 2. Abandon the merged change
jj abandon <change-id-of-merged-change>

# 3. Rebase the rest of the stack onto updated main
jj rebase -s <next-change-in-stack> -d main

# 4. Delete the old bookmark
jj bookmark delete <merged-bookmark>

# 5. Update the GitHub PR base branch
gh pr edit <next-pr-number> --base main

# 6. Push the rebased stack
jj git push --all
```

Repeat for each PR as it merges from the bottom up.

### Regular merge (non-squash)

Simpler — jj recognizes the merge after fetch:
```bash
jj git fetch
jj rebase -b <next-change> -d main
```

## Generated Commits

Same `auto:` prefix convention as git. The description IS the command:

```bash
jj describe -m "auto(codegen): graphqlme"
jj new    # start next change
```

During stack edits, regenerate rather than manually fix:
1. Note the command from the description
2. `jj abandon <generated-change-id>`
3. `jj new <parent>` and re-run the command
4. `jj describe -m "auto(scope): <same command>"`

## Recovery

jj's operation log makes recovery trivial compared to git reflog:

```bash
# Undo the last operation
jj op undo

# View operation history
jj op log

# Restore to a specific operation
jj op restore <op-id>
```

Every `jj` command is an operation. If propagation, squash, or rebase goes wrong, `jj op undo` restores the entire repo state in one step.

## Conflict Resolution

Three categories with different handling — same taxonomy as git stacking.

### Mechanical Conflicts

Whitespace changes and trivially resolvable markers. Does NOT include import reordering.

**Action**: Resolve and continue. In jj, conflicts are materialized in the working copy — edit the conflict markers, then the change auto-snapshots clean.

### Semantic Conflicts

Logic changes from a parent change affect code in a child change. The child's assumptions about behavior may no longer hold.

**Action**: Pause and confirm with the user. Describe what changed in the parent and what the child assumes. jj makes this easier to inspect: `jj diff -r <parent-change>` shows exactly what changed.

### Generated Conflicts

Conflicts in `auto:` changes — lockfiles, codegen output, schema files.

**Action**: Abandon the change and re-run the command from its description. NEVER merge generated content manually.

```bash
jj abandon <generated-change-id>
jj new <parent>
# re-run command
jj describe -m "auto(scope): <same command>"
```

## Decomposition

Same ordering principle as git: foundational changes first, then implementation, then integration, then consumer-facing.

- **Within a change**: refactor → implement → wire → UI/API
- **Within a stack**: root change absorbs structural risk; children make the now-easy changes
- **Each change MUST be independently reviewable and mergeable**
- Stack when total diff > ~400 lines with natural layers; single change otherwise

## Common Failure Modes

| Symptom | Cause | Recovery |
|---------|-------|----------|
| Bookmark not on expected change | Forgot `jj bookmark set` after extending | `jj bookmark set <name> -r <change-id>` |
| Orphaned change after squash-merge | GitHub squash-merge creates new commit jj doesn't recognize | `jj abandon <change-id>`, rebase children onto main |
| Push rejected | Remote bookmark diverged (someone else pushed) | `jj git fetch`, inspect with `jj log`, decide with user |
| Wrong change edited | `jj edit <wrong-id>` then made edits | `jj op undo` restores previous state cleanly |
| Bookmark deleted accidentally | `jj bookmark delete` on wrong name | `jj op undo` to restore |
| Stack shows conflicts after parent edit | Child changes have semantic dependency on parent | Resolve conflicts in each child's working copy (edit into the change), or `jj op undo` and ask user |

## Useful Revsets

```bash
jj log -r 'trunk()..@'              # your entire stack
jj log -r 'mutable()'               # all non-trunk changes
jj log -r '::@ & mutable()'         # ancestors back to trunk
jj log -r '<change-id>::'           # all descendants of a change
```

## Terminology

- **Change**: A logical unit of work, identified by a stable change ID
- **Revision**: A concrete commit hash — changes on rewrite
- **Bookmark**: jj's equivalent of a git branch — points at a revision, follows rewrites
- **Stack**: Chain of changes where each is a child of the previous
- **Operation**: A recorded jj action, reversible via `jj op undo`
