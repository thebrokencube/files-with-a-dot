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

After a successful push, verify the published artifact and report the external
system's result. When the verified composition has a local output, run
`folio touch <target>` to record its input digest without changing the output.
Use `folio touch <target> --final` only for a non-batch, non-forest target with a
direct external output and an existing local review copy. Final is provider-neutral;
it records operator confirmation, not a provider-specific status.

**Batch orchestration**: For batch targets with multiple items, iterate through
each item in order. One review gate per item. Report progress as
`[N/total] pushed: {id}`. On failure mid-batch, report which succeeded and which remain.

## Notion Templates

Proposal structure and its Feedback table are applied at compose time (`references/proposal.md`), so the
output file already contains them. Publish pushes that file unchanged.

**Notion targets are write-only.** All edits happen in the local output file. Publish
replaces the full page content. Never use `update_content` for incremental Notion edits —
compose the final version locally, then publish once with `replace_content`.

The iterative loop stays local:

```
align on changes → edit output file → review → repeat → publish once
```

Do not edit Notion directly during the review loop. The local output file is the source
of truth; Notion is a render target.

### Notion rendering constraints

Each of these has cost a fetch-fix-push cycle. Always fetch the page back and confirm it rendered.

- **Pass real newline and tab characters**, never `\n` / `\t` escapes. Escaped, they reach Notion's markdown
  engine as backslash-escapes, get stripped to bare `n`/`t`, and the whole body renders as one paragraph with
  visible `<table>` markup. Applies to `notion-create-pages` `content` and `notion-update-page`
  `replace_content` / `new_str` / `content_updates`.
- **Tables:** `<tr>`/`<td>` only — a `<thead>` or `<tbody>` collapses the header into the first data row.
- **Bold:** `**markdown**`, not `<strong>`, which renders as raw HTML.
- **Empty cells** need `<td> </td>` with a space; `<td></td>` merges columns.

## Other Publish Targets

| Target type | Publish approach |
|------------|-----------------|
| Google Docs | `manual:paste-from-markdown` — copy compiled output, paste into doc. **Review normalization**: when pulling Google Docs content for diff, strip escaped chars and normalize whitespace before comparing. |
| Slack | `mcp:slack` — send message via Slack MCP tool |
| Clipboard | `pbcopy < <output-file>` — copy compiled output to clipboard for manual paste |
| Confluence | `mcp:jira-confluence` — update page via Confluence MCP tool |

## Error Handling

- **Push fails mid-batch**: Report which outputs succeeded and which failed. Do not auto-retry — let the user decide whether to re-run.
- **No compiled output**: Stop before the review gate. Report "no compiled output for {target}" and suggest `/folio compose` first.
