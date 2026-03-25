# Reference

## Check Catalog

dendrik validates 29 checks across three layers. Each check has an ID, severity, and actionable remediation.

### Go Layer (10 checks)

Build infrastructure and documentation structure.

| ID | Severity | Description |
|----|----------|-------------|
| `go-mod-linked` | error | go.mod exists and go.work links this tool |
| `main-dispatch` | error | main.go has `func main()` with `os.Exit(run*(...))` |
| `cmd-file-exists` | error | At least one `cmd_*.go` file exists |
| `makefile-targets` | error | Makefile has `build`, `test`, `check` targets |
| `readme-exists` | error | README.md exists in tool directory |
| `readme-sections` | warning | README.md has Install, Quick Start, Commands, Code Structure sections |
| `claude-md-exists` | warning | CLAUDE.md exists in tool directory |
| `docs-naming` | error | Files in `docs/` match numbered kebab-case (`NN-name.md`) |
| `docs-getting-started` | warning | `docs/01-getting-started.md` exists when `docs/` is present |
| `readme-doc-links` | error | Links in README.md Documentation section resolve to existing files |

### Skill Layer (9 checks)

Agent discovery and skill file quality.

| ID | Severity | Description |
|----|----------|-------------|
| `skill-exists` | error | `skill/SKILL.md` exists |
| `skill-frontmatter` | error | Valid YAML frontmatter with name (1-64 chars) and description (1-1024 chars) |
| `skill-extra-fields` | warning | No unexpected frontmatter fields outside the Agent Skills spec |
| `skill-links` | error | Markdown links in SKILL.md resolve to existing files |
| `ref-naming` | warning | Reference files in `references/` follow kebab-case naming |
| `skill-size` | error | SKILL.md body does not exceed 500 lines |
| `argument-hint` | error | If `user_invocable: true`, then `argument-hint` is present |
| `arrow-refs` | error | Arrow references (`->`) in SKILL.md and references resolve to existing files |
| `activation-guidance` | warning | Description includes activation guidance ("use when", "for tasks that") |
| `activation-metadata` | error | If trigger/skip_when/related fields present, they are valid |

### Bridge Layer (10 checks)

Platform integration between tools and the shared library.

| ID | Severity | Description |
|----|----------|-------------|
| `dendrik-import` | error | At least one .go file imports `pkg/dendrik` |
| `exit-constants` | error | No bare integer returns in `cmd_*.go`; no `os.Exit()` outside main.go |
| `json-output` | error | If `--json` flag exists, code uses `dendrik.WriteResult` or `dendrik.WriteError` |
| `go-work-sync` | error | go.work `use` entries match `cmd/*/` directories with go.mod |
| `symlink-entries` | error | `symlink_map.txt` has entries for binary and skill directory |
| `makefile-gofiles` | warning | Makefile GOFILES find path includes `../../pkg/dendrik` |
| `no-json-encoder` | error | No `json.NewEncoder` in `cmd_*.go` files |
| `no-raw-json` | warning | No `fmt.Print(string(` in `cmd_*.go` files with `--json` flag |
| `run-has-json` | warning | All `cmd_*.go` run functions register a `--json` flag |

## Severity Model

| Severity | Meaning | Gate |
|----------|---------|------|
| **error** | Contract violation -- must fix before tagging | Blocks release |
| **warning** | Convention gap -- should fix | Advisory only |

`--strict` promotes all warnings to errors, collapsing the distinction. Use this for pre-release validation.

Design rationale: errors protect downstream consumers (broken builds, missing agent entry points). Warnings protect future maintainers (missing docs, convention gaps).

## Output Formats

### Human Output (default)

```
  E [go-mod-linked] go.work does not link this tool (go.work)
    Add `./cmd/jf` to the `use` block in go.work.
  W [readme-sections] README.md missing section: ## Code Structure (README.md)
    Add a `## Code Structure` section to README.md.

jf: 1 error(s), 1 warning(s)
```

### JSON Output (--json)

```json
{
  "ok": true,
  "data": {
    "tool": "jf",
    "errors": 1,
    "warnings": 1,
    "results": [
      {
        "check_id": "go-mod-linked",
        "severity": "error",
        "message": "go.work does not link this tool",
        "file": "go.work",
        "remediation": "Add `./cmd/jf` to the `use` block in go.work."
      }
    ]
  }
}
```

JSON output wraps results in a `ResultEnvelope` (`ok` + `data`), consistent with all dendrik platform tools.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All checks passed (or only warnings in non-strict mode) |
| 1 | Errors found |
| 2 | External error (e.g., can't find go.work, can't read tool directory) |

## Flags

| Flag | Description |
|------|-------------|
| `--json` | Structured JSON output (ResultEnvelope format) |
| `--strict` | Promote warnings to errors |
| `--no-color` | Disable colored output |
| `--explain <id>` | Show rationale and remediation for a specific check |

## The Contract Philosophy

Why 29 checks and not 5 or 100?

The contract is the minimum set of conventions that keep independently-developed tools composable:

- **Go layer** ensures shared build infrastructure works: every tool builds the same way, has the same entry point pattern, and provides standard documentation.
- **Skill layer** ensures agent discoverability: every tool can be found, invoked, and understood by Claude Code through well-formed SKILL.md files.
- **Bridge layer** ensures platform integration: every tool uses the shared library correctly, produces structured output, and registers itself in the dotfiles ecosystem.

Tools are free to diverge on everything else -- command structure, internal architecture, feature scope. The contract covers only what matters for interoperability.
