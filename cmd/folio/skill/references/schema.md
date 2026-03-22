# folio.yml Schema Reference

Shared across workflows. Contains YAML structure reference for Claude. Go code owns validation logic.

## folio.yml Schema

All paths relative to the directory containing folio.yml, unless prefixed with `vault:`.

The `vault:` prefix resolves to `~/.folio/vault/` — a shared knowledge layer outside any project. Use for cross-cutting references that multiple projects source from.

```yaml
schema: 1                              # Required. 1 or 2.
project: "Name"                        # Required. Human-readable.

sources:                               # Optional. Project-level sources.
  # Primary: local file you wrote
  - path: file.md
    depends_on: [other-source.md]    # Optional. Declares source ordering.

  # Vault: cross-cutting reference from shared vault
  - path: vault:research/2026-03-01-comparable-dvc.md

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
    how: "What/how to compose"
    blocked_by: [other-id]             # Optional DAG edges
    branch: "feature/my-branch"        # Optional. Used by stack workflow.
    pr: "123"                          # Optional. PR number for branch.
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
        how: "Per-node composition instructions"
        children:
          - id: "PROJ-200"
            label: "Project"
            file: project/README.md

cross_references:
  - fact: "Description of the fact"
    source_of_truth: "path/to/file.md § Section Name"
    also_appears_in: ["other/file.md § Section Name"]

observations: []                       # All captured items — replaces former tasks + pending.
```

The `§` separator means: read file before `§`, locate section after `§`. Some cross-references may be descriptive — do best-effort comparison.

## Schema 2 Changes

Schema 2 (`schema: 2`) introduces:

- **`observations:`** single list for all captured items (replaced former `pending:` + `tasks:`).
- **`plan` type** (alias: `brief`). `folio new plan <topic>` scaffolds `work/active/<date>-<topic>/README.md`.
- **Reference labels**: research, insight, guide, domain, review. Mapped from old names (survey→research, synthesis→research, pattern→insight).
- **`how:` optional**: Missing `how` produces a warning (data-declaration target), not an error.

`folio init` generates schema 2 by default. All projects now use `observations:` directly.
