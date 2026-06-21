# Find

Search across folio knowledge for a topic. Vault-first — cross-cutting references are checked before project-scoped content. **`find` is the v1 cross-store discovery mechanism** — it fans out across every registered store, not just the home.

## Stores: fan out across the whole registry

Before searching, enumerate the registered stores:

```
folio stores list --json   # → [{name, path, kind}, ...]
```

`~/.folio` is the home of all folios + KBs; `stores.yml` registers others (a work folio, external ADR/RADR repos). When no `stores.yml` exists, this returns just the implicit `vault` store and `find` behaves exactly as it always has (back-compat).

Search each store according to its `kind`, and **label every hit with its store name**:
- **`kind: folio`** — structure-aware: vault-first, then `active/` projects' `reference/` and `work/` trees (the tiers below), within that store's path.
- **`kind: external`** — content grep over the store path (it's just markdown — ADRs, RADRs, wiki notes). No folio structure is assumed; read-only.

Don't double-scan the home's own `vault/` as a peer — it's covered by the vault tier.

## Search Order (within each folio store)

1. **Vault** (`<store>/vault/`) — filenames and content across all labels (research, domain, guide, insight)
2. **Current project** — if a folio.yml is in scope, search its `reference/` and `work/` trees
3. **All active projects** (`<store>/active/`) — broaden to every project's references and work dirs

## How to Search

- Grep for the query (case-insensitive) in filenames first, then file contents
- Iterate stores in registry order; report matches per store before moving on
- For each tier, report matches before moving to the next tier
- Stop expanding scope when results are sufficient (user has what they need)
- Link a cross-store result with the `<store>:<path>` reference form (e.g. `radr:0042-foo.md`)

## Output Format

```
vault matches:
  - vault:research/2026-03-22-background-agent-landscape.md — [first line or title]

work (folio) matches:
  - work:reference/spike/2026-03-01-migration-written-culture.md — [title]

radr (external) matches:
  - radr:0042-adopt-radrs.md — [first line or title]
```

Present results as a compact list with file paths and titles. Offer to read any match in detail.
