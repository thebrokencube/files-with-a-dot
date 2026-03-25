# Reference

## Command Reference

| Command | Level | Jira | `--json` | Description |
|---------|-------|------|----------|-------------|
| `setup` | 0 | no | yes | Check prerequisites |
| `init` | 0 | no | no | Create `.jf/forest.yml` |
| `schema` | 0 | no | n/a | Emit JSON Schema |
| `version` | 0 | no | no | Show version |
| `tree` | 1 | no | yes | Show forest hierarchy |
| `list` | 1 | no | yes | Flat list of all nodes |
| `show` | 1 | no | yes | Single-node detail view |
| `status` | 1 | no | yes | Forest summary with staleness |
| `validate` | 1 | no | yes | Check forest integrity |
| `rm` | 1 | no | no | Remove node files (experimental) |
| `push` | 2 | yes | yes | Push descriptions to Jira |
| `pull` | 2 | yes | yes | Pull descriptions from Jira |
| `sync` | 2 | yes | yes | Bidirectional push + pull |
| `create-missing` | 2 | yes | no | Create Jira tickets for TBD nodes |
| `search` | 2 | yes | yes | Search Jira tickets (experimental) |
| `clone` | 2 | yes | no | Scaffold forest from Jira hierarchy |
| `view` | 2 | yes | yes | Fetch remote issue details from Jira |

Common flags: `--dry-run`, `--yes` (batch mode), `--dir` (working directory), `--json` (structured output).

### Maturity

Most commands are stable. Exceptions:

- **`search`** -- experimental. Thin JQL wrapper, interface may change.
- **`rm`** -- experimental. No cascade, no confirmation prompt.
- **`clone --depth`** -- experimental. Limits recursive child fetching (default 0 = unlimited).
- **Park workflow** -- experimental. Skill-level operation (not a CLI command). Moves a ticket to a parking lot by combining `jf rm` with Jira transitions via MCP.

## Frontmatter Reference

Every `.md` file in `.jf/` becomes a node if it has a `jira:` field in YAML frontmatter.

### Fields

| Field | Required | Values | Default | Description |
|-------|----------|--------|---------|-------------|
| `jira` | yes | `PROJ-123` or `TBD` | -- | Jira key or placeholder |
| `label` | no | any string | derived | Display name |
| `type` | no | Epic, Story, Task, etc. | `forest.yml defaults.type` | Issue type (for create-missing) |
| `sync` | no | `push`, `pull`, `both` | derived from mutability | Sync direction override |
| `order` | no | integer >= 0 | 0 (alphabetical) | Sibling sort order |

### Example

```yaml
---
jira: ACME-456
label: API redesign
type: Epic
order: 1
---

## API Redesign

Description content here...
```

### Inheritance

Field values cascade: `forest.yml defaults` -> per-node frontmatter override. If absent from both, `sync` defaults to `both` (derives effective direction from content mutability).

### Label Derivation

Priority order:

1. `label:` in frontmatter
2. First `# heading` in content
3. Filename stem (`feature.md` -> `feature`; `README.md` -> parent directory name)

## forest.yml Schema

```yaml
schema: 1
defaults:
  sync: both          # push | pull | both (optional, defaults to "both")
  type: Story         # Jira issue type for create-missing
  project: ACME       # Jira project key
  field: description  # description | comment
```

| Field | Required | Description |
|-------|----------|-------------|
| `schema` | yes | Schema version (currently `1`) |
| `defaults.sync` | no | Default sync direction for all nodes |
| `defaults.type` | no | Default issue type for `create-missing` |
| `defaults.project` | no | Default Jira project key |
| `defaults.field` | no | Which Jira field to push to (`description` or `comment`) |

## Lint and Mutability

jf restricts pushable content to a markdown subset that roundtrips cleanly through ADF conversion. Nodes using unsupported constructs become read-only (pull-only).

**Supported (pushable):**

- h2 headings, paragraphs
- Flat unordered and ordered lists
- Bold, italic, strikethrough, inline code
- Links (absolute URLs only)
- Horizontal rules

**Unsupported (makes node read-only):**

- h1 or h3+ headings
- Tables, fenced code blocks, blockquotes
- Nested lists, checkboxes, images, relative links

Run `jf status` to see which nodes are mutable vs read-only:

```
Effective direction:
  3 push+pull (mutable)
  2 pull-only (2 read-only demoted)
  1 empty
```

To make a read-only node pushable, rewrite its content using only supported constructs. jf detects the change automatically on next sync.

## Safety Model

jf enforces a safety model on sync operations:

- **Empty content never pushes** -- no override
- **Remote unreachable blocks push** -- no proceed-anyway option
- **Conflicts require explicit resolution** -- `--resolve local` or `--resolve remote`
- **First sync with content on both sides blocks** -- requires TTY confirmation
- **Batch operations in non-TTY mode** -- require `--yes` to execute

Use `--dry-run` to preview any operation before executing.

## Troubleshooting

### "No forest.yml found"

You're not inside a forest's working directory. Either:

- Run `jf init` to create `.jf/forest.yml`
- Use `--dir` to point at the working directory (parent of `.jf/`)
- Check that `.jf/forest.yml` exists -- `FindForest()` walks up from your current directory

### "marklassian conversion failed"

The markdown -> ADF (Atlassian Document Format) conversion failed. marklassian is the converter library. Check:

- Is Node.js installed? (`node --version`)
- Use `--plain-text` to push as plain text (loses rich formatting)

### "acli: unauthorized"

Auth token expired: `acli auth login`, then `jf setup` to verify.

### "duplicate key ACME-123"

Two `.md` files in `.jf/` have the same `jira:` key. Run `jf validate` to see which files conflict.

### "CONFLICT (both local and remote changed)"

A bidirectional node was edited both locally and in Jira since last sync:

```bash
jf sync --resolve local        # keep your version
jf sync --resolve remote       # accept Jira's version
```

### "TBD node missing type"

A `jira: TBD` node needs a `type:` field. Set it in frontmatter or in `forest.yml` defaults:

```yaml
defaults:
  type: Story
```
