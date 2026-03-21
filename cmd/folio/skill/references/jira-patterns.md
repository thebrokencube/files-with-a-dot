# Jira Patterns

Reference for JQL translation and bulk operations via acli.

## NL-to-JQL Translation

Translate natural language requests into JQL queries for `acli jira workitem search --jql`.

| User says | JQL |
|-----------|-----|
| "my tickets" / "assigned to me" | `assignee = currentUser() AND resolution = Unresolved` |
| "my open bugs" | `assignee = currentUser() AND type = Bug AND resolution = Unresolved` |
| "what's in progress" | `assignee = currentUser() AND status = 'In Progress'` |
| "recent updates in PROJECT" | `project = PROJECT AND updated >= -7d ORDER BY updated DESC` |
| "sprint tickets" | `sprint in openSprints() AND assignee = currentUser()` |
| "epic children" | `parent = PROJ-123 ORDER BY rank` |
| "recently updated" | `updated >= -7d ORDER BY updated DESC` |
| "by status" | `project = PROJ AND status = '<status>'` |
| "by type" | `project = PROJ AND type = <type>` |
| "text search" | `text ~ '<query>'` |
| "unresolved in project" | `project = PROJ AND resolution = Unresolved ORDER BY priority DESC` |
| "created this week" | `project = PROJ AND created >= startOfWeek()` |
| "blocked tickets" | `status = Blocked AND assignee = currentUser()` |

### Common JQL Fragments

```
# Combine fragments with AND
assignee = currentUser()
resolution = Unresolved
sprint in openSprints()
updated >= -7d
created >= startOfWeek()
parent = PROJ-123
type in (Bug, Story, Task)
status in ('To Do', 'In Progress')
labels = 'my-label'
priority in (Highest, High)
ORDER BY rank ASC
ORDER BY updated DESC
ORDER BY priority DESC, created ASC
```

## Bulk Operations

acli supports bulk operations directly via JQL -- no piping through xargs needed. Always use `--yes` to skip interactive confirmation.

### Batch Assign

```bash
acli jira workitem assign --jql "assignee = EMPTY AND priority = Highest" --assignee "user@example.com" --yes
```

### Batch Transition

```bash
acli jira workitem transition --jql "status = 'To Do' AND assignee is not EMPTY" --status "In Progress" --yes
```

### Batch Edit

```bash
acli jira workitem edit --jql "project = PROJ AND labels = old-label" --labels "new-label" --yes
```

### Batch Archive

```bash
acli jira workitem archive --jql "project = PROJ AND status = Done AND updated < -30d" --yes
```

### Batch Comment

```bash
acli jira workitem comment create --jql "project = PROJ AND status = 'In Progress'" --body "Batch update note"
```

### Batch Create from File

```bash
# From JSON
acli jira workitem create-bulk --from-json issues.json

# From CSV
acli jira workitem create-bulk --from-csv issues.csv
```
