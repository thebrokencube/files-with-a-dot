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

## Jira Conventions

For ticket naming, description structure, content rules, and project field defaults, see the `/jf` skill's `references/conventions.md`. All Jira standards are canonical in `/jf`.

## Jira Push Pipeline

Tree targets with `system: jira` use `jf` (Jira Forest CLI) for push operations. `folio jira push` delegates to `jf push` under the hood.

**Level 0 (single file)**:
```bash
jf push BEN-48284 epic.md
```

**Level 1 (forest-aware)**:
```bash
jf sync              # push all + pull all
jf status --json     # check what's stale
```

**folio wrapper** (adds autoTouch for folio staleness):
```bash
folio jira push --id BEN-48284 --source epic.md
```

`jf push` runs: strip frontmatter -> compile (markdown to ADF via marklassian) -> push (acli edit) -> record in `.jf/state.json`.

**Supported markdown**: tables, fenced code blocks, blockquotes, nested lists, task lists, all heading levels, bold, italic, code, links. marklassian handles the full CommonMark spec. See the `/jf` skill's `references/gotchas.md` for content pitfalls (relative links, @mentions, size limits).

## Jira Creation Pipeline

> **IMPORTANT:** Always use `folio jira` for all Jira write operations. MCP Jira tools are read-only in this workflow (permitted for field discovery where acli has no equivalent). When using MCP for reads, always pass the `fields` parameter — see tooling.yml for standard/minimal/extended field sets.

For tree nodes that do not yet have a Jira key, two approaches:

**Forest-level creation** (preferred for multiple TBD nodes):
```bash
jf create-missing --dry-run    # preview what would be created
jf create-missing              # create all TBD nodes (pre-order traversal)
```

**Single-ticket creation** (folio wrapper):
```bash
folio jira create --json /tmp/{slug}-create.json --source {source}.md
```

Both run: Create (acli, captures key) -> Push description (via `jf push`).

**Before running**, build the creation JSON per the `/jf` skill's `references/conventions.md` (Project Defaults section). Project-specific field values come from `~/.jf.yml`.

After creation, rename the source file from slug to the new ticket key and clean up the creation JSON.

## Post-Push

After a successful push, run `folio touch <target>` to update the target's local output mtime.
This clears staleness so `folio status` reflects the push. For manual pushes (Google Docs,
clipboard), remind the user to run `folio touch` after they've pasted.

**Batch orchestration**: For tree and batch targets with multiple nodes/items, iterate through
each node in order (bottom-up for trees). One review gate per node. Report progress as
`[N/total] pushed: {id}`. On failure mid-batch, report which succeeded and which remain.

## Other Publish Targets

| Target type | Publish approach |
|------------|-----------------|
| Google Docs | `manual:paste-from-markdown` — copy compiled output, paste into doc. **Review normalization**: when pulling Google Docs content for diff, strip escaped chars and normalize whitespace before comparing. |
| Slack | `mcp:slack` — send message via Slack MCP tool |
| Clipboard | `folio pbcopy <target>` — copies first local output to clipboard for manual paste |
| Confluence | `mcp:jira-confluence` — update page via Confluence MCP tool |

## Error Handling

- **Push fails mid-batch**: Report which outputs succeeded and which failed. Do not auto-retry — let the user decide whether to re-run.
- **No compiled output**: Stop before the review gate. Report "no compiled output for {target}" and suggest `/folio compose` first.
