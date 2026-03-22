# Jira Gotchas

Common pitfalls when working with Jira via jf, acli, and MCP tools. For jf-specific
troubleshooting (error messages and fixes), see
[docs/USAGE.md](../docs/USAGE.md#troubleshooting).

## Content Rendering

- **Relative links**: Jira ADF doesn't support relative links. Use absolute URLs.
- **@mentions**: The ADF `mention` node requires an `accountId`, not a display name. Plain text workaround: write `@display-name` as literal text instead of a mention node.
- **Description size**: Jira descriptions have a ~32KB ADF limit. Large documents may silently truncate. Split into a summary description + linked child tickets or attachments.
- **Images**: Inline images in markdown won't render in Jira. Upload attachments separately and reference them.
- **HTML in markdown**: Raw HTML passes through marklassian but Jira strips most HTML tags.

## MCP Token Optimization

- **Always use `responseContentFormat: "markdown"`** on `getJiraIssue`, `searchJiraIssuesUsingJql`, `editJiraIssue`. Est. 60-70% token reduction over default ADF responses.
- **Always pass the `fields` parameter** on MCP reads. A bare `getJiraIssue` returns ~100KB (~25-30K tokens), 91% waste. Standard set: `summary,status,assignee,description,priority,issuetype,parent,subtasks,issuelinks,labels,comment`.

## acli Quirks

- **`acli create` silently drops inline ADF descriptions.** Always use two-phase creation: create the ticket (no description), then push the description separately via `acli edit` or `jf push`. This is what `jf create-missing` does.
- **`acli` flat JSON format differs from REST API nested format.** Fields like `assignee` are flat strings in acli output, not nested objects. Don't mix acli JSON with REST API expectations.
- **Story points can't be edited via acli after creation.** The `story_points` field is read-only through `acli edit`. Use MCP `editJiraIssue` or the REST API directly.
- **Components require `additionalAttributes` in JSON, no CLI flag.** When creating tickets with components, use the JSON payload: `"additionalAttributes": {"components": [{"name": "Component"}]}`.
- **Rate limiting**: Jira Cloud rate limits at ~100 requests/minute. For large forests, expect throttling during bulk sync.
- **Auth token scope**: `JIRA_API_TOKEN` must have write access to the target project. Read-only tokens fail silently on push.

## Field Discovery

- **Use MCP `getJiraIssueTypeMetaWithFields` for custom field discovery.** acli's `--generate-json` gives you the creation template, but MCP gives you the full field metadata including allowed values and required fields. Some projects have mandatory custom fields (e.g., T-Shirt Size estimates) that must be included via `additional_fields` on creation.
