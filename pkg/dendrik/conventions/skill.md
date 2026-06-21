# Skill Conventions

dendrik is the shared foundation the dotfiles CLI tools (folio, jf, dot) are built
on; this file holds its skill conventions. See `cmd/dendrik/docs/00-what-is-dendrik.md`.

Conventions codified from the jf and folio skill implementations. This is the
source of truth for the skill conventions enforced by the `dendrik lint` contract
(see `contract.go`) and `dendrik new` (Track 5).

**Agent Skills standard alignment**: dendrik's skill validation is grounded
in the Agent Skills standard (agentskills.io). Layer 1 checks (`skill-exists`
through `skill-size`) validate any SKILL.md against the standard. Layer 2
checks (`argument-hint` through `activation-metadata`) add dendrik-specific
conventions. The `pkg/dendrik/agentskills/` package implements Layer 1 as a
standalone validator.

---

## SKILL.md Structure

### Frontmatter (required)

```yaml
---
name: toolname
description: "Verb-rich description with trigger phrases for discovery"
---
```

Required fields: `name`, `description`.

Optional fields:
- `user_invocable: true` — skill can be invoked with `/toolname`
- `argument-hint: "[command] [args]"` — required if `user_invocable: true`
- `version` — skill version string
- `compatibility` — compatibility constraints
- `metadata` — arbitrary key-value metadata object
- `trigger` — string or string array; custom routing condition for when to activate
- `skip_when` — string or string array; custom routing condition for when NOT to activate
- `related` — string or string array; cross-references to related skills

The `name` field must be present even for plugin-distributed skills (confirmed
in dendrik architecture hardening — contradicts the original replicant doc
which proposed removing it).

Conditional activation fields (`trigger`, `skip_when`, `related`) are optional.
If present, they must be non-empty strings or string arrays (check ID:
`activation-metadata`). These fields are not in the core Agent Skills spec but
are supported by the dendrik validator as extension fields.

### Description guidelines

The description is how Claude Code discovers skills. Include:

- **Action verbs** that match user intent (managing, creating, editing, pushing)
- **Tool/file names** the skill operates on (Jira, folio.yml, tickets)
- **Workflow names** (plan, compose, review, lifecycle)

Examples from production:

> jf: "Use when managing Jira tickets, creating/editing issues, pushing
> descriptions, searching with JQL, managing ticket lifecycle (parking,
> repurposing), or when needing Jira conventions, project field defaults,
> and content gotchas."

> folio: "Use when planning non-trivial tasks, composing outputs, or managing
> knowledge work projects. Lifecycle toolkit with folio.yml-driven
> source-to-target composition and diverge-converge planning."

### Word budget

| Section | Budget | Notes |
|---------|--------|-------|
| Main SKILL.md | 600–3000 words | Core commands, workflows, quick reference |
| Each reference file | Unbounded | Deep detail on one topic |

The main SKILL.md should be self-sufficient for common operations. Reference
files provide depth for complex workflows. Current production sizes:

- jf: ~666 words main + 7 reference files
- folio: ~1,704 words main + 12 reference files

---

## Directory Layout

### CLI-backed skill (jf, folio)

```
cmd/{cli}/skill/
  SKILL.md                    # Main skill file
  references/                 # Progressive depth
    getting-started.md
    configuration.md
    conventions.md
    ...
  tooling.yml                 # Optional: external system routing
```

### Plugin-compatible layout

For dual-homing (symlink + plugin install), skills use the plural `skills/`
directory with a CLI subdirectory:

```
cmd/{cli}/skills/{cli}/
  SKILL.md
  references/
```

Note: jf and folio currently use `cmd/{cli}/skill/` (singular). Migration to
`skills/` (plural) is part of Track 8 (plugin packaging).

How this layout gets packaged into a plugin and published across harnesses: see
`distribution.md`.

### Standalone skill (no CLI)

```
skills/{name}/
  SKILL.md
  references/                 # Optional
```

Examples: commit, stacked-pr, dotfiles, nvim — simpler skills without
a backing CLI binary or reference files.

---

## Progressive Disclosure

### Arrow syntax

Link from the main SKILL.md to reference files using arrow syntax:

```markdown
-> See references/quick-push.md for details.
-> Read references/gather.md for full workflow
```

Arrows must point to existing files. `dendrik lint` validates these links
(check ID: `arrow-refs`).

### When to use references

- **In main SKILL.md**: Commands, flags, quick workflows, decision trees
- **In references/**: Multi-step procedures, configuration guides, gotchas,
  patterns, examples

Rule of thumb: if a section exceeds ~300 words and serves a single workflow,
extract it to a reference file.

### Reference file naming

Use descriptive kebab-case names that match the topic:

- `quick-push.md` — not `ref-1.md`
- `forest-management.md` — not `trees.md`
- `jql-patterns.md` — not `search-help.md`

---

## External Tooling Integration

### tooling.yml

Optional file that declares which external systems a skill needs. Currently
only used by folio:

```yaml
external_systems:
  - name: jira
    mcp_server: jira-confluence
    capabilities: [read_issues, edit_issues, search_jql]
  - name: gdrive
    mcp_server: gdrive
    capabilities: [search, read, create, update]
```

This is advisory — it helps the skill discover MCP tools at runtime but
doesn't enforce anything.

---

## Patterns Across Skills

### CLI + skill hybrid model

```
+----------------------------+     +----------------------------+
|       Skill Layer          |     |        CLI Layer           |
|                            |     |                            |
| SKILL.md + references/     |     | Binary on PATH             |
| Creative work:             |     | Deterministic work:        |
|   plan, compose, review    |     |   validate, push, status   |
|                            |     |                            |
| Agent interface            |     | Execution engine           |
+----------------------------+     +----------------------------+
        |                                    ^
        | Bash tool calls (often --json)     |
        +------------------------------------+
```

The skill layer handles creative/judgmental work (planning, composition,
review). The CLI layer handles deterministic operations (validation, file
manipulation, API calls). Skills call CLIs via Bash tool, typically with
`--json` for structured output.

### What skills should NOT do

- Call `os.Exit()` or assume terminal state
- Embed CLI flag details that could drift from the actual binary
- Duplicate information available via `{cli} --help`
- Include code examples longer than 5 lines (put them in references/)
- Use raw `json.NewEncoder` or `fmt.Print(string(` to output JSON (use `dendrik.Output`)
- Hardcode file paths that vary per environment (use config or env vars)
- Provide structured `--help` or `--capabilities` (SKILL.md IS the agent discovery layer)
- Add `--verbose` flags (agents need full data via `--json` or minimum; verbosity is a human concept)
