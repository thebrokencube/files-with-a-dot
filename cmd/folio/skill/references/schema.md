# folio.yml Schema Reference

Shared across workflows. Contains YAML structure reference for Claude. Go code owns validation logic.

## folio.yml Schema

All paths relative to the directory containing folio.yml, unless prefixed with a registered `<store>:` name.

The `vault:` prefix resolves to `~/.folio/vault/` — a shared knowledge layer outside any project. Use for cross-cutting references that multiple projects source from. `vault` is simply the implicitly-registered store (see **stores.yml** below); any other registered store can be referenced the same way.

## stores.yml — the store registry (multi-store)

`~/.folio` is the home of all folios + KBs. A global `~/.folio/stores.yml` indexes every folio and external KB you work across. It is a **logical index** — registered stores stay in their own repos; nothing physically moves.

```yaml
schema: 1
stores:
  vault: { path: ~/.folio/vault,              kind: folio }     # the shared personal vault (implicit if file absent)
  work:  { path: ~/work-folio,                kind: folio }     # a second folio home
  radr:  { path: ~/workspace/guideline/radrs, kind: external }  # an external KB (read-only), NOT a folio
```

- **`kind: folio`** — a full folio home: listed, structure-aware in `find`, writable (`--folio <store>:<project>`), validated.
- **`kind: external`** — a non-folio KB you read from (ADRs, RADRs, wikis, docs): content-grep in `find`, **read-only**, never scanned for folio structure; a missing target **warns**, never errors. ADR/RADR are not special types — they are just `external` stores.

**Absent `stores.yml` ⇒ implicit `{vault: {<home>/vault, folio}}`** — single-home behavior, byte-for-byte unchanged.

**`<store>:` references**: `<store>:<path-within-store>`. A registered prefix resolves against that store's root; an unknown store-shaped prefix (`bogus:foo.md`) fails loud; a path that merely contains a colon (`a/b:c.md`) resolves normally. List the registry with `folio stores list [--json]`.

**Write-routing**: `--folio <store>:<project>` targets a project in any folio store (matched in its `active/` or `archive/`). External stores are not write targets.

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

    # Forest variant (mutually exclusive with batch) — jf-managed Jira hierarchy
    forest:
      root: work/active/YYYY-MM-DD-topic   # Path to forest root (contains forest.yml)
      how_default: "Default compilation instruction for all nodes."
      how_overrides:
        PROJ-100: "Initiative overview: goals, approach, and summary."
        PROJ-200: "Epic overview: context, requirements, and scope."

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
- **`plan` type** (alias: `brief`). `folio new plan <topic>` scaffolds `work/active/<date>-<topic>/README.md`. If a work dir already exists for the topic (e.g., from a prior `folio new design`), the README.md is added to it instead of creating a new dir.
- **`design` type** scaffolds inside a work directory: `folio new design <topic>` creates `work/active/<date>-<topic>/reference/design/<date>-<topic>.md`, auto-creating the work dir if it doesn't exist. Designs always live in work directories — not at project-level `reference/design/`.
- **Reference labels**: research, insight, guide, domain, review. Mapped from old names (survey→research, synthesis→research, pattern→insight).
- **`how:` optional**: Missing `how` produces a warning (data-declaration target), not an error.

`folio init` generates schema 2 by default. All projects now use `observations:` directly.
