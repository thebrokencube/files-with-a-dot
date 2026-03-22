# Lifecycle: Park & Unpark

Forest lifecycle operations for deactivating and reactivating Jira tickets. These are skill-level workflows that combine `jf rm` (local file removal) with MCP-driven Jira operations (transitions, reparenting, field clearing).

## Park

Deactivate tickets: transition to Done, clear content, move to parking lot, remove from forest.

```
/jf park KEY1 KEY2... [--into LEADER]
```

### Workflow

1. **Parking lot onboarding**: If no parking lot epic is known for the project, ask the user for the project key, then create a parking lot epic via MCP `createJiraIssue` (summary: "[Parking Lot]", type: Epic). Remember the epic key for subsequent park operations in this session.

2. **Per key** (bottom-up — park children before parents):
   a. **Transition to Done**: Call MCP `getTransitionsForJiraIssue` to discover available transitions. Find the transition whose target category is "Done" (do not hardcode transition IDs). Call MCP `transitionJiraIssue`.
   b. **Clear content**: Call MCP `editJiraIssue` to set summary to `"[parked]"` and clear the description.
   c. **Reparent**: Call MCP `editJiraIssue` to set the parent field to the parking lot epic.

3. **Remove from forest**: Run `jf rm KEY1 KEY2...` to delete the local files. The child guard in `jf rm` enforces bottom-up order — if a key has children in the forest, those must be parked first.

4. **Report**: Summarize what was parked, any failures.

### Flags

- `--into LEADER`: Informational only — indicates which leader/initiative the parked work was under. Does not affect Jira operations.

### Bottom-up enforcement

If KEY has children in the forest, `jf rm` will refuse to remove it. Park the children first (either explicitly or by listing them before the parent in the KEY list).

## Unpark

Reactivate tickets: transition to To Do, reparent under target epic, scaffold local file, optionally pull content.

```
/jf unpark KEY1 KEY2... --epic EPIC
```

### Workflow

1. **Per key**:
   a. **Transition to To Do**: Call MCP `getTransitionsForJiraIssue` to discover available transitions. Find the transition whose target category is "To Do". Call MCP `transitionJiraIssue`.
   b. **Reparent**: Call MCP `editJiraIssue` to set the parent field to the target EPIC.

2. **Scaffold**: For each key, write a `.md` file into the forest directory with frontmatter:
   ```yaml
   ---
   jira: KEY
   sync: push
   ---
   ```
   Place the file according to forest conventions (in the epic's subdirectory if one exists, or at the forest root).

3. **Pull content** (optional): Run `jf pull` in forest mode to populate the file with the current Jira description. Only if the ticket has meaningful content (not `"[parked]"`).

4. **Report**: Summarize what was unparked, where files were created.

### Required flags

- `--epic EPIC`: The Jira epic key to reparent unparked tickets under. Required — unparking without a destination is not supported.
