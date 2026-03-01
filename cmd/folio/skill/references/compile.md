# Compile Workflow

Read by `/folio compile [target]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

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
    transform: distill                 # Required: distill|extract|adapt|compose
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

## Compile Steps

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

## Tree Target Compilation

Each node compiles independently from its own file. Nodes do NOT consume child outputs. Compile bottom-up (children before parents).

1. Get per-node status from `folio project status --json` (the `tree` field in target)
2. Read tree definition from folio.yml for per-node `file` and `instructions`
3. Walk bottom-up. For each stale/missing node:
   a. Read the node's `file`
   b. Apply node's `instructions` (fall back to target-level if none)
   c. If `compiled_dir`/`compiled_ext` set: use Jira Push Pipeline (see SKILL.md)
   d. Otherwise: compile and push via tooling.yml method for `tree.system`
   e. Skip push for non-system-ID nodes (descriptive slugs — report as "no external target")
4. Touch the target's local `path:` output to update mtime (if one exists)

## Batch Target Compilation

Multiple items sharing one set of instructions. Each item has its own source and output.

1. Get per-item status from `folio project status --json` (`batch_items` field)
2. For each stale/missing item:
   a. Read item's `source` file
   b. Apply target-level `instructions`
   c. Push via tooling.yml. Output: `{external: batch.system, id: item.output.id, field: item.output.field || batch.field}`
3. Touch the target's local `path:` output (if one exists)
