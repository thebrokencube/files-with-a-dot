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

Tree targets with `system: jira` and `compiled_ext: .json` use a three-phase pipeline:

```
source .md -> lint (md-to-adf --lint) -> precompile (md-to-adf --acli) -> compiled .json -> push (acli)
```

| Placeholder | Resolves from |
|---|---|
| `{id}` | Tree node `id` (Jira key) |
| `{source}` | Tree node `file` |
| `{compiled}` | `{compiled_dir}/{id}{compiled_ext}` |

Example:
```bash
md-to-adf --lint epic.md                                               # 1. Lint
md-to-adf --acli BEN-48284 < epic.md > compiled/jira/BEN-48284.json   # 2. Precompile
acli jira workitem edit --from-json compiled/jira/BEN-48284.json --yes # 3. Push
```

**md-to-adf limitations** (caught by `--lint`): no tables, no fenced code blocks, no blockquotes, no nested lists, no h3+. Flatten source files before composition.

## Jira Creation Pipeline

> **IMPORTANT:** Always use `acli` + `md-to-adf` for all Jira write operations. Never use MCP tools (`jira-confluence`) for Jira creates or edits. Never hand-write ADF.

For tree nodes that do not yet have a Jira key, use the creation pipeline instead of the push pipeline:

```
source .md -> lint (md-to-adf --lint) -> convert (md-to-adf, raw) -> build creation JSON -> create (acli) -> capture key -> rename outputs
```

| Placeholder | Resolves from |
|---|---|
| `{source}` | Tree node `file` |
| `{compiled_dir}` | Tree `compiled_dir` |

1. **Lint**: `md-to-adf --lint {source}` (same as push pipeline)
2. **Convert to raw ADF**: `md-to-adf < {source}` — raw mode (no `--acli` flag) produces the ADF document object directly, suitable for embedding in the creation JSON `description` field
3. **Build creation JSON**: Assemble the creation payload and write to `{compiled_dir}/{slug}-create.json`:
   ```json
   {
     "projectKey": "BEN",
     "type": "Story",
     "summary": "Title from source",
     "parentIssueId": "BEN-12345",
     "description": { "...raw ADF from step 2..." },
     "additionalAttributes": {
       "components": [{ "name": "Component" }],
       "customfield_NNNNN": "value"
     }
   }
   ```
   `parentIssueId` is optional for top-level issues. Discover required fields with `acli jira workitem create --generate-json`.
4. **Create**: `acli jira workitem create --from-json {compiled_dir}/{slug}-create.json`
5. **Capture key**: Parse the returned Jira key (e.g., `BEN-54321`) from acli stdout
6. **Rename outputs**: Rename source and compiled files from descriptive slug to ticket key. Delete the intermediate `-create.json` after successful creation. Output files are always named by ticket key, not descriptive names:
   ```bash
   mv output/jira/{slug}.md output/jira/BEN-54321.md
   rm {compiled_dir}/{slug}-create.json
   ```
7. **Generate edit JSON**: Run `md-to-adf --acli BEN-54321 < output/jira/BEN-54321.md > {compiled_dir}/BEN-54321.json` to produce the standard push-pipeline format for future edits

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
