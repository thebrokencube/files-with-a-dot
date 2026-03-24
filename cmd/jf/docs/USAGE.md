# Usage

## Quick Start

### Prerequisites

- **Node.js** — required for markdown → ADF (Atlassian Document Format) conversion
- **acli** — Atlassian CLI (`brew install acli`)
- **Jira auth** — run `acli auth login` to authenticate

Verify with:

```bash
jf setup
```

### Two Entry Paths

**Clone an existing Jira hierarchy:**

```bash
jf clone PROJ-123                    # derives sync from mutability (default)
jf clone --sync pull PROJ-123       # force pull-only (read-only forest)
cd scaffolded-directory-name
jf tree
jf sync
```

**Start a greenfield forest:**

```bash
mkdir my-effort && cd my-effort
jf init --project PROJ
# Create .md files with frontmatter (see Frontmatter Reference below)
jf push
```

## Level Progression

jf supports incremental adoption. Start at Level 0 and graduate as your needs grow.

### Level 0: Ad-hoc (no forest)

Push a single markdown file to a Jira ticket description. No `forest.yml` needed.

```bash
jf push ACME-123 notes.md
```

Pull a ticket's description to a local file:

```bash
jf pull ACME-123 output.md
```

### Level 1: Persistent forest

Create a forest to manage multiple nodes from a single directory.

```bash
jf init --project ACME
# Create feature.md with frontmatter:
#   ---
#   jira: ACME-456
#   ---
jf sync
```

### Level 2: Hierarchies and TBD nodes

Plan a hierarchy before tickets exist. Use `jira: TBD` in frontmatter, then create
them in Jira.

```bash
# Create README.md (parent) and child .md files with jira: TBD
jf create-missing --dry-run    # preview what would be created
jf create-missing              # create tickets, rewrite TBD → real keys
jf push --subtree ACME-100      # push only a branch of the tree
```

### Level 3: Bidirectional sync with conflict resolution

When others edit tickets in Jira, pull their changes and merge with yours.
Bidirectional sync is the default — no `sync:` field needed.

```bash
jf sync                        # detects conflicts, skips them by default
jf sync --resolve local        # local wins on conflict
jf sync --resolve remote       # remote wins on conflict
```

## Archetype Workflows

### (a) Ad-hoc: Single-ticket operations

**When:** You need to push a description to an existing ticket, search for tickets,
or create one ticket from a markdown file. No forest required.

**Push a description:**

```bash
jf push ACME-123 design-doc.md
```

**Search for tickets:**

```bash
jf search "login timeout" --project ACME
jf search --project ACME --type Epic
```

**What success looks like:** `✓ Pushed ACME-123 description (2048 bytes)` or search results
listing matching tickets.

**Common failure:** `✗ marklassian conversion failed` — Node.js isn't installed or the
markdown has unsupported syntax. Fix: install Node.js, or use `--plain-text` to push as plain text.

**When to graduate:** When you're managing 3+ tickets from the same directory, create a
forest with `jf init`.

### (b) Small effort: Clone and sync

**When:** You're working on an epic with a handful of children. You want to edit
descriptions locally and sync them back.

**Workflow:**

1. Clone the epic's hierarchy:
   ```bash
   jf clone ACME-100
   cd api-redesign
   ```
2. Edit markdown files in your editor
3. Check what's stale:
   ```bash
   jf status
   ```
4. Sync changes:
   ```bash
   jf sync
   ```
5. Inspect the tree:
   ```bash
   jf tree
   ```

**What success looks like:** `Pushed 5/5 nodes` followed by `Pulled 2/2 nodes`.

**Common failure:** `✗ acli: unauthorized` — auth token expired. Fix: `acli auth login`.

**When to graduate:** When you need to create new tickets within the hierarchy (TBD nodes)
or handle bidirectional edits.

### (c) Large-scale planning: Scaffold and create

**When:** You're planning a large effort — multiple epics, stories, and tasks — and want
to design the hierarchy locally before creating tickets in Jira.

**Workflow:**

1. Initialize a forest:
   ```bash
   mkdir q3-roadmap && cd q3-roadmap
   jf init --project ACME
   ```
2. Create directory structure with README.md parents and child .md files. Use
   `jira: TBD` and `type: Epic`/`Story`/`Task` in frontmatter.
