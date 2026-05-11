# Lifecycle: Park

Forest lifecycle operation for permanently deactivating Jira tickets. Combines `jf rm` (local file removal) with MCP-driven Jira operations (transitions, reparenting, field clearing).

## Canonical Parked State

Every parked ticket MUST match this spec. The park workflow enforces it; the audit command detects drift.

| Field | Value | Notes |
|---|---|---|
| `summary` | `[parked]` | Lowercase, no trailing text |
| `description` | empty ADF doc | `{"version": 1, "type": "doc", "content": []}` |
| `status` | per `parking_lots.<PROJ>.status` | Varies by project — check `~/.jf.yml` |
| `parent` | per `parking_lots.<PROJ>.epic` | Parking lot epic |
| `issuetype` | Task | If above Task level (Epic, Project Name, Initiative), demote to Task. Task/Story/Bug left as-is. |
| `assignee` | null | No one responsible |
| `labels` | `[]` | No phantom label pollution |
| `fixVersions` | `[]` | No false version commitments |
| `components` | project default (or `[]` if not required) | Required in some projects. Leave as project default if clearing fails. |

Required custom fields (fields Jira won't let you clear) are left unchanged. Optional custom fields are left unchanged — not worth the API calls to discover and clear per-ticket.

## Prerequisites

**Cloud ID**: All MCP Jira tools require a `cloudId` parameter. Read it from `~/.jf.yml` at `cloud_id` — do NOT call `getAccessibleAtlassianResources` each session.

**Config file** (`~/.jf.yml`): Persistent storage for parking lot settings, keyed by project. The agent reads and writes this file directly (plain YAML, no binary support needed). Format:
```yaml
parking_lots:
  PROJ:
    epic: PROJ-999
    status: Triage        # target status name for parked tickets
  OTHER:
    epic: OTHER-456
    status: Backlog
```
The `status` field is the **Jira status name** to transition parked tickets into (e.g., `"Backlog"`, `"Triage"`). This is the status name as it appears in Jira, NOT the status category. Match it against the transition's `name` or `to.name` field, not `statusCategory.name`. Parked tickets are dormant placeholders, not completed work — do not use "Done". If `status` is missing, default to `"Backlog"` and ask the user to confirm.

## Park

Deactivate tickets that are noise — over-decomposed, no longer needed, or superseded. The ticket becomes a blank placeholder in a parking lot epic, available for repurposing later.

```
/jf park KEY1 KEY2...
```

### Workflow

All Jira operations (steps 1-3) MUST complete before `jf rm` (step 4). Do not delete local files until Jira state is settled — if Jira operations fail, the forest files are your recovery path.

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
       issueTypeName: "Epic",
       summary: "[Parking Lot]",
       additional_fields: { ... any required custom fields ... }
     })
     ```
   - **Save to `~/.jf.yml`** — add/update the project entry under `parking_lots`. Create the file if it doesn't exist.

#### Step 2: Pre-park guards

For each key, before any mutations:

**2a. Higher-level type with children**: If the ticket is an Epic, Project Name, or Initiative, check for Jira children:
```
searchJiraIssuesUsingJql({
  cloudId: "<cloud-id>",
  jql: "parent = PROJ-123",
  fields: ["summary"],
  maxResults: 1
})
```
If children exist: **block**. Report: "PROJ-123 is an Epic with children — park or reparent children first." Type demotion to Task would orphan them. Skip this key.

**2b. Issue links**: Fetch the ticket and check for `issuelinks`. If present: **warn** in the final report ("PROJ-123 has N issue links — review before parking"). Do not block or clear links — they may be the only record of why the ticket existed.

#### Step 3: Per-key Jira operations

Process keys **bottom-up** (children before parents). This matters for both `jf rm` (child guard) AND the Jira operations — don't transition a parent while its children are still active.

Read the target status from `~/.jf.yml` for this project (`parking_lots.<PROJECT>.status`). If not set, default to `"Backlog"` and ask the user to confirm.

For each key:

**3a. Type demotion** (if needed):

If the ticket's `issuetype` is above Task level (Epic, Project Name, Initiative), convert to Task:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "issuetype": { "name": "Task" }
  }
})
```
Task, Story, Bug, Design Task — leave as-is.

