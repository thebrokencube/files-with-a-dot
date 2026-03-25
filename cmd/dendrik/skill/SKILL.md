---
name: dendrik
description: "Use when validating tool conventions, checking contract compliance, or debugging lint failures in dotfiles CLI tools. Runs a 29-check contract across Go, Skill, and Bridge layers."
user_invocable: true
argument-hint: "lint <path> [--json] [--strict] [--explain <id>]"
---

# dendrik

Tool contract linter for the dotfiles platform. Validates that CLI tools in `cmd/*/` follow the 29-check dendrik convention contract across three layers.

## Quick Reference

| Command | What it does |
|---------|-------------|
| `dendrik lint <path>` | Run all 29 checks against a tool directory |
| `dendrik lint <path> --json` | Structured JSON output |
| `dendrik lint <path> --strict` | Promote warnings to errors |
| `dendrik lint --explain <id>` | Show rationale and remediation for a check |

## Layers

| Layer | Checks | What it validates |
|-------|--------|-------------------|
| Go | 6 | go.mod, go.work, main dispatch, Makefile targets, README |
| Skill | 9 | SKILL.md frontmatter, links, arrow refs, activation guidance, size |
| Bridge | 10 | dendrik imports, exit constants, JSON output, symlink entries, go.work sync |

## Check IDs

### Go Layer
- `go-mod-linked` — go.mod exists and go.work links this tool
- `main-dispatch` — main.go delegates to run*() via os.Exit
- `cmd-file-exists` — at least one cmd_*.go file exists
- `makefile-targets` — Makefile has build, test, check targets
- `readme-exists` — README.md exists in tool directory
- `readme-sections` — README has Install, Quick Start, Commands, Code Structure sections

### Skill Layer
- `skill-exists` — SKILL.md exists at skill/SKILL.md
- `skill-frontmatter` — valid name and description in frontmatter
- `skill-extra-fields` — no unexpected frontmatter fields
- `skill-links` — markdown links resolve to existing files
- `ref-naming` — reference files follow kebab-case naming
- `skill-size` — SKILL.md body under 500 lines
- `argument-hint` — user_invocable: true requires argument-hint
- `arrow-refs` — arrow references resolve to existing files
- `activation-guidance` — description includes routing hints

### Bridge Layer
- `dendrik-import` — at least one file imports pkg/dendrik
- `exit-constants` — no bare integer returns in cmd_*.go
- `json-output` — --json flag produces structured output
- `go-work-sync` — go.work entries match cmd/*/ directories
- `symlink-entries` — symlink_map.txt has binary and skill entries
- `makefile-gofiles` — Makefile find path includes pkg/dendrik
- `no-json-encoder` — no json.NewEncoder in cmd_*.go
- `no-raw-json` — no raw JSON passthrough with --json flag
- `run-has-json` — run* functions register --json flag
- `activation-metadata` — trigger/skip_when/related fields are valid

-> Read references/contract-checks.md for full check details with remediation examples.

## Interpreting Results

Severity levels:
- **Error** — contract violation, must fix before tagging
- **Warning** — convention gap, should fix but non-blocking

Use `dendrik lint --explain <check-id>` to see the rationale and remediation for any check.

## Typical Workflow

```bash
# After making changes to a tool
cd ~/.dotfiles
dendrik lint cmd/jf           # check compliance
dendrik lint cmd/jf --strict  # strict mode for pre-tag validation

# When adding a new tool
dendrik lint cmd/newtool      # see what's missing
dendrik lint --explain symlink-entries  # understand a specific check
```

## Conventions

-> Read references/cli-conventions.md for CLI conventions (exit codes, flags, output, command structure).

-> Read references/skill-conventions.md for skill conventions (frontmatter, layout, progressive disclosure).
