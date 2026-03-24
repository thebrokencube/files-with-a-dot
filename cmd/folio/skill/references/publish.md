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

Forest targets use `jf` (Jira Forest CLI) for all push operations.

**Level 0 (single file)**:
```bash
jf push PROJ-123 foundations/package-imports/README.md
```

**Level 1 (forest-aware)**:
```bash
jf sync              # push all stale + pull all pull-eligible nodes
jf status --json     # check what's stale
```

`jf push` runs: strip frontmatter -> compile (markdown to ADF via marklassian) -> push (acli edit) -> record in `.jf/state.json`.

**Pushable markdown**: h2 headings, paragraphs, flat lists, bold, italic, strikethrough, inline code, absolute links, horizontal rules. Content using unsupported constructs (tables, code blocks, nested lists, h1/h3+) is read-only and demoted to pull-only. See the `/jf` skill's USAGE.md "Lint and Mutability" section for the full list.

## Jira Creation Pipeline

For TBD nodes (frontmatter `jira: TBD`), use jf to create the Jira tickets:

```bash
jf create-missing --dry-run    # preview what would be created
jf create-missing              # create all TBD nodes (pre-order traversal)
```

**Before running**, build the creation JSON per the `/jf` skill's `references/conventions.md` (Project Defaults section). Project-specific field values come from `~/.jf.yml`.

## Post-Push

After a successful push, run `folio touch <target>` to update the target's local output mtime.
This clears staleness so `folio status` reflects the push. For manual pushes (Google Docs,
clipboard), remind the user to run `folio touch` after they've pasted.

**Batch orchestration**: For batch targets with multiple items, iterate through
each item in order. One review gate per item. Report progress as
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
