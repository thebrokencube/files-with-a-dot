# Lifecycle: Park

Forest lifecycle operation for permanently deactivating Jira tickets. Combines `jf rm` (local file removal) with MCP-driven Jira operations (transitions, reparenting, field clearing).

## Prerequisites

**Cloud ID**: All MCP Jira tools require a `cloudId` parameter. Discover it once per session by calling `getAccessibleAtlassianResources` — it returns a list of sites with `id` fields. Use that `id` as `cloudId` for all subsequent calls.

**Config file** (`~/.jf.yml`): Persistent storage for parking lot epic keys, keyed by project. The agent reads and writes this file directly (plain YAML, no binary support needed). Format:
```yaml
parking_lots:
  PROJ: PROJ-999
  OTHER: OTHER-456
```

## Park

Deactivate tickets that are noise — over-decomposed, no longer needed, or superseded. The ticket becomes a blank placeholder in a parking lot epic, available for repurposing later.

```
/jf park KEY1 KEY2...
```

### Workflow

All Jira operations (steps 1-2) MUST complete before `jf rm` (step 3). Do not delete local files until Jira state is settled — if Jira operations fail, the forest files are your recovery path.

#### Step 1: Parking lot onboarding

Extract the project key from the provided keys (e.g., `PROJ-123` → `PROJ`). If keys span multiple projects, resolve a parking lot for each.

Resolve the parking lot epic for each project in this order:

1. **Check `~/.jf.yml`** — read the file and look for the project under `parking_lots`. If found, use that epic key.
2. **Search Jira** — if not in config, search for an existing one:
   ```
   searchJiraIssuesUsingJql({
     cloudId: "<cloud-id>",
     jql: "project = PROJ AND issuetype = Epic AND summary ~ \"[Parking Lot]\"",
     fields: ["summary", "status"]
   })
   ```
   If found, use that epic and save it to `~/.jf.yml`.
3. **Create one** — if neither config nor search found it:
   - First discover required fields for Epic in this project:
     ```
     getJiraIssueTypeMetaWithFields({
       cloudId: "<cloud-id>",
       projectKey: "PROJ",
       issueTypeName: "Epic"
     })
     ```
     The response lists all fields and which are `required`. Some projects have mandatory custom fields (e.g., T-Shirt Size estimates). Include any required fields in the create call using `additional_fields`.
   - Create the epic:
     ```
     createJiraIssue({
       cloudId: "<cloud-id>",
       projectKey: "PROJ",
       issueType: "Epic",
       summary: "[Parking Lot]",
       additional_fields: { ... any required custom fields ... }
     })
     ```
   - **Save to `~/.jf.yml`** — add/update the project entry under `parking_lots`. Create the file if it doesn't exist.

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

**2b. Clear content and reparent** (two calls):

First, set summary and reparent:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "summary": "[parked]",
    "parent": { "key": "PROJ-999" }
  }
})
```

Then, clear the description with an empty ADF document:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  description: "{\"version\":1,\"type\":\"doc\",\"content\":[]}",
  responseContentFormat: "adf"
})
```
- Do NOT pass `description: null` — the MCP tool rejects it.
- The description must be an empty ADF document string, not null or empty string.
- `parent: { "key": "..." }` sets the parent to the parking lot epic. This is the standard Jira next-gen/team-managed parent field format.

**On failure for any key**: Log the error, skip that key, continue with remaining keys. At the end, only pass successfully-parked keys to `jf rm`.

#### Step 3: Remove from forest

Run `jf rm` with only the keys that succeeded in step 2. Use the `--dir` flag to specify the forest directory if not running from within it:
```bash
jf rm --dir /path/to/forest PROJ-123 PROJ-456 ...
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

To repurpose, first set the new identity:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "summary": "new ticket summary",
    "parent": { "key": "PROJ-100" }
  }
})
```
Then set the description (as ADF or markdown via `description` parameter) and transition back to the appropriate status (typically "To Do" — use `getTransitionsForJiraIssue` to find it, matching `to.statusCategory.name` of `"To Do"`). The ticket keeps its original key but gets a new identity.
