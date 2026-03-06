# Publish Workflow

Read by `/folio publish [target]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

Publishing sends composed output to external systems. It is the final step in the knowledge graduation chain: gather -> compose -> publish.

## Steps

1. Run `folio validate`. Stop if invalid.
2. Identify the target to publish. If no target specified, list targets with external outputs.
3. For each external output on the target:
   a. Resolve the push method from tooling.yml (`external:` field -> system -> push method)
   b. If local compiled output exists, read it as the content to push. If no compiled output exists, stop and report "no compiled output for {target}" — do not proceed to the gate.
   c. **Review gate (hard)**: Present target ID, external system, push method, and first 5 lines of content. Wait for explicit "yes" before pushing. One gate per external output.
   d. Execute the push via the resolved method
4. Report what was published and where.

## Publish Methods

Resolved from tooling.yml (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `push` method.

| Method | What happens |
|--------|-------------|
| `cli:<tool>` | Run a shell command (e.g., `acli`, `gh`) |
| `mcp:<server>` | Call an MCP tool (e.g., `jira-confluence`, `gdrive`) |
| `manual` | Present content to user for manual action |
| `manual:<hint>` | Manual with guidance (e.g., `manual:paste-from-markdown`) |

Unlisted systems: push = manual (present to user).

## Jira Push Pipeline

Tree targets with `system: jira` use `folio jira` commands. A single command handles lint, conversion, and push:

```bash
folio jira push --id BEN-48284 --source epic.md
```

This runs: Lint (gate) -> Convert (markdown to ADF) -> Push (acli edit). Stops on lint failure.

To inspect the intermediate ADF JSON without pushing:

```bash
folio jira compile --id BEN-48284 --source epic.md --output compiled/jira/BEN-48284.json
```

**Markdown constraints** (enforced by lint): no tables, no fenced code blocks, no blockquotes, no nested lists, no h3+, no images, no bare `[` or relative links. Only `##` headings, flat lists, paragraphs, and inline bold/code/links. Flatten source files before composition.

## Jira Creation Pipeline

> **IMPORTANT:** Always use `folio jira` for all Jira write operations. MCP Jira tools are read-only in this workflow (permitted for field discovery where acli has no equivalent).

For tree nodes that do not yet have a Jira key, use the two-phase creation command. It creates a barebones ticket (no description), captures the key, then pushes the description separately. This two-phase approach is more reliable than embedding ADF in the creation payload.

```bash
folio jira create --json /tmp/{slug}-create.json --source {source}.md
```

This runs: Lint (gate) -> Create (acli, captures key) -> Compile (using new key) -> Push (description).

**Before running**, build the creation JSON. **The creation JSON must NOT contain a `description` field** — `acli create` silently drops or malforms inline ADF descriptions.

```json
{
  "projectKey": "BEN",
  "type": "Story",
  "summary": "Title from source",
  "parentIssueId": "BEN-12345",
  "additionalAttributes": {
    "components": [{ "name": "Component" }],
    "customfield_NNNNN": "value"
  }
}
```

Write to `/tmp/{slug}-create.json`. `parentIssueId` is optional for top-level issues. Discover required fields with `acli jira workitem create --generate-json`. For BEN project-specific field values, see the jira skill.

After creation, rename the source file from slug to the new ticket key and clean up the creation JSON.

## Other Publish Targets

| Target type | Publish approach |
|------------|-----------------|
| Google Docs | `manual:paste-from-markdown` — copy compiled output, paste into doc |
| Slack | `mcp:slack` — send message via Slack MCP tool |
| Clipboard | `folio pbcopy <target>` — copies first local output to clipboard for manual paste |
| Confluence | `mcp:jira-confluence` — update page via Confluence MCP tool |

## Error Handling

- **Push fails mid-batch**: Report which outputs succeeded and which failed. Do not auto-retry — let the user decide whether to re-run.
- **No compiled output**: Stop before the review gate. Report "no compiled output for {target}" and suggest `/folio compose` first.
