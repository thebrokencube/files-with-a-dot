# folio.yml Schema Reference

Shared across workflows. Contains the YAML structure reference for the Agent Skill. Go code owns validation logic.

## folio.yml Schema

All unprefixed paths are relative to the directory containing folio.yml. A registered `<store>:` prefix resolves from that store's canonical root.

The `vault:` prefix resolves to the owning Folio store's `vault/` when a project context is present, then to the selected content work root for store-only operations. A literal registered store named `vault` takes precedence. In legacy isolated-home mode it resolves below the isolated `FOLIO_HOME` root. The vault has no folio.yml.

## stores.yml — the store registry (multi-store container)

`FOLIO_UMBRELLA` is the plain **control-plane directory** that owns `stores.yml`; it is not a content store. Each registry entry names an independent store repository. `FOLIO_HOME` is the selected content work root for a command and may be a session workspace.

```yaml
schema: 2                              # Container registry schema
default: <work-store>                  # Default Folio store outside a store cwd
stores:
  <work-store>: { path: ~/.folio/<work-store>, kind: folio, remote: git@github.com:<org>/<work-folio>.git }
  folio-vault:         { path: ~/.folio/folio-vault, kind: folio, remote: git@github.com:thebrokencube/folio-vault.git }
  adr:                 { path: ~/.folio/adr, kind: external, remote: <adr-remote> }
```

- **`default:`** — the Folio store selected outside registered store roots when no explicit `FOLIO_HOME` is supplied. A cwd inside a registered Folio store overrides it.
- **`kind: folio`** — a full Folio content store: listed, structure-aware in `find`, writable, and syncable through `folio home push/pull`.
- **`kind: external`** — a non-Folio KB used by `find`; it is read-only to Folio and pullable through `folio home pull`.
- **`kind: code`** — a code repository used by fleet/workarea operations; it is never a Folio project root.
- **`kind: dot`** — the dotfiles repository; it is managed by `dot`, not Folio project commands.
- **`remote:`** — informational; the store's own VCS handles synchronization.

`stores.yml` is private and per-machine. Keep it in the private dotfiles overlay and expose it at `<umbrella>/stores.yml`.

**Absent `stores.yml`** means legacy isolated-home mode. `FOLIO_HOME` is then the content root and no registry fan-out is required.

**`<store>:` references** use `<store>:<path-within-store>`. A registered prefix resolves against its canonical root; an unknown store-shaped prefix fails loudly. A path that merely contains a colon (for example `a/b:c.md`) remains a normal path.

**Write routing**: `--folio <store>:<project>` targets a project in a Folio store matched in `active/` or `archive/`. External, code, and dot stores are not Folio project write targets.


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
      - path: compiled/output.md       # Local authored output; digest freshness when recorded
      - external: jira                 # External (resolved via tooling.yml)
        id: "PROJ-456"
        field: description             # Sub-resource within external system
    composed_at: "2026-09-12T10:00:00Z"
    inputs_sha256: "<64 lowercase hex characters>"
    final: false                        # Terminal local review target

    # Batch variant (mutually exclusive with tree)
    batch:
      system: google_docs              # Default external for all items
      field: body                      # Default field for all items
      items:
        - id: "item-name"
          source: compiled/tab.md
          output: { id: "google-doc-id", field: "Tab Name" }
          inputs_sha256: "<64 lowercase hex characters>"

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

## Composition Snapshots

`folio touch <target>` records `composed_at` and the SHA-256 digest of declared local inputs
without rewriting output files. The two fields are an all-or-nothing pair. `--final` is
reserved for a non-batch, non-forest target with at least one direct external output and an
existing local review copy. Final targets stop inbound Folio propagation, while local stale or
missing output state remains observable.

## Schema 2 Changes

Schema 2 (`schema: 2`) introduces:

- **`observations:`** single list for all captured items (replaced former `pending:` + `tasks:`).
- **`plan` type** (alias: `brief`). `folio new plan <topic>` scaffolds `work/active/<date>-<topic>/README.md`. If a work dir already exists for the topic (e.g., from a prior `folio new design`), the README.md is added to it instead of creating a new dir.
- **`design` type** scaffolds inside a work directory: `folio new design <topic>` creates `work/active/<date>-<topic>/reference/design/<date>-<topic>.md`, auto-creating the work dir if it doesn't exist. Designs always live in work directories — not at project-level `reference/design/`.
- **`sketch` type** scaffolds inside a work directory: `folio new sketch <capability>` creates `work/active/<date>-<topic>/reference/sketch/index.html` (a fixed-name HTML page), auto-creating the work dir. A colocatable lifecycle type like `design`, sitting between spike and design in the lifecycle.
- **Reference labels**: research, insight, guide, domain, review. Mapped from old names (survey→research, synthesis→research, pattern→insight).
- **`how:` optional**: Missing `how` produces a warning (data-declaration target), not an error.

`folio init` generates schema 2 by default. All projects now use `observations:` directly.
