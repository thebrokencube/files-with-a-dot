# Skill Conventions

Canonical source: `pkg/dendrik/conventions/skill.md`

Covers SKILL.md structure (frontmatter, description guidelines, word budget), directory
layout, progressive disclosure (arrow syntax, reference files), and the CLI+skill hybrid
model. See the canonical source for full details, layout examples, and tooling.yml spec.

## Quick Reference (from canonical source)

- **Required frontmatter**: `name` (lowercase+hyphens, 1-64 chars), `description` (1-1024 chars)
- **If user_invocable**: must include `argument-hint`
- **Word budget**: 600-3000 words in main SKILL.md, unbounded in references
- **Arrow syntax**: arrow references point to `references/*.md` for progressive disclosure
- **Reference naming**: kebab-case (e.g., `quick-push.md`, not `ref-1.md`)
- **Hybrid model**: Skill handles creative work, CLI handles deterministic work

## Description as Routing Table

The description is the primary mechanism for skill discovery and activation. Treat it as
a routing specification, not a summary.

**Formula**: `[What it does]. [Trigger phrases/keywords]. [Domain concepts].`

- Lead with what the skill does — this appears in autocomplete
- Include trigger phrases users would naturally say (not formal commands)
- Include domain-specific terms that appear in user questions
- For troubleshooting skills: include actual error strings
- Avoid generic triggers like "help" or "fix" — too many skills match

**Strong**: "Use when managing Jira tickets, creating/editing issues, pushing descriptions,
searching with JQL, managing ticket lifecycle (parking, repurposing), or when needing Jira
conventions, project field defaults, and content gotchas."

**Weak**: "Use when working with Jira."

## Writing Style

Write skill bodies using **imperative/infinitive form** (verb-first instructions), not
second person. Use objective, instructional language.

| Do | Don't |
|---|---|
| "Run `folio validate` to check integrity." | "You should run validate to check." |
| "Parse the frontmatter using YAML." | "You can parse the frontmatter." |
| "To accomplish X, do Y." | "If you need to do X, you should do Y." |

Tables and lists for structured information. Prose only when flow matters.

## Single Responsibility (8-Word Test)

Describe the skill's job in 8 words or fewer. If you need "and," it's doing too much.

Split signals:
- Multiple unrelated `## Step` sections doing different things
- Description requiring "and" to explain what it does
- Skill handling both happy path and error recovery for different workflows
- Body exceeds 300 lines despite good progressive disclosure

Prefer composing focused skills over building monolithic ones. An orchestrator skill can
call focused skills in sequence.

## Anti-Patterns Quick Reference

| Anti-pattern | Failure mode | Fix |
|---|---|---|
| Monolithic (>200 lines) | Triggers too broadly, hard to maintain | Decompose by responsibility |
| Knowledge inline | Duplicated, drifts from reality | Reference library pattern |
| Vague description | Misroutes or doesn't activate | Add trigger phrases and domain terms |
| Recipes without knowledge | Can't handle novel problems | Add reference docs that explain HOW, not just WHAT |
| Complex bash inline | Untestable, fragile | Extract to `scripts/` |
| Prose walls | Burns context, obscures workflow | Tables and lists for structured info |
