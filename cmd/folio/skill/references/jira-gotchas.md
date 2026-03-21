# Jira Gotchas

Universal Jira integration pitfalls. Applies to any skill or workflow that touches Jira.

## Content Rendering

- **@mentions don't resolve via ADF API.** The ADF `mention` node requires an `accountId`, not a display name. Plain text workaround: write `@display-name` as literal text instead of a mention node.
- **Relative links break in Jira.** Jira renders links as-is -- `./foo.md` won't resolve. Always use absolute URLs.
- **Long descriptions may truncate.** Keep under ~32KB of ADF JSON. If your markdown compiles to more, split into a summary description + linked child tickets or attachments.

## MCP Token Optimization

- **Always use `responseContentFormat: "markdown"`** on `getJiraIssue`, `searchJiraIssuesUsingJql`, `editJiraIssue`. Est. 60-70% token reduction over default ADF responses.
- **Always pass the `fields` parameter** on MCP reads. A bare `getJiraIssue` returns ~100KB (~25-30K tokens), 91% waste. Standard set: `summary,status,assignee,description,priority,issuetype,parent,subtasks,issuelinks,labels,comment`.

## acli Quirks

- **`acli` flat JSON format differs from REST API nested format.** Fields like `assignee` are flat strings in acli output, not nested objects. Don't mix acli JSON with REST API expectations.
- **Story points can't be edited via acli after creation.** The `story_points` field is read-only through `acli edit`. Use MCP `editJiraIssue` or update via the REST API directly.
- **Components require `additionalAttributes` in JSON, no CLI flag.** When creating tickets with components, use the JSON payload: `"additionalAttributes": {"components": [{"name": "Component"}]}`.
- **`acli create` silently drops inline ADF descriptions.** Always use two-phase creation: create the ticket (no description), then push the description separately via `acli edit`.

## Field Discovery

- **Use MCP `getJiraIssueTypeMetaWithFields` for custom field discovery.** acli's `--generate-json` gives you the creation template, but MCP gives you the full field metadata including allowed values and required fields.
