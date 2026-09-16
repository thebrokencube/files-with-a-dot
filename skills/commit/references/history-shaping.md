# Review Feedback and History Shaping

Use this contract after accepting review feedback or before reshaping commit history for handoff, push, or a pull-request update.

## Classify the Repair

Assign every accepted repair to its logical owner before changing history:

- **Current uncommitted unit** — apply the repair directly before creating its commit or change.
- **Earlier committed or pushed unit** — edit or fix up the owning unit with the native Git or jj procedure, propagate descendants in causal order, and rerun checks for every retained unit.
- **Independent logical unit** — create a separate commit or change when the repair has a purpose independent of existing units.
- **No owner yet** — leave the repair uncommitted until a unit owns its behavior. Never amend or fix up into an unrelated unit.

## Vocabulary

| Term | Meaning |
|---|---|
| **Accepted repair** | A review finding agreed for correction. |
| **Logical owner** | The unit whose purpose and behavior the repair changes. |
| **Retained green unit** | A commit or change that remains independently valid under the repository's applicable green-commit checks. |
| **Review boundary** | A published parent/child boundary whose reviewed parent tree and review scope remain intact. |
| **Walkability pass** | A final history-only pass before handoff, push, or pull-request update. |
| **Final-tree invariant** | Every history-only operation preserves the tree, and the final shaped tip matches the captured pre-pass tip tree. |

## Run the Walkability Pass

1. Capture the clean pre-pass tip and tree.
2. Inspect retained units oldest to newest.
3. Require meaningful conventional subjects, one logical purpose per diff, causal order, no WIP or fixup subjects, and green checks at every retained boundary.
4. Apply only history-only structural operations.
5. Compare the tree after every operation and compare the final shaped tip with the captured pre-pass tip.

If the pass requires a content repair, stop the pass, classify that repair, complete it through its logical owner, and restart with a new pre-pass tip.

## Stop Conditions

Stop and resolve the cause before continuing on:

- semantic conflict;
- ambiguous logical owner or stack parent;
- remote movement;
- unauthorized published-boundary movement;
- generated conflict;
- failed retained-unit check;
- changed tree after a history-only operation; or
- cross-repository rewrite.

Commit history records logical purpose and causal order. Do not add a review-provenance ledger, review archive, or commit-body convention. Do not use `jj absorb` as the default repair operation.

## Native Procedures

- Git: read `git-stacking.md` for fixup targeting, propagation, published-boundary checks, and recovery.
- jj: read `jj-stacking.md` for editing the owning change, descendant rebasing, tree checks, and recovery.