**3b. Transition to target status:**
```
getTransitionsForJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123"
})
```
The response contains a `transitions` array. Each transition has a `name` and a `to` object with `name` and `statusCategory`. Find the transition where `name` or `to.name` matches the target status from config (e.g., `"Triage"`). Then:
```
transitionJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  transitionId: "<id from matching transition>"
})
```

- If the ticket is **already in the target status** (no matching transition and current status name matches): skip this step, continue to 3c.
- If **no direct transition exists** to the target status: try transitioning through available states toward it. If that fails, report the issue and skip this key.

**3c. Clear fields, set summary, and reparent** (two calls):

First, set summary, reparent, and clear metadata fields:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "summary": "[parked]",
    "parent": { "key": "PROJ-999" },
    "assignee": null,
    "labels": [],
    "fixVersions": []
  }
})
```

Then, clear the description with an empty ADF document. This must be a separate call because `contentFormat: "adf"` would affect how other fields are interpreted:
```
editJiraIssue({
  cloudId: "<cloud-id>",
  issueIdOrKey: "PROJ-123",
  fields: {
    "description": {"version": 1, "type": "doc", "content": []}
  },
  contentFormat: "adf"
})
```
- Do NOT pass `description: null` or an empty string — the MCP tool rejects both.
- The description must be an empty ADF document object, passed inside `fields`.
- `contentFormat: "adf"` tells the MCP tool to interpret the description value as ADF, not markdown.

**On failure for any key**: Log the error, skip that key, continue with remaining keys. At the end, only pass successfully-parked keys to `jf rm`.

#### Step 4: Remove from forest

Run `jf rm` with only the keys that succeeded in step 3. Use the `--dir` flag to specify the forest directory if not running from within it:
```bash
jf rm --dir /path/to/forest PROJ-123 PROJ-456 ...
```

The child guard in `jf rm` will refuse to remove a node with children. Since step 3 processed bottom-up, children should already be removed. If `jf rm` fails on a key, report it — something went wrong with ordering.

#### Step 5: Report

Summarize:
- How many tickets parked successfully
- Which keys failed and why (transition unavailable, edit failed, children exist, etc.)
- Issue link warnings (if any)
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

**Re-apply project defaults after repurposing.** Parking clears components, labels, and assignee. After setting the new identity, re-apply project field defaults from `~/.jf.yml` (components, custom fields) so the ticket matches project conventions. If the repurposed ticket needs a type other than Task (e.g., Epic), change `issuetype` via `editJiraIssue`.

## Audit parking lot

Detect drift from the canonical parked state. Run ad-hoc when you suspect inconsistency, or after bulk parking operations.

```
/jf audit-parking [PROJECT]
```

### Workflow

1. **Read config** — get `parking_lots.<PROJECT>.epic` and `parking_lots.<PROJECT>.status` from `~/.jf.yml`. If PROJECT not specified, audit all configured projects.

2. **Fetch all children** of the parking lot epic:
   ```
   searchJiraIssuesUsingJql({
     cloudId: "<cloud-id>",
     jql: "parent = PROJ-999 ORDER BY key ASC",
     fields: ["summary", "status", "issuetype", "assignee", "labels", "fixVersions", "components"]
   })
   ```

3. **Check each ticket** against the canonical state table. Collect violations:
   - Summary != `[parked]` (exact match, case-sensitive)
   - Status != expected status from config
   - Type is above Task level (Epic, Project Name, Initiative)
   - Assignee is set
   - Labels non-empty
   - fixVersions non-empty
   - Status = Done (flag separately — these may need deletion, not re-parking)

4. **Report** violations per ticket: key, which fields are non-canonical, current values.

5. **Offer normalization**: "Normalize N drifted tickets?" If yes, run the park workflow (Step 3 only — they're already in the parking lot) on each drifted ticket to bring it to canonical state.
