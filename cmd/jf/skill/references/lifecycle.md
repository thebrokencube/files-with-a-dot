# Lifecycle: Park

Forest lifecycle operation for permanently deactivating Jira tickets. Combines `jf rm` (local file removal) with MCP-driven Jira operations (transitions, reparenting, field clearing).

## Park

Deactivate tickets that are noise — over-decomposed, no longer needed, or superseded. The ticket becomes a blank placeholder in a parking lot epic, available for repurposing later.

```
/jf park KEY1 KEY2...
```

### Workflow

1. **Parking lot onboarding**: If no parking lot epic is known for the project, ask the user for the project key, then create a parking lot epic via MCP `createJiraIssue` (summary: "[Parking Lot]", type: Epic). Remember the epic key for subsequent park operations in this session.

2. **Per key** (bottom-up — park children before parents):
   a. **Transition to Done**: Call MCP `getTransitionsForJiraIssue` to discover available transitions. Find the transition whose target category is "Done" (do not hardcode transition IDs). Call MCP `transitionJiraIssue`.
   b. **Clear content**: Call MCP `editJiraIssue` to set summary to `"[parked]"` and clear the description.
   c. **Reparent**: Call MCP `editJiraIssue` to set the parent field to the parking lot epic.

3. **Remove from forest**: Run `jf rm KEY1 KEY2...` to delete the local files. The child guard in `jf rm` enforces bottom-up order — if a key has children in the forest, those must be parked first.

4. **Report**: Summarize what was parked, any failures.

### Bottom-up enforcement

If KEY has children in the forest, `jf rm` will refuse to remove it. Park the children first (either explicitly or by listing them before the parent in the KEY list).

## Repurposing parked tickets

Parked tickets are blank placeholders sitting in the parking lot epic. When you need new tickets (e.g., via `jf create-missing`), check the parking lot first — you can repurpose an existing parked ticket instead of creating a new one. This keeps Jira ticket counts down and avoids orphaned keys.

To repurpose: set a new summary, description, and parent via MCP `editJiraIssue`, then transition back to the appropriate status. The ticket keeps its original key but gets a new identity.
