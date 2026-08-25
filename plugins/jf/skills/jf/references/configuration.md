# Configuration

`~/.jf.yml` is the persistent config store for jf. It holds per-environment values
that don't belong in the forest — cloud IDs, project field defaults, parking lot settings.


## Schema

```yaml
cloud_id: "your-cloud-id-here"

projects:
  PROJ:
    components: ["My Component"]
    some_custom_field: "Required Value"   # customfield_NNNNN
  OTHER: {}

parking_lots:
  PROJ:
    epic: PROJ-999
    status: Backlog
```

## Keys

| Key | Type | Purpose |
|-----|------|---------|
| `cloud_id` | string | Atlassian Cloud ID for all MCP Jira tool calls. Pass directly — do not call `getAccessibleAtlassianResources` each session. |
| `projects` | map | Per-project field defaults for ticket creation. Keys are project codes. Values are flat key-value pairs matching acli's `additionalAttributes` field names. |
| `parking_lots` | map | Per-project parking lot config. `epic`: the parking lot epic key. `status`: target Jira status name for parked tickets. |

## Adding a New Project

1. Run `acli jira workitem create --generate-json` with the target project key
2. Use MCP `getJiraIssueTypeMetaWithFields` for full metadata including allowed values
3. Add the project under `projects:` with its components and custom fields
4. If needed, add a `parking_lots:` entry (or let the park workflow create one)

## Consumers

The jf CLI reads `parking_lots` for the park workflow. Other keys (`cloud_id`, `projects`)
are read by skills at the agent level — they inform ticket creation and MCP calls.
