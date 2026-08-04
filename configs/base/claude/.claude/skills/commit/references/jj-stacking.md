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

## Restructuring an Existing Stack

### Moving a change into the right commit

| Tool | Use |
|---|---|
| `jj squash -u --into <change> <paths>` | move file-level changes into a specific commit |
| `jj squash -u --from X --into Y <paths>` | move a change *down* the stack |
| `jj squash -u -r <child>` | fold a child into its direct parent — a union of diffs, so it **never** conflicts |
| `jj split <paths>` | partition by fileset, no diff editor |
| `jj rebase -r <rev> -A/-B <target>` | reorder, descendants preserved |
| `jj absorb` | line-level auto-fixup to the nearest mutable ancestor that last touched those lines |

**Target the newest commit that touches the file.** That applies cleanly. A fix that *depends* on later code
must land at or above the commit introducing that code, or the lower commit stops compiling.

**Pre-flight every non-adjacent move — one command, decisive:**

```
jj log -r 'A..B & files("PATH")'
```

Zero intervening touchers → safe. Several → expect a cascade, and prefer a different seam.

**`jj absorb` is not a general fixup tool.** It is safe only where the lines have one unambiguous owner. On
pervasively-shared files it cascades conflicts across many commits at once and leaves orphan hunks. A late
cross-cutting sweep (a comment trim, a rename) becomes the "newest toucher" of nearly everything and poisons
absorb's targeting — re-run the revset with `& ~<sweep>` to find the real owner.

**Partial-file move with no interactive editor:** `jj edit <source>`, delete the block, `jj new <dest>`,
paste it back, `jj squash`. The tree-invariant check below catches any drift.

### Verify a restructure preserved behaviour

After a reflow, the final tree must be unchanged:

```
jj diff --from <original-tip-commit-id> --to <new-tip> --stat     # must be empty
```

Run it after **every** structural step, not only at the end. When content intentionally moved to a sibling,
the diff shows exactly those files and nothing else.

**Moving a change across a stacked-PR boundary needs three checks, not one:**

1. child tip tree identical before/after — behaviour preserved
2. parent tip tree identical to what the remote already had — the donor PR's diff never moved, so its review
   comments survive
3. `jj diff --from main --to <parent-bookmark> --git | grep -c '<moved-symbol>'` → **0**

Checks 1 and 2 both pass even if the lines never moved. Check 3 is what proves the boundary is right.

**Validate each chunk in isolation before opening its PR.** A branch authored as one green unit often has
its lint or test fixes in a *later* commit than the code they cover, so the base PR fails CI alone even
though the tip is green. `jj new <bookmark>`, run the CI-parity checks, fix *at that chunk*.

### Stack-wide rename

- **Content — use `jj fix`, never per-commit edits.** Configure a sed tool
  (`jj config set --repo fix.tools.<name>.command`, plus `.patterns`), then `jj fix -s 'roots(main..<tip>)'`.
  It rewrites each commit's changed lines with **zero new conflicts**. Editing the base and letting it rebase
  cascades into every descendant.
- `jj fix` touches content only, and only paths matching `patterns`. Repo-root files outside the globs are
  missed — `grep -rIl -i <oldname>` afterwards and fix stragglers at their owning commit.
- **File moves go at the file's creating commit** (`jj log -r 'roots(main..<tip> & files("PATH"))'`). Every
  later commit editing it still gets a delete/modify conflict, bounded by the toucher count.
- One file per step, verifying `@` after each. Rapid successive `jj edit`s let the working copy drift.

### When a change fights the stack, move the seam

A semantic edit to a file that many later commits churn cascades as modify/modify. Combining two PRs — rebase
the upper onto `main`, close the lower, fold its commits in — reaches the same end state with a base change
and one close instead of hand-merging N commits.

**Re-point bases before deleting a bookmark.** Deleting a bookmark that is another PR's base auto-closes that
PR, and a closed PR's base cannot be re-pointed (`gh pr edit --base` fails) — it has to be recreated under a
new number.

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
| A jj command hangs for minutes | `squash`/`describe` opened `$EDITOR`; any squash whose *source* has a description does | Always `-u` / `--use-destination-message`; run with `JJ_EDITOR=true` (`JJ_EDITOR=:` fails — `:` is a shell builtin). Diagnose: `ps -o %cpu,time -p <pid>` near 0% means blocked, and `pgrep -P <pid>` names the editor |
| `jj new <change-id>` leaves `@` unmoved | the change is divergent; jj prints only a `Hint:` line | Pass the 12-hex commit id, then confirm with `jj log -r '@-'`. List copies with `jj log -r 'change_id(<id>)'` and abandon the obsolete ones |
| `jj git push` rejects for conflicts but `jj status` is clean | a mid-stack commit records a conflict that a later commit resolves | Check the whole stack: `jj log -r 'main..@' -T 'if(conflict,"x ","") ++ description.first_line() ++ "\n"'`. Trust the flag, not `jj diff -r <conflicted>`, whose output is scrambled |
| `jj diff`/`jj status` shows stale state | `--ignore-working-copy` skips the snapshot | Run a plain `jj st` first |
| One operation must be undone without rewinding others' work | `jj op restore` rewinds the whole repo, including concurrent sessions' bookmarks | `jj operation revert <op-id>` reverts that operation alone |
| Changes vanish from a descendant after peeling a commit out | the descendant edits the same file in non-overlapping regions, so jj auto-merges with no conflict | Capture the descendant's bytes first (`jj file show -r <tip> PATH`) and write them back; re-run the tree-invariant check |
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
