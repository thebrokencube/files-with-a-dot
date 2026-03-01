---
name: folio
description: Knowledge work lifecycle — research, plan, compile, audit. Replaces the plan skill's PLAN_MANIFEST.md with folio.yml and derived status.
user_invocable: true
---

# Folio

Lifecycle toolkit for knowledge work. Local source files compile into external targets (Jira descriptions, Google Docs, specs). `folio.yml` declares structure; status is derived from file mtimes.

**Two layers**: The CLI (`folio` binary) handles deterministic operations (validate, status, init, home). Claude workflows handle creative operations (compile, audit, add-pending). Workflows use CLI commands as building blocks.

## Quick Orientation

Before handling any folio request, check for a folio.yml in the current directory (or use `--folio PATH`).

| What you find | What's available |
|---|---|
| No folio.yml | No folio infrastructure needed. |
| folio.yml with local outputs only | Local compilation targets. |
| folio.yml with `external:` outputs | External system integration via co-located `tooling.yml`. |

---

## folio.yml Schema

All paths relative to the directory containing folio.yml.

```yaml
schema: 1                              # Required. Must be 1.
project: "Name"                        # Required. Human-readable.

sources:                               # Optional. Project-level sources.
  # Primary: local file you wrote
  - path: file.md

  # External: remote system resource
  - external: jira
    id: "PROJ-123"

  # Derived: local cache of external content
  - path: ref.md
    derived_from:
      - external: web
        url: "https://..."
        cached: "2026-01-15"           # YYYY-MM-DD, used for age reporting
        notes: "optional"

  # Code: repository reference
  - external: github
    id: "Org/repo"

repositories:                          # Optional. URL templates for code links.
  name: "https://github.com/Org/repo/blob/main/{path}"

targets:
  target-id:
    instructions: "What/how to transform"
    transform: distill                 # Required: distill|extract|adapt|compose|scaffold
    blocked_by: [other-id]             # Optional DAG edges
    sources:
      - path: relative/file.md
    outputs:
      - path: compiled/output.md       # Local (mtime-tracked)
      - external: jira                 # External (resolved via tooling.yml)
        id: "PROJ-456"
        field: description             # Sub-resource within external system

    # Batch variant (mutually exclusive with tree)
    batch:
      system: google_docs              # Default external for all items
      field: body                      # Default field for all items
      items:
        - id: "item-name"
          source: compiled/tab.md
          output: { id: "google-doc-id", field: "Tab Name" }

    # Tree variant (mutually exclusive with batch)
    tree:
      system: jira                     # Required
      field: description
      compiled_dir: compiled/jira/     # Workflow convention (not validated by CLI)
      compiled_ext: .json              # Workflow convention (not validated by CLI)
      root:
        id: "PROJ-100"
        label: "Initiative"
        file: README.md
        instructions: "Per-node compilation instructions"
        children:
          - id: "PROJ-200"
            label: "Project"
            file: project/README.md

cross_references:
  - fact: "Description of the fact"
    source_of_truth: "path/to/file.md § Section Name"
    also_appears_in: ["other/file.md § Section Name"]

tasks: []                              # Open work items (list of strings)
pending: []                            # Backlog items (list of strings)
```

The `§` separator means: read file before `§`, locate section after `§`. Some cross-references may be descriptive — do best-effort comparison.

---

## Tooling Resolution

External outputs resolve their push/pull method from `tooling.yml` (co-located with this skill file). Read `external:` from the target output, look up that system in tooling.yml, get the `pull`/`push` methods.

**Method types**: `cli:<tool>` = shell command, `mcp:<server>` = MCP tool call, `manual` = present to user, `manual:<hint>` = manual with guidance. Unlisted systems: pull=skip, push=manual.

### Jira Push Pipeline

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

**md-to-adf limitations** (caught by `--lint`): no tables, no fenced code blocks, no blockquotes, no nested lists, no h3+. Flatten source files before compilation.

---

## Workflows

Everything below is Claude-driven — not CLI commands. If any CLI command fails, run `folio setup --check` first.

### CLI Pass-Throughs

These slash commands run the corresponding CLI command and report results:

| Command | Runs |
|---|---|
| `/folio setup` | `folio setup` |
| `/folio status` | `folio project status` (mention `/folio compile` if stale targets exist) |
| `/folio validate` | `folio project validate` |
| `/folio init` | `folio project init --name "Name"` (ask for name if not provided) |
| `/folio home <cmd>` | `folio home <subcommand>` — run `folio home --help` for available commands |

### `/folio compile [target]`

Compile sources into targets. Compilation is distillation — sources are working memory; targets are communication condensed for their audience.

**Steps:**

1. Run `folio project validate`. Stop if invalid.
2. Run `folio project status --json`. Identify stale/missing/unknown targets (or filter to the specific target requested).
3. Read `blocked_by` edges from folio.yml (not in status JSON). Resolve compilation order: upstream targets first.
4. For each target, in DAG order:
   a. Read source files from target's `sources`
   b. Read `instructions` from folio.yml
   c. Apply the transformation
   d. **Local outputs** (`path:`): write compiled file
   e. **External outputs** (`external:`): resolve push method from tooling.yml
5. Run `folio project status` again to report final state.

**Code references**: Use `repositories` URL patterns from folio.yml for clickable links in targets that support them.

#### Tree target compilation

Each node compiles independently from its own file. Nodes do NOT consume child outputs. Compile bottom-up (children before parents).

1. Get per-node status from `folio project status --json` (the `tree` field in target)
2. Read tree definition from folio.yml for per-node `file` and `instructions`
3. Walk bottom-up. For each stale/missing node:
   a. Read the node's `file`
   b. Apply node's `instructions` (fall back to target-level if none)
   c. If `compiled_dir`/`compiled_ext` set: use [Jira Push Pipeline](#jira-push-pipeline)
   d. Otherwise: compile and push via tooling.yml method for `tree.system`
   e. Skip push for non-system-ID nodes (descriptive slugs — report as "no external target")
4. Touch the target's local `path:` output to update mtime (if one exists)

#### Batch target compilation

Multiple items sharing one set of instructions. Each item has its own source and output.

1. Get per-item status from `folio project status --json` (`batch_items` field)
2. For each stale/missing item:
   a. Read item's `source` file
   b. Apply target-level `instructions`
   c. Push via tooling.yml. Output: `{external: batch.system, id: item.output.id, field: item.output.field || batch.field}`
3. Touch the target's local `path:` output (if one exists)

### `/folio audit [scope]`

Project health check — like `git status` for the compilation system.

**Scope**: no arg or `local` = local only. `external` = also fetch and compare external targets. Specific target ID = validate just that target externally.

**Steps:**

1. Run `folio project validate`. Report errors.
2. Run `folio project status`. Report all targets.
3. For each `cross_references` entry: read source_of_truth and each also_appears_in, flag differences. Descriptive references: report as not machine-checkable.
4. (External scope only) Fetch external targets via tooling.yml pull method, compare against local. For format differences (ADF vs markdown), compare structural elements not literal text.

**Output:**
```
## Status
- [target-id]: clean / stale / missing / unknown

## Cross-References
- [fact]: consistent / warning / error

## External Validation (if requested)
- [system] [id]: matches / differs
```

Audit only reports. It does not fix anything.

### `/folio add-pending`

Add item to the `pending` list in folio.yml.

1. If text provided with command, use it. Otherwise ask.
2. Read folio.yml, locate or create `pending:` list
3. Append new item string
4. Write with targeted editing (don't reformat the whole file)
