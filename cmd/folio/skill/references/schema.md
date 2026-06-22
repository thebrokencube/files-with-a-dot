# folio.yml Schema Reference

Shared across workflows. Contains YAML structure reference for Claude. Go code owns validation logic.

## folio.yml Schema

All paths relative to the directory containing folio.yml, unless prefixed with a registered `<store>:` name.

The `vault:` prefix resolves to the **active store's** `vault/` — a shared knowledge layer outside any project, folio-local to that store (in single-home mode, `~/.folio/vault/`). Use for cross-cutting references that multiple projects source from. `vault` is **not** a registered store; it is intrinsic to whichever folio store contains the current project (see **stores.yml** below). A registered `<store>:` name is referenced the same way.

## stores.yml — the store registry (multi-store container)

`~/.folio` is a plain **umbrella directory** (NOT a repo) that physically **contains** each store as an independent git repo nested as a sibling, dir-named by its remote repo name. A `~/.folio/stores.yml` registers every store plus the **default**.

```yaml
schema: 2                              # container model: default + nested stores
default: <work-store>           # acted on when invoked from the umbrella with no --folio
stores:
  <work-store>: { path: ~/.folio/<work-store>, kind: folio,    remote: git@github.com:<org>/<work-folio>.git }
  folio-vault:         { path: ~/.folio/folio-vault,         kind: folio,    remote: git@github.com:thebrokencube/folio-vault.git }
  adr:                 { path: ~/.folio/adr,                 kind: external, remote: <adr-remote> }
```

- **`default:`** — the store folio acts on when invoked from the umbrella with no `--folio` and no `folio.yml` in cwd. **cwd inside a registered store always overrides the default.**
- **`kind: folio`** — a full folio home: listed, structure-aware in `find`, writable (`--folio <store>:<project>`), validated. `folio home push/pull <store>` sync it to its own remote.
- **`kind: external`** — a non-folio KB you read from (ADRs, RADRs, wikis, docs): content-grep in `find`, **read-only**, never scanned for folio structure; a missing target **warns**, never errors. **External stores are pullable (`folio home pull <store>`) but NEVER pushed** — contributions go through that repo's own PR flow.
- **`remote:`** — informational (the store's own repo handles its remote); records where each store clones from for the migration runbook.

**`vault` is folio-LOCAL, never a registered store.** A folio store MAY have its own `vault/` subdir; `vault:` resolves relative to the **active store's** `vault/` (e.g. from inside `<work-store>`, `vault:` → `~/.folio/<work-store>/vault`). There is no global vault and no `vault` registry entry.

**`stores.yml` is dotfile-managed**, not hand-written: it is the base+private merge target `configs/base/folio/stores.base.yml` (personal `folio-vault` + `default: folio-vault`) merged with `~/.dotfiles.private/folio-stores.yml` (the work store + a `default` override) → `~/.folio/stores.yml` via `dot sync`. Base alone is valid (a personal one-store container), so the work overlay is purely additive.

**Absent `stores.yml` ⇒ implicit empty registry** — legacy single-home behavior, byte-for-byte unchanged. This is the transitional bridge for an un-migrated machine; see `references/container-migration.md` to migrate.

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
