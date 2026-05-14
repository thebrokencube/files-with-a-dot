---
name: dendrik
description: "Use when reviewing any SKILL.md, CLAUDE.md, reference file, or agentic documentation for quality — structure, description, progressive disclosure, writing style. Also use when validating tool conventions, checking contract compliance, or debugging lint failures in dotfiles CLI tools. Triggers: 'review my skill', 'check this skill', 'review this doc', 'is this SKILL.md good', 'skill quality', 'dendrik', 'lint', 'convention check'."
user_invocable: true
argument-hint: "[review [path] | lint <path> [--json] [--strict] | --explain <id>]"
---

# dendrik

Convention contract linter and quality reviewer for skills and agentic documentation.

## Triage

When invoked, determine which lens to use:

| Signal | Lens |
|---|---|
| `dendrik lint <path>` or explicit lint request | **Conventions** — full contract lint |
| `dendrik review <path>` or path to a SKILL.md | **Skill Review** — quality rubric |
| Path to CLAUDE.md, reference file, or doc | **Doc Review** — doc quality rubric |
| `dendrik --explain <id>` | **Explain** — show check rationale |
| No arguments, no clear context | Detect context (see below) |

### Context Detection (bare invocation)

When `/dendrik` is called with no arguments, infer what to review:

1. **File just edited or created** in this conversation → review that file, infer lens
2. **Active skill or doc work** visible in conversation → review the artifact in progress
3. **Multiple candidates** → list them and ask which to review
4. **Nothing detected** → ask: "What should I review?"

Never guess ambiguously. If unsure, ask.

---

## Lens: Conventions

Full contract lint for dotfiles CLI tools in `cmd/*/`. Delegates to the `dendrik` CLI.

```bash
cd ~/.dotfiles
dendrik lint <path>            # all checks
dendrik lint <path> --strict   # promote warnings to errors
dendrik lint <path> --json     # structured output
dendrik lint --explain <id>    # rationale for a check
```

Three layers validated: Go (6 checks), Skill (9 checks), Bridge (10 checks).

-> Read references/contract-checks.md for full check details with remediation examples.

-> Read references/cli-conventions.md for CLI conventions (exit codes, flags, output).

---

## Lens: Skill Review

Quality review for any SKILL.md and its surrounding directory.

1. Read the target SKILL.md in full
2. Read the surrounding directory (references/, scripts/) for context
3. Read `references/review-rubric.md` — evaluate against 5 areas:
   - Structure & Progressive Disclosure
   - Description Quality
   - Single Responsibility
   - Writing Quality
   - Completeness
4. Score each area as pass/warn/fail
5. Present top 5 findings ranked by impact
6. End with 1-2 "do these first" action items

-> Read references/review-rubric.md for scoring criteria and examples.

-> Read references/skill-conventions.md for convention details.

---

## Lens: Doc Review

Quality review for CLAUDE.md files, reference docs, and agentic documentation.

1. Read the target document in full
2. Read `references/review-rubric.md` — evaluate against doc review areas:
   - Structure
   - Clarity
   - Progressive Disclosure
   - Link Integrity
   - Actionability
3. Score each area as pass/warn/fail
4. Present top 5 findings ranked by impact

-> Read references/review-rubric.md (Doc Review Areas section) for criteria.

---

## Output Format

All lenses produce findings in the same format:

```
**[Area] — [pass/warn/fail] [brief verdict]**
> Specific observation. Concrete suggestion.
```

Maximum 5 findings. If everything passes, say so — don't manufacture problems.

## Reference Library

| Reference | When to read |
|---|---|
| `references/review-rubric.md` | Skill review or doc review lens |
| `references/skill-conventions.md` | Skill review — convention details |
| `references/cli-conventions.md` | Conventions lens — CLI details |
| `references/contract-checks.md` | Conventions lens — all 25 check IDs |
