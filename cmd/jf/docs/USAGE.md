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
jf clone PROJ-123
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

```bash
# In forest.yml or per-node frontmatter: sync: both
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
markdown has unsupported syntax. Fix: install Node.js, or use `--force` to push as plain text.

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
sync — set `sync: both` and use `jf sync --resolve`.

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
| `discover` | 1 | no | yes | stable | Discover nodes and show tree |
| `tree` | 1 | no | no | stable | Show forest hierarchy |
| `list` | 1 | no | yes | stable | Flat list of all nodes |
| `show` | 1 | no | yes | stable | Single-node detail view |
| `status` | 1 | no | yes | stable | Forest summary with staleness |
| `validate` | 1 | no | yes | stable | Check forest integrity |
| `rm` | 1 | no | no | experimental | Remove node files |
| `push` | 2 | yes | no | stable | Push descriptions to Jira |
| `pull` | 2 | yes | no | stable | Pull descriptions from Jira |
| `sync` | 2 | yes | no | stable | Push stale + pull pull-mode nodes |
| `create-missing` | 2 | yes | no | stable | Create Jira tickets for TBD nodes |
| `search` | 2 | yes | no | experimental | Search Jira tickets by text |
| `clone` | 2 | yes | no | stable | Scaffold forest from Jira hierarchy |

## Frontmatter Reference

Every markdown file in a forest becomes a node if it has a `jira:` field in YAML frontmatter.

### Fields

| Field | Required | Values | Default | Description |
|-------|----------|--------|---------|-------------|
| `jira` | yes | `PROJ-123` or `TBD` | — | Jira key or placeholder |
| `label` | no | any string | derived | Display name for the node |
| `type` | no | Epic, Story, Task, etc. | `forest.yml defaults.type` | Jira issue type (used by create-missing) |
| `sync` | no | `push`, `pull`, `both` | `forest.yml defaults.sync` | Sync direction |
| `order` | no | integer >= 0 | 0 (alphabetical) | Sibling sort order (lower first) |

### Example

```yaml
---
jira: ACME-456
label: API redesign
type: Epic
sync: both
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
`sync` which defaults to `push` in `forest.yml` template from `jf init`).

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
- Use `--force` to push as plain text (loses rich formatting)

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

A `sync:both` node was edited both locally and in Jira since the last sync. Options:
```bash
jf sync --resolve local     # keep your local version
jf sync --resolve remote    # accept Jira's version
```

### "unknown flag after positional arguments"

jf requires flags before positional arguments:
```bash
# Wrong:
jf push ACME-123 notes.md --force

# Right:
jf push --force ACME-123 notes.md
```

### "TBD node missing type"

A `jira: TBD` node needs a `type:` field for ticket creation. Set it in frontmatter or
in `forest.yml` defaults:
```yaml
defaults:
  type: Story
```
