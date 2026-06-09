# Review Rubric

Scoring criteria for reviewing skills and agentic documentation. Read by the
dendrik skill during skill review and doc review lenses.

## How to Use

1. Read the target file in full. For skills, also read the surrounding directory
   (references/, scripts/).
2. Evaluate each area below. Score as **pass**, **warn**, or **fail**.
3. Surface the top 5 findings ranked by impact.
4. Format each finding using the output template at the bottom.
5. End with 1-2 concrete "do these first" action items.
6. If everything is solid, say so. Do not manufacture findings.

---

## Skill Review Areas

### 1. Structure & Progressive Disclosure

**Check:**
- Is the SKILL.md body under 500 lines? Under ~3000 words?
- Is domain knowledge in `references/`, not inline?
- Are reference files linked with arrow syntax or explicit paths?
- Does the skill load only what it needs, when it needs it?
- Is inline content actionable (commands, sequences, checks) not explanatory?
- Are structured content blocks (queries, templates, constants, schemas) extracted
  into individual files? (See **Inline Structured Content** sub-check below.)

**Pass:** Lean SKILL.md with clear pointers to reference files. Knowledge is separated
from workflow. The agent reads references on demand.

**Warn:** Some inline knowledge that could be extracted but doesn't severely bloat the file.
Body is functional but over ~200 lines.

**Fail:** Embedded knowledge walls. Body over 500 lines. References exist but aren't
linked from the main file.

**Example — pass** (folio):
> ~1700 words main SKILL.md + 12 reference files. Workflows are concise routing tables.
> Deep procedure lives in references loaded via arrow syntax.

**Example — fail:**
> 400-line SKILL.md that explains Kafka consumer groups, entity relationships, and
> configuration details inline alongside startup commands. All loaded on every invocation.

#### Inline Structured Content

Structured content blocks — SQL queries, output templates, document templates, constant
tables, JSON/YAML schemas, multi-line code examples — belong in individual files, not
inline. Each block should be a standalone file that the main document links to.

**What to extract:**

| Content type | Signal | Target file pattern |
|---|---|---|
| SQL query (>10 lines) | `SELECT`, `INSERT`, `WITH` blocks | One file per query, named by what it answers |
| Output / document template | Heredoc, HTML, placeholder-heavy blocks | One file per template |
| Constants / magic numbers | Repeated IDs, enum values, config tables | Single shared constants file |
| Code example (>15 lines) | Reference material, not illustrative snippet | One file per example |
| JSON/YAML schema | Full schema definitions | One file per schema |

**Extraction principles:**
- Name files by purpose, not by sequence (`eligibility-population.md`, not `query-1.md`)
- Co-locate with the workflow that consumes them (e.g., queries near the walkthrough
  that references them)
- Link from the parent document — the reader should discover extracted files through
  navigation, not by browsing directories
- Include a template file when the directory will grow (scaffolds consistency for new additions)
- Keep brief illustrative snippets (≤5 lines) inline — extraction is for reference
  material that has independent utility

**Scoring:**
- **Pass:** No inline structured content blocks, or only brief illustrative snippets (≤5 lines).
- **Warn:** 1-2 inline blocks of 10-30 lines. Functional but would benefit from extraction.
- **Fail:** 3+ inline blocks, or any single block >30 lines. Burns context on every load
  and prevents independent reuse.

**Example — pass** (required_retirement_plan):
> Queries in individual files under `queries/`, each following a template with metadata,
> SQL with placeholders, placeholder docs, and notes. Constants in a shared reference file.
> Parent workflow docs link to queries by step number.

**Example — fail:**
> CLAUDE.md with a 50-line SQL query, a 30-line output template, and a 20-line constants
> table all inline. Every agent load pays the full token cost; queries can't be reused
> or updated independently.

---

### 2. Description Quality

**Check:**
- Does the description lead with what the skill does?
- Does it include specific trigger phrases users would naturally say?
- Does it include domain-specific terms that appear in user questions?
- Is the description over-generic ("helps with X") instead of specific?
- Formula: `[What it does]. [Trigger phrases/keywords]. [Domain concepts].`

**Pass:** Description functions as a routing table — specific triggers and domain terms
that drive accurate activation.

**Warn:** Functional but missing key trigger phrases. Activation works for common cases
but misses edge cases.

**Fail:** Vague summary like "helps with backend development" or "manages infrastructure."
Agent cannot reliably route to this skill.

**Example — pass** (jf):
> "Use when managing Jira tickets, creating/editing issues, pushing descriptions,
> searching with JQL, managing ticket lifecycle (parking, repurposing), or when needing
> Jira conventions, project field defaults, and content gotchas."

**Example — fail:**
> "Use when working with tickets."

---

### 3. Single Responsibility

**Check:**
- Can the skill's job be described in 8 words or fewer?
- Is it mixing workflow instructions with domain knowledge?
- Does it combine multiple unrelated capabilities?
- Does the description require "and" to explain what it does?

**Pass:** One clear job. The 8-word test passes cleanly.

**Warn:** One primary job with a minor secondary concern that could be split but isn't
urgent.

**Fail:** Multiple distinct jobs that should be separate skills. Fails the 8-word test.

**Example — pass** (commit): "Format and create VCS commits" (6 words).

**Example — fail:** A skill that does startup AND troubleshooting AND monitoring.
Requires "and" twice to describe. Split into 3 skills.

---

### 4. Writing Quality

