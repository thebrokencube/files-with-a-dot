# Configuration

`~/.jf.yml` is the persistent config store for jf. It holds per-environment values that don't belong in generic skill docs — cloud IDs, project field defaults, parking lot settings.

## Schema

```yaml
# Atlassian Cloud ID for MCP tool calls
# Discover via getAccessibleAtlassianResources if unknown
cloud_id: "your-cloud-id-here"

# Project-specific field defaults for ticket creation
# Used by conventions.md creation JSON templates
projects:
  PROJ:
    components: ["My Component"]
    some_custom_field: "Required Value"   # customfield_NNNNN
  OTHER: {}

# Parking lot epics for lifecycle/park workflow
# See references/lifecycle.md
parking_lots:
  PROJ:
    epic: PROJ-999
    status: Backlog
```

## Keys

| Key | Type | Purpose |
|-----|------|---------|
| `cloud_id` | string | Atlassian Cloud ID for all MCP Jira tool calls. Pass directly — do not call `getAccessibleAtlassianResources` each session. |
| `projects` | map | Per-project field defaults. Keys are project codes (e.g., `PROJ`). Values are flat key-value pairs matching acli's `additionalAttributes` field names. |
| `parking_lots` | map | Per-project parking lot config. `epic`: the parking lot epic key. `status`: target Jira status name for parked tickets (e.g., `"Backlog"`). |

## Adding a New Project

1. Run `acli jira workitem create --generate-json` with the target project key to discover required fields
2. Use MCP `getJiraIssueTypeMetaWithFields` for full metadata including allowed values
3. Add the project under `projects:` with its components and custom fields
4. If the project needs lifecycle management, add a `parking_lots:` entry (or let the park workflow create one automatically)

## Consumers

The jf CLI currently reads `parking_lots` for the park workflow. Other keys (`cloud_id`, `projects`) are read by skills at the agent level — they inform ticket creation and MCP calls but don't require CLI support yet.