3. Validate the structure:
   ```bash
   jf validate
   ```
4. Preview what would be created:
   ```bash
   jf create-missing --dry-run
   ```
5. Create tickets in Jira:
   ```bash
   jf create-missing
   ```
6. Push descriptions:
   ```bash
   jf push
   ```
7. For incremental updates to a subtree:
   ```bash
   jf push --subtree ACME-200
   ```

**What success looks like:** `Created 12/12 tickets` then `Pushed 12/12 nodes`.

**Common failure:** `✗ TBD node missing type` — add `type:` to frontmatter or set a
default in `forest.yml`. Fix: `jf validate` to see all issues.

**When to graduate:** When others start editing tickets in Jira and you need bidirectional
sync. This is the default behavior — just use `jf sync --resolve` when conflicts arise.

### (d) On-call / triage: Investigate and annotate

**When:** You're triaging an area — reviewing an epic's state, checking staleness,
adding investigation notes.

**Workflow:**

1. Clone the epic to get a local snapshot:
   ```bash
   jf clone ACME-500
   cd the-cloned-directory
   ```
2. Check forest health:
   ```bash
   jf status
   jf tree
   ```
3. Search for related tickets:
   ```bash
   jf search "cache invalidation" --project ACME
   ```
4. Edit descriptions with investigation notes in your editor
5. Push changes back:
   ```bash
   jf sync
   ```

**What success looks like:** Quick turnaround: clone → review → annotate → sync in minutes.

**Common failure:** `⚠ CONFLICT (both local and remote changed)` — someone else edited
the same ticket. Fix: `jf sync --resolve local` to keep your notes, or `--resolve remote`
to accept theirs.

## Command Quick Reference

| Command | Level | Requires Jira | `--json` | Maturity | Description |
|---------|-------|---------------|----------|----------|-------------|
| `setup` | 0 | no | yes (`--check --json`) | stable | Check prerequisites |
| `init` | 0 | no | no | stable | Create forest.yml |
| `schema` | 0 | no | n/a (outputs JSON) | stable | Emit JSON Schema |
| `version` | 0 | no | no | stable | Show version |
| `tree` | 1 | no | yes | stable | Show forest hierarchy |
| `list` | 1 | no | yes | stable | Flat list of all nodes |
| `show` | 1 | no | yes | stable | Single-node detail view |
| `status` | 1 | no | yes | stable | Forest summary with staleness |
| `validate` | 1 | no | yes | stable | Check forest integrity |
| `rm` | 1 | no | no | experimental | Remove node files |
| `push` | 2 | yes | yes | stable | Push descriptions to Jira (`--dry-run`, `--yes`, `--plain-text`) |
| `pull` | 2 | yes | yes | stable | Pull descriptions from Jira (`--dry-run`, `--yes`) |
| `sync` | 2 | yes | yes | stable | Push stale + pull pull-eligible nodes (`--dry-run`, `--yes`, `--resolve`, `--json`) |
| `create-missing` | 2 | yes | no | stable | Create Jira tickets for TBD nodes |
| `search` | 2 | yes | yes | experimental | Search Jira tickets by text |
| `clone` | 2 | yes | no | stable | Scaffold forest from Jira hierarchy (`--sync push\|pull\|both` override, default: omit sync) |

## Safety

jf enforces a 3-tier safety model on all sync operations. See SKILL.md for the full tier breakdown.

Key behaviors:
- **Empty content never pushes** — no flag or mechanism to override
- **Remote unreachable blocks push** — no proceed-anyway option
- **Conflicts require explicit resolution** — `--resolve local` or `--resolve remote`
- **First sync with content on both sides blocks** — requires interactive TTY confirmation
- **Batch operations in non-TTY mode** — require `--yes` to execute

Use `--dry-run` to preview what any operation will do before executing. Add `--json` for machine-parseable plan output.

## Recovery

If you see a state inconsistency warning after sync:
1. Run `jf sync --dry-run` to assess current state
2. Delete `.jf/state.json` to reset all baselines (next sync treats all nodes as first-sync)

### Rebuild From Scratch

When a forest is in an awkward state (broken baselines, corrupted frontmatter, stale state):

