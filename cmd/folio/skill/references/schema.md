# folio.yml Schema Reference

Shared across workflows. Contains YAML structure reference for Claude. Go code owns validation logic.

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