**Check:**
- Does the body use imperative/infinitive form (verb-first instructions)?
- Is it free of second-person ("you should", "you need to")?
- Are instructions objective and instructional ("To accomplish X, do Y")?
- Is content concise — no filler, no restated context?
- Are tables and lists used for structured information instead of prose?

**Pass:** Clean imperative style. Tables for structured data. No second-person.

**Warn:** Occasional second-person or passive voice but content is clear.

**Fail:** Predominantly second-person. Explanations that should be instructions.
Prose walls where tables would serve better.

**Example — pass:**
> "Run `folio validate` to check structural integrity. If validation fails,
> read the error output and fix the referenced file."

**Example — fail:**
> "You should run the validate command to make sure everything is good. If you
> see errors, you'll need to look at what went wrong and fix it."

---

### 5. Completeness

**Check:**
- Do all referenced files exist? (arrow refs, markdown links, script paths)
- Does the frontmatter have required fields? (`name`, `description`)
- If `user_invocable: true`, is `argument-hint` present?
- Are there dead references to files that don't exist?
- For CLI-backed skills: does the skill mention how to invoke the CLI?

**Pass:** All links resolve. Frontmatter is complete. No dead references.

**Warn:** Minor gaps — e.g., a reference file mentioned but not yet created.

**Fail:** Broken links. Missing required frontmatter. Structural issues that
`dendrik lint` would catch.

---

## Doc Review Areas

For CLAUDE.md files, reference docs, and other agentic documentation:

### 1. Structure

- Is content organized with clear headings?
- Is information scannable (tables, lists) not buried in prose?
- Is the document appropriately sized for its role?

### 2. Clarity

- Are instructions actionable and unambiguous?
- Could an agent follow these instructions without additional context?
- Are terms defined or linkable?

### 3. Progressive Disclosure

- Does the doc front-load the most important information?
- Are details deferred to linked references where appropriate?
- Is the reader's time respected — no unnecessary preamble?
- Are structured content blocks (queries, templates, constants, schemas) extracted
  into individual files? Apply the same **Inline Structured Content** sub-check
  from Skill Review Area 1 — the thresholds, content types, and scoring are identical.

### 4. Link Integrity

- Do all file paths and references resolve?
- Are cross-references accurate and up to date?

### 5. Actionability

- Does the doc drive behavior (not just inform)?
- Are rules concrete enough to follow consistently?
- Could two different agents reading this produce similar behavior?

---

## Harness Portability (all lenses)

Evaluate on every skill and doc review. Most strongly relevant when the artifact targets
multiple harnesses (Claude Code, Cursor, Codex, …), declares itself harness-agnostic, or
labels parts of itself "harness-specific" / "Claude-specific". For a single-harness artifact
that makes no portability claim, score **pass** and move on.

**Check:**

- **Over-claimed harness-specificity.** Is anything labeled harness-specific actually portable?
  Orchestration — dispatching a subagent, isolating a workspace (it's `git worktree`
  underneath), running linters, file operations, cleanup — are portable *capabilities*. Only
  the **binding** (the concrete tool/command that performs the step — a specific Agent/Task
  tool, a `/slash` command, an editor-specific agent) is harness-specific. Flag portable logic
  siloed in a harness-specific file or section.
- **Recipe vs adapter separation.** The portable procedure (the "recipe") should live once in
  the agnostic layer (shared docs, `AGENTS.md`) as the source of truth, with a thin per-harness
  "adapter" holding only the bindings. Flag the same workflow duplicated across the agnostic and
  harness layers (drift risk), or a whole procedure stranded in one harness's file when only a
  step or two is genuinely bound.
- **Capability vs binding language.** Steps should read as capabilities ("dispatch a reviewer",
  "isolate the PR into a workspace") with the concrete tool deferred to the adapter — not one
  harness's tool name hardcoded as if it were the only way.
- **Placement.** No harness-isms (specific tool names, `/slash` commands, dotdir paths) in
  agnostic docs; no portable steps stranded in a single harness's overlay.

**Pass:** Portable steps live once in the agnostic layer as capabilities; only true bindings
(dispatch mechanism, isolation mechanism, tool/command names) sit in thin per-harness adapters.
Or: single-harness artifact making no portability claim.

**Warn:** Mostly separated, but a few portable steps are described in harness-specific terms,
or there is minor duplication between the recipe and an adapter.

**Fail:** A portable procedure is wholesale labeled "harness-specific"; the same workflow is
duplicated across the agnostic and harness layers; or agnostic docs hardcode one harness's
tools.

**Example — fail:**
> A "review a PR" workflow (isolate a worktree → run linters → dispatch a reviewer subagent →
> report → cleanup) placed entirely in a `Claude-specific` reference file. Only the reviewer
> *dispatch* and the *isolation* mechanism are harness-bound; the rest is portable and belongs
> in the agnostic recipe, with a thin adapter naming just the two bindings.

---

## Output Format

Present each finding as:

```
**[Area] — [pass/warn/fail] [brief verdict]**
> Specific observation. Concrete suggestion.
```

Example:

```
**Description — fail: too generic for auto-activation**
> "Help with local development" gives no trigger phrases. Add specific terms:
> 'start servers', 'run specs', 'migrate database'.
```

## Presentation Rules

- Maximum 5 findings, ranked by impact.
- Lead with the highest-impact issue.
- End with "do these 1-2 things first" — specific, actionable next steps.
- If the skill is strong, say so. Do not manufacture findings.
- Use pass/warn/fail vocabulary consistently.
