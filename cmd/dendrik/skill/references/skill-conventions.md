# Skill Conventions

Canonical source: `pkg/dendrik/conventions/skill.md`

Covers SKILL.md structure (frontmatter, description guidelines, word budget), directory layout, progressive disclosure (arrow syntax, reference files), and the CLI+skill hybrid model.

Key points for quick reference:

- **Required frontmatter**: `name` (lowercase+hyphens, 1-64 chars), `description` (1-1024 chars)
- **If user_invocable**: must include `argument-hint`
- **Word budget**: 600-3000 words in main SKILL.md, unbounded in references
- **Arrow syntax**: arrow references point to `references/*.md` for progressive disclosure
- **Reference naming**: kebab-case (e.g., `quick-push.md`, not `ref-1.md`)
- **Hybrid model**: Skill handles creative work, CLI handles deterministic work

See the canonical source for full details, layout examples, and tooling.yml spec.
