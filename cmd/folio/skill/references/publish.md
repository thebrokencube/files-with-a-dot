# Publish Workflow

Read by `/folio publish [target]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

Publishing sends composed output to external systems. It is the final step in the knowledge graduation chain: gather -> compose -> publish.

## Steps

1. Run `folio validate`. Stop if invalid.
2. Identify the target to publish. If no target specified, list targets with external outputs.
3. For each external output on the target:
   a. Resolve the push method from tooling.yml (`external:` field -> system -> push method)
   b. If local compiled output exists, read it as the content to push
   c. Execute the push via the resolved method
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

## Other Publish Targets

| Target type | Publish approach |
|------------|-----------------|
| Google Docs | `manual:paste-from-markdown` — copy compiled output, paste into doc |
| Slack | `mcp:slack` — send message via Slack MCP tool |
| Clipboard | `folio pbcopy <target>` — copies first local output to clipboard for manual paste |
| Confluence | `mcp:jira-confluence` — update page via Confluence MCP tool |
