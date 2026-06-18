---
name: dendrik
description: "Use when reviewing any SKILL.md, CLAUDE.md, reference file, or agentic documentation for quality — structure, description, progressive disclosure, writing style. Also use when validating tool conventions, checking contract compliance, or debugging lint failures in dotfiles CLI tools. Triggers: 'review my skill', 'check this skill', 'review this doc', 'is this SKILL.md good', 'skill quality', 'dendrik', 'lint', 'convention check'."
user_invocable: true
argument-hint: "[review [path] | lint <path> [--json] [--strict] | --explain <id>]"
---

# dendrik

dendrik is the platform — a library of composable primitives for agentic tooling — today expressed as a convention contract (`dendrik lint`) and a type-dispatching review framework (`/dendrik`). See docs/00-what-is-dendrik.md.

## Triage

When invoked, determine which lens to use:

| Signal | Lens |
|---|---|
| `dendrik lint <path>` or explicit lint request | **Conventions** — full contract lint |
| `dendrik review <file\|repo\|PR>` or any reviewable doc (incl. SKILL.md) | **Review** — detect type → dispatch |
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

Three layers validated: Go (build infrastructure), Skill (agent discovery), Bridge (integration).

-> Read references/contract-checks.md for full check details with remediation examples.

-> Read references/cli-conventions.md for CLI conventions (exit codes, flags, output).

---

## Lens: Review

Quality review of agentic documents — by *type*. Each type (SKILL.md, CLAUDE.md, AGENTS.md,
ARCHITECTURE.md, README.md, skill reference files, slash-commands) has a different job,
audience, load model, and dominant failure mode, so review applies type-appropriate criteria,
not one generic rubric. SKILL.md is just one type among many.

### Detect type (first match wins)

| Priority | Signal | Leaf |
|---|---|---|
| 1 | path `references/*.md` under a skill dir | skill-reference → `review-type-skill.md` |
| 1 | path under `.claude/commands/` or a `commands/*.md` command file | slash-command → `review-type-skill.md` |
| 2 | filename `SKILL.md` | skill → `review-type-skill.md` |
| 2 | filename `CLAUDE.md` | claude-md → `review-type-claude-md.md` |
| 2 | filename `AGENTS.md` | agents-md → `review-type-agents-md.md` |
| 2 | filename `ARCHITECTURE.md` / `*-architecture.md` | architecture-md → `review-type-architecture-md.md` |
| 2 | filename `README.md` | readme-md → `review-type-readme-md.md` |
| 3 | frontmatter `name`+`description` | skill → `review-type-skill.md` |
| — | no match | generic → `review-shared.md` only |

Ambiguous? State the inferred type and ask before proceeding (see Context Detection).

### Per-artifact unit

1. Read the target in full (+ surrounding dir if a skill).
2. Detect type (above).
3. Load `references/review-shared.md` (universal dimensions + Right Layer? + scoring/output).
4. Load the one matching `references/review-type-*.md` leaf (skip if generic).
5. Apply the shared dimensions (incl. Right Layer?) + the leaf's type-local dimensions.
6. Score pass/warn/fail; present top findings + 1-2 "do these first" items.

### Targets: file | repo | PR

- **file** → run the per-artifact unit once. Done.
- **repo / dir** or **PR** → many artifacts + relationships between them.
  -> Read references/review-orchestration.md for input resolution (glob / `gh pr diff`),
  the cross-artifact composition pass (the Right Layer? opinion across files), caps, and the
  multi-artifact output skeleton.

-> Read references/skill-conventions.md for convention details (skill targets).

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
| `references/review-shared.md` | Review lens — every review (universal dims, Right Layer?, scoring, output) |
| `references/review-type-skill.md` | Review lens — detected type: skill / skill reference / slash-command |
| `references/review-type-claude-md.md` | Review lens — detected type: CLAUDE.md |
| `references/review-type-agents-md.md` | Review lens — detected type: AGENTS.md |
| `references/review-type-architecture-md.md` | Review lens — detected type: ARCHITECTURE.md |
| `references/review-type-readme-md.md` | Review lens — detected type: README.md |
| `references/review-orchestration.md` | Review lens — repo/PR target (multi-artifact) |
| `references/skill-conventions.md` | Review lens — skill convention details |
| `references/cli-conventions.md` | Conventions lens — CLI details |
| `references/contract-checks.md` | Conventions lens — all check IDs |
