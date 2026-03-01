# /folio stack

Unified stack management workflow. Bridges folio topology (`dag --branches`) with stacked-pr skill mechanics (propagation, fixup targeting, push strategy).

## Prerequisites

- folio.yml with targets that have `branch` fields set
- stacked-pr skill loaded (referenced by name, rules not duplicated here)

## Actions

### check (default)

Morning standup view. Shows the branch topology with staleness overlay.

1. Run `folio dag --branches --json --status --folio <path>` to get topology with status
2. Parse the `BranchTopology` JSON
3. Present the tree to the user:
   - Show each branch with its base, PR, and status (clean/stale/missing)
   - Highlight stale targets with their transitive cause (`stale << parent-target`)
4. Suggest next action based on state:
   - **All clean**: "Stack is clean — ready to push" (suggest `/folio stack push` if unpushed)
   - **Stale targets with stale parents**: Suggest `/folio stack propagate` (parent branch changed, children need rebase)
   - **Directly stale targets**: Suggest `/folio compose <target>` (source files changed, output needs recomposition)

### propagate

Walk topology in propagation order, rebase each child onto its new parent.

1. Run `folio dag --branches --json --folio <path>` to get topology
2. Compute propagation order from the tree (pre-order traversal = root-to-leaf)
3. For each branch in propagation order, follow the stacked-pr skill's **Propagation Workflow**:
   - Record old tips before rewriting
   - Rebase child onto new parent tip: `git rebase --onto <new-parent-tip> <old-parent-tip> <child>`
   - Verify green commits (stacked-pr: "Every commit MUST remain green after propagation")
   - Handle conflicts per stacked-pr conflict categories (mechanical → resolve, semantic → pause, generated → drop-and-rerun)
4. After all branches propagated, run `folio stale --json --folio <path>`
5. **Review gate (soft)**: Present rebased branches, conflicts resolved, and remaining stale targets. Proceed unless user objects.
6. If stale targets remain (source files changed independently of branch propagation), suggest `/folio compose <target>`

### push

Push all stack branches parent-first.

1. Run `folio dag --branches --json --folio <path>` to get topology
2. Walk in propagation order (parent-first = same order as propagation)
3. **Review gate (hard)**: Present branches to push, local vs remote tips for each, and force-with-lease status. Wait for explicit "yes" before pushing.
4. Push each branch per stacked-pr skill's **Push Strategy**:
   - Use `--force-with-lease` (never bare `--force`)
   - Push parent before children
   - If `--force-with-lease` is rejected, stop and report — do not override

When possible, push all branches in a single command:
```bash
git push --force-with-lease origin branch-a branch-b branch-c
```

## Error Handling

- **No folio.yml found**: Report error, suggest `folio init`
- **No targets with branch fields**: Report "no stack branches found", suggest adding `branch:` to targets
- **Propagation conflict**: Follow stacked-pr conflict resolution rules. Mechanical: resolve. Semantic: pause and confirm with user. Generated: drop and re-run.

## Cross-References

- **stacked-pr skill**: Propagation Workflow, Push Strategy, Conflict Resolution, Generated Commits
- **folio CLI**: `dag --branches --json --status`, `stale --json`
- **Design doc**: `~/.folio/active/files-with-a-dot/reference/stack-management-design.md`