1. Note the root key from `forest.yml` or root README.md frontmatter
2. Push any local-only content you want to preserve: `jf sync`
3. Delete the local forest directory
4. Re-clone: `jf clone <ROOT_KEY> --dir <parent-dir>`

Clone pulls all descriptions fresh and establishes clean baselines. This is the fastest
recovery path — Jira is the source of truth for remote content.

## Frontmatter Reference

Every markdown file in a forest becomes a node if it has a `jira:` field in YAML frontmatter.

### Fields

| Field | Required | Values | Default | Description |
|-------|----------|--------|---------|-------------|
| `jira` | yes | `PROJ-123` or `TBD` | — | Jira key or placeholder |
| `label` | no | any string | derived | Display name for the node |
| `type` | no | Epic, Story, Task, etc. | `forest.yml defaults.type` | Jira issue type (used by create-missing) |
| `sync` | no | `push`, `pull`, `both` | derived from mutability | Sync direction override (omit for default; `both` is equivalent to omitting) |
| `order` | no | integer >= 0 | 0 (alphabetical) | Sibling sort order (lower first) |

### Example

```yaml
---
jira: ACME-456
label: API redesign
type: Epic
order: 1
---

# API Redesign

Description content here...
```

### Inheritance

Field values cascade from `forest.yml` defaults to per-node frontmatter overrides:

```
forest.yml defaults.sync → frontmatter sync: override
forest.yml defaults.type → frontmatter type: override
```

If a field is absent from both frontmatter and forest.yml defaults, it's empty (except
`sync` which defaults to `both` — the engine derives effective direction from content mutability).

### Label Derivation Chain

The node label is derived in priority order:

1. `label:` field in frontmatter (highest priority)
2. First `# heading` in the markdown content
3. Filename stem (e.g., `feature.md` → `feature`)
   - Special case: `README.md` → parent directory name

## Troubleshooting

### "No forest.yml found"

You're not inside a forest directory. Either:
- Run `jf init` to create one, or
- Check your working directory — `forest.yml` is found by walking up from `.`
- Use `--dir` to point at the right directory

### "marklassian conversion failed"

[marklassian](https://github.com/JeromeErasmus/extended-markdown-adf-parser) is the
bundled Node.js library that converts markdown to ADF. This error means conversion failed. Check:
- Is Node.js installed? (`node --version`)
- Use `--plain-text` to push as plain text (loses rich formatting)

### "acli: unauthorized" / "acli cannot reach Jira"

Auth token expired or missing:
```bash
acli auth login
jf setup    # verify all prerequisites
```

### "duplicate key ACME-123"

Two `.md` files in the forest have the same `jira:` key. Each key must be unique
(except `TBD`). Remove or re-key one of the files. Run `jf validate` to see which files
conflict.

### "CONFLICT (both local and remote changed)"

A bidirectional node was edited both locally and in Jira since the last sync. Options:
```bash
jf sync --resolve local     # keep your local version
jf sync --resolve remote    # accept Jira's version
```

### "unknown flag after positional arguments"

Flags can appear before or after positional arguments:
```bash
jf push --plain-text ACME-123 notes.md   # works
jf push ACME-123 notes.md --plain-text   # also works
```

If you see this error, the flag name is likely misspelled or unsupported.

### "TBD node missing type"

A `jira: TBD` node needs a `type:` field for ticket creation. Set it in frontmatter or
in `forest.yml` defaults:
```yaml
defaults:
  type: Story
```

## Lint and Mutability

jf restricts pushable content to a markdown subset that roundtrips cleanly
through ADF conversion. Nodes that use unsupported constructs are read-only
(pull-only).

**Supported markdown:**
- h2 headings
- Paragraphs
- Unordered and ordered lists (flat, not nested)
- Bold, italic, strikethrough, inline code
- Links (absolute URLs only)
- Horizontal rules

**Not supported (will make node read-only):**
- h1 or h3+ headings
- Tables
- Fenced code blocks
- Blockquotes
- Nested lists
- Checkboxes
- Images
- Relative links

Run `jf status` to see which nodes are mutable vs read-only and why.

To make a read-only node pushable, rewrite its content using only supported
constructs. jf will automatically detect the change and mark it mutable on the
next sync.
