# Find

Search across folio knowledge for a topic. Vault-first — cross-cutting references are checked before project-scoped content.

## Search Order

1. **Vault** (`~/.folio/vault/`) — filenames and content across all labels (research, domain, guide, insight)
2. **Current project** — if a folio.yml is in scope, search its `reference/` and `work/` trees
3. **All active projects** (`~/.folio/active/`) — broaden to every project's references and work dirs

## How to Search

- Grep for the query (case-insensitive) in filenames first, then file contents
- For each tier, report matches before moving to the next tier
- Stop expanding scope when results are sufficient (user has what they need)

## Output Format

```
Vault matches:
  - vault/research/2026-03-22-background-agent-landscape.md — [first line or title]
  - vault/domain/agent-ecosystem-map.md — [first line or title]

Project matches (current: files-with-a-dot):
  - reference/spike/2026-03-01-migration-written-culture.md — [first line or title]

Cross-project matches:
  - ret/guideline-written-culture/reference/research/2026-02-01-openai-harness.md — [title]
```

Present results as a compact list with file paths and titles. Offer to read any match in detail.
