# Testing: Safe-Sync Test Harness

Developer-only test infrastructure for validating the safe-sync engine. Creates a dedicated 13-node test forest with real Jira tickets covering all plan rules.

**IMPORTANT**: `jf test run` makes real Jira API calls. It is NOT part of `go test ./...` and requires human oversight. The Go unit tests in `cmd_testharness_test.go` only test local file-generation logic against temp directories — they never call Jira.

## Prerequisites

**Cloud ID**: Read from `~/.jf.yml` at `cloud_id` — do NOT call `getAccessibleAtlassianResources`.

**Jira auth**: `jf test setup --seed-baselines`, `jf test run`, and `jf test reset` require valid Jira credentials (they make read-only API calls). `jf test run --execute` makes write calls.

## Full Setup Workflow

### Phase 1: Epic Resolution

Resolve a test epic using 3-tier resolution (same pattern as parking lot in lifecycle.md):

1. **Check `.test-config.yml`** — if the test forest already exists and has an epic key, reuse it.
2. **Search Jira**:
   ```
   searchJiraIssuesUsingJql({
     cloudId: "<cloud-id>",
     jql: "project = BEN AND issuetype = Epic AND summary ~ \"[jf Test Forest]\"",
     fields: ["summary", "status"]
   })
   ```
3. **Create epic** via MCP if not found:
   - First discover required fields: `getJiraIssueTypeMetaWithFields({ cloudId, projectKey: "BEN", issueTypeName: "Epic" })`
   - Create: `createJiraIssue({ cloudId, projectKey: "BEN", issueTypeName: "Epic", summary: "[jf Test Forest] Safe-sync validation", ... })`

### Phase 2: Create Jira Tickets

Create 11 tickets under the test epic (nodes 12-13 don't need tickets — they use TBD and a nonexistent key).

Each ticket needs a specific description state:

| Node | Ticket Description State |
|------|-------------------------|
| 1 (empty-push) | Leave description **empty** |
| 2 (first-push-safe) | Leave description **empty** |
| 3 (first-push-conflict) | Set **substantive** description |
| 4 (first-pull-safe) | Set **substantive** description |
| 5 (first-pull-conflict) | Set **substantive** description |
| 6 (local-changed) | Set **substantive** description |
| 7 (overwrite-blocked) | Set **substantive** description |
| 8 (conflict) | Set **substantive** description |
| 9 (unchanged) | Set **substantive** description |
| 10 (both-local-only) | Set **substantive** description |
| 11 (both-remote-only) | Set **substantive** description |

**Create ticket** (example for node 3):
```
createJiraIssue({
  cloudId: "<cloud-id>",
  projectKey: "BEN",
  issueTypeName: "Story",
  summary: "[jf Test] first-push-conflict",
  parentKey: "<epic-key>"
})
```

**Set description** (for tickets needing substantive content):
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueKey: "BEN-XXX",
  description: {
    type: "doc",
    version: 1,
    content: [{
      type: "paragraph",
      content: [{ type: "text", text: "Test content for <node-name>." }]
    }]
  }
})
```

Nodes 1 and 2 are left with empty descriptions — skip the editJiraIssue call for those.

### Phase 3: Key Handoff

Collect all created ticket keys and pass them to setup:

```bash
jf test setup \
  --epic BEN-XXXXX \
  --dir ~/.jf/test/safe-sync \
  --keys empty-push=BEN-111,first-push-safe=BEN-112,first-push-conflict=BEN-113,first-pull-safe=BEN-114,first-pull-conflict=BEN-115,local-changed=BEN-116,overwrite-blocked=BEN-117,conflict=BEN-118,unchanged=BEN-119,both-local-only=BEN-120,both-remote-only=BEN-121
```

This generates: forest.yml, 13 .md files, .test-config.yml.

### Phase 4: Baseline Seeding

```bash
jf test setup --seed-baselines --dir ~/.jf/test/safe-sync
```

This fetches remote ADF hashes from Jira (read-only) and seeds state.json with baseline entries for nodes 6-11. The baseline hash strategy creates the desired "changed"/"unchanged" relationships by storing real hashes for unchanged sides and synthetic stale hashes for changed sides.

## Validation Workflow

### Per-Track Validation

Run after completing each track's commits:

```bash
jf test run 0 --dir ~/.jf/test/safe-sync   # Track 0: forest structure
jf test run 1 --dir ~/.jf/test/safe-sync   # Track 1: inline safety guards
jf test run 2 --dir ~/.jf/test/safe-sync   # Track 2: engine plan output
jf test run 3 --dir ~/.jf/test/safe-sync   # Track 3: full pipeline
jf test run 4 --dir ~/.jf/test/safe-sync   # Track 4: grep checks
```

### Conflict Resolution Validation

```bash
jf test run --resolve local --dir ~/.jf/test/safe-sync
jf test run --resolve remote --dir ~/.jf/test/safe-sync
```

Changes the expected outcome for node 8 (conflict): BLOCKED(conflict) becomes PUSH or PULL respectively.

### Full Round-Trip (After Track 3)

```bash
jf test run --execute --dir ~/.jf/test/safe-sync
```

Exercises the full Jira mutation pipeline. Verify results via MCP:
- Push nodes: check Jira description was updated (`getJiraIssue`)
- Pull nodes: check local file was updated (read the .md file)
- Blocked nodes: confirm no mutation occurred
- state.json: verify per-node entries were recorded

## Reset

Restore the test forest to its baseline state after modifications:

```bash
jf test reset --dir ~/.jf/test/safe-sync
```

Regenerates .md files from node definitions and re-seeds baselines (requires Jira auth for read-only API calls to re-fetch remote hashes).

## Teardown

```bash
jf test teardown --dir ~/.jf/test/safe-sync
```

Removes the local forest directory and .test-config.yml. Prints the exact `jf park` command needed for Jira ticket cleanup — the agent or user runs those separately.

## Test Node Reference

| # | Name | Sync | Expected | Rule |
|---|------|------|----------|------|
| 1 | empty-push | push | BLOCKED(empty) | Emptiness guard |
| 2 | first-push-safe | push | PUSH | First sync, other side empty |
| 3 | first-push-conflict | push | BLOCKED(first-push) | First sync, other has content |
| 4 | first-pull-safe | pull | PULL | First sync, other side empty |
| 5 | first-pull-conflict | pull | BLOCKED(first-pull) | First sync, other has content |
| 6 | local-changed | push | PUSH | This side changed |
| 7 | overwrite-blocked | push | BLOCKED(overwrite) | Other side changed |
| 8 | conflict | both | BLOCKED(conflict) | Both sides changed |
| 9 | unchanged | both | SKIP | Neither changed |
| 10 | both-local-only | both | PUSH | sync:both, local changed only |
| 11 | both-remote-only | both | PULL | sync:both, remote changed only |
| 12 | tbd-skip | push | SKIP | TBD key |
| 13 | remote-err | push | BLOCKED(remote-unknown) | Remote unreachable |
