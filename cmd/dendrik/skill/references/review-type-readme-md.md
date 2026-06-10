# Review Leaf — README.md

Type-local dimensions for a README.md. Applied with `review-shared.md`. Audience is humans-first;
agents read it as fallback context.

## Dimensions

### Human-first
README.md is for humans (intro, quick-start, orientation). Judge it as a human document — clear,
welcoming, scannable.

### Agent detail extracted to AGENTS.md
Agent-specific instructions (build/test commands for agents, conventions, constraints) clutter a
README. **warn** when agent detail lives here — that's what AGENTS.md is for; move it and keep
README focused on humans.

### Quick-start works
The quick-start / install steps are accurate and runnable as written. **fail** on stale or
broken setup instructions.

### Relationship to CLI G5/G6 (note, not overlap)
For a dendrik CLI tool, the Conventions lens already enforces README *existence* (`readme-exists`,
G5) and required `## sections` (`readme-sections`, G6) deterministically. This leaf is the
**quality** counterpart — it judges whether those sections are good, not whether they exist.
Complementary neighbors; no overlap.
