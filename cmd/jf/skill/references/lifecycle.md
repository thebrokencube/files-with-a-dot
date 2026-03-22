# Lifecycle: Park

Forest lifecycle operation for permanently deactivating Jira tickets. Combines `jf rm` (local file removal) with MCP-driven Jira operations (transitions, reparenting, field clearing).

## Prerequisites

**Cloud ID**: All MCP Jira tools require a `cloudId` parameter. Discover it once per session by calling `getAccessibleAtlassianResources` — it returns a list of sites with `id` fields. Use that `id` as `cloudId` for all subsequent calls.

## Park

Deactivate tickets that are noise — over-decomposed, no longer needed, or superseded. The ticket becomes a blank placeholder in a parking lot epic, available for repurposing later.

```
/jf park KEY1 KEY2...
```

### Workflow

All Jira operations (steps 1-2) MUST complete before `jf rm` (step 3). Do not delete local files until Jira state is settled — if Jira operations fail, the forest files are your recovery path.

#### Step 1: Parking lot onboarding

If no parking lot epic is known for this project:
- Extract the project key from the provided keys (e.g., `PROJ-123` → `PROJ`). If keys span multiple projects, create one parking lot per project.
- Create the parking lot epic:
  ```
  createJiraIssue({
    cloudId: "<cloud-id>",
    projectKey: "PROJ",
    issueType: "Epic",
    summary: "[Parking Lot]"
  })
  ```
- Hold the returned epic key in conversation context for subsequent park calls in this session. There is no persistent storage — if a new session needs it, ask the user or search for `"[Parking Lot]"` in the project.

#### Step 2: Per-key Jira operations

Process keys **bottom-up** (children before parents). This matters for both `jf rm` (child guard) AND the Jira operations — don't transition a parent to Done while its children are still active.

For each key:

**2a. Transition to Done:**
```
getTransitionsForJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123"
})
```
The response contains a `transitions` array. Each transition has a `to` object with `statusCategory.name`. Find the transition where `to.statusCategory.name` is `"Done"`. Then:
```
transitionJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  transitionId: "<id from matching transition>"
})
```

- If the ticket is **already Done** (no "Done" transition available and current status category is "Done"): skip this step, continue to 2b.
- If **no direct Done transition exists** (workflow requires intermediate steps): try transitioning through available states toward Done (e.g., To Do → In Progress → Done). If that fails, report the issue and skip this key.

**2b. Clear content and reparent** (single call):
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "summary": "[parked]",
    "description": null,
    "parent": { "key": "PROJ-999" }
  }
})
```
- `description: null` clears the description field.
- `parent: { "key": "..." }` sets the parent to the parking lot epic. This is the standard Jira next-gen/team-managed parent field format.
- Combine summary, description, and parent into one `editJiraIssue` call to minimize API calls and avoid partial-update risk.

**On failure for any key**: Log the error, skip that key, continue with remaining keys. At the end, only pass successfully-parked keys to `jf rm`.

#### Step 3: Remove from forest

Run `jf rm` with only the keys that succeeded in step 2:
```bash
jf rm PROJ-123 PROJ-456 ...
```

The child guard in `jf rm` will refuse to remove a node with children. Since step 2 processed bottom-up, children should already be removed. If `jf rm` fails on a key, report it — something went wrong with ordering.

#### Step 4: Report

Summarize:
- How many tickets parked successfully
- Which keys failed and why (transition unavailable, edit failed, etc.)
- The parking lot epic key for reference

### Bottom-up enforcement

If KEY has children in the forest, `jf rm` will refuse to remove it. Park the children first (either explicitly or by listing them before the parent in the KEY list). The agent must sort the provided keys bottom-up before processing — use `jf tree` or `jf list --json` to determine the hierarchy.

## Repurposing parked tickets

Parked tickets are blank placeholders sitting in the parking lot epic. When you need new tickets (e.g., via `jf create-missing`), check the parking lot first — you can repurpose an existing parked ticket instead of creating a new one. This keeps Jira ticket counts down and avoids orphaned keys.

To repurpose via a single `editJiraIssue` call:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "summary": "new ticket summary",
    "description": <ADF or markdown content>,
    "parent": { "key": "PROJ-100" }
  }
})
```
Then transition back to the appropriate status (typically "To Do" — use `getTransitionsForJiraIssue` to find it, matching `to.statusCategory.name` of `"To Do"`). The ticket keeps its original key but gets a new identity.
