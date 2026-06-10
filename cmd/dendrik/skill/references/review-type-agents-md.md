# Review Leaf — AGENTS.md

Type-local dimensions for an AGENTS.md. Applied with `review-shared.md`. AGENTS.md is the
portable, cross-tool source of truth ("a README for agents") — the recipe other tools' adapters
point at.

## Dimensions

### Source of truth (the expected home)
AGENTS.md is where project conventions and commands belong. **pass:** AGENTS.md carries the
project's agent-facing conventions and any adapters point at it. **warn:** AGENTS.md is thin or
absent while a project CLAUDE.md (or `.cursorrules`) is fat — the content is in the wrong file.

### Behavior-changing content only
Include only what changes agent behavior: build/test commands, conventions, constraints,
non-standard patterns. **pass:** every section changes behavior. **warn:** architecture prose /
system overviews — ecosystem guidance discourages an "Architecture" section here; it costs
tokens without changing behavior. Move it to ARCHITECTURE.md and reference it (see Right Layer?).

### Nearest-file scoping (monorepos)
In a monorepo, nested AGENTS.md files resolve by directory proximity (nearest wins). **pass:** a
nested file scopes to its subtree and doesn't restate the root. **warn:** duplicated or
mis-scoped content across levels.

### Not duplicated in adapters
**pass:** adapters (CLAUDE.md, `.cursorrules`) are thin pointers to AGENTS.md. **warn:** the same
instructions duplicated across AGENTS.md and an adapter — drift risk; make the adapter a pointer.
