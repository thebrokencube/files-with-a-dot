# Review Leaf — Skill (SKILL.md, skill reference files, slash-commands)

Type-local dimensions for a skill. Applied with `review-shared.md`. Routing:
- **SKILL.md** → all areas below.
- **Skill reference file** (`references/*.md`) → Structure & Progressive Disclosure (esp. the
  Inline Structured Content sub-check) + Completeness; skip Description/Single-Responsibility
  (those are SKILL.md-level).
- **Slash-command** (`.claude/commands/*.md`) → same as SKILL.md, plus: check for drift vs a
  sibling skill of the same name (commands and skills are the same mechanism — they should not
  silently diverge).

## 1. Structure & Progressive Disclosure

**Check:** SKILL.md body under 500 lines / ~3000 words? Domain knowledge in `references/`, not
inline? Reference files linked with arrow syntax or explicit paths? Loads only what it needs,
when it needs it? Inline content actionable (commands, sequences, checks) not explanatory?

**Pass:** Lean SKILL.md with clear pointers to references; knowledge separated from workflow.
**Warn:** Some extractable inline knowledge; body over ~200 lines.
**Fail:** Embedded knowledge walls; body over 500 lines; references exist but aren't linked.

### Inline Structured Content (sub-check)
Structured blocks — SQL queries, output/document templates, constant tables, JSON/YAML schemas,
multi-line code examples — belong in individual files, not inline.

| Content type | Signal | Target |
|---|---|---|
| SQL query (>10 lines) | `SELECT`/`WITH` blocks | one file per query, named by what it answers |
| Output / document template | heredoc, placeholder-heavy | one file per template |
| Constants / magic numbers | repeated IDs, enum tables | single shared constants file |
| Code example (>15 lines) | reference material | one file per example |
| JSON/YAML schema | full schema | one file per schema |

Keep brief illustrative snippets (≤5 lines) inline. **Pass:** no inline blocks, or only ≤5-line
snippets. **Warn:** 1-2 blocks of 10-30 lines. **Fail:** 3+ blocks, or any single block >30
lines.

## 2. Description Quality

**Check:** Leads with what the skill does? Includes specific trigger phrases and domain terms?
Not over-generic ("helps with X")? Formula: `[What it does]. [Trigger phrases]. [Domain concepts].`

**Pass:** Functions as a routing table — specific triggers and domain terms drive accurate
activation. **Warn:** functional but missing key triggers. **Fail:** vague summary the agent
can't route on ("helps with backend development").

## 3. Single Responsibility

**Check:** Job describable in 8 words or fewer? Not mixing workflow with domain knowledge? Not
combining unrelated capabilities? Description doesn't need "and" to explain itself?

**Pass:** one clear job, 8-word test passes. **Warn:** one primary job + a minor splittable
secondary concern. **Fail:** multiple distinct jobs; fails the 8-word test.

## 4. Completeness

**Check:** All referenced files exist (arrow refs, markdown links, script paths)? Required
frontmatter present (`name`, `description`)? If `user_invocable: true`, is `argument-hint`
present? Any dead references? For CLI-backed skills: does it mention how to invoke the CLI?

**Pass:** all links resolve, frontmatter complete, no dead references. **Warn:** minor gap (a
referenced file not yet created). **Fail:** broken links, missing required frontmatter — the
structural issues `dendrik lint` would catch.
