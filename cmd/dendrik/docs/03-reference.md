# Reference

## Check Catalog

dendrik validates a contract across three layers, each guarding a different concern:

- **Go layer** — build infrastructure: go.mod/go.work wiring, main dispatch, Makefile targets, README/docs structure.
- **Skill layer** — agent discovery: SKILL.md existence, frontmatter, links, arrow refs, activation guidance.
- **Bridge layer** — integration: shared-library imports, exit constants, structured JSON output, symlink and go.work sync.

The enumerated catalog — every check ID, severity, and remediation — lives in one derived reference, kept honest against the canonical `Contract` slice in `pkg/dendrik/conventions/contract.go`:

-> See [`plugins/dendrik/skills/dendrik/references/contract-checks.md`](../../../plugins/dendrik/skills/dendrik/references/contract-checks.md) for the full check catalog. Use `dendrik lint --explain <id>` for any individual check's rationale.

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

Why this contract and not a much smaller or much larger one?

The contract is the minimum set of conventions that keep independently-developed tools composable:

- **Go layer** ensures shared build infrastructure works: every tool builds the same way, has the same entry point pattern, and provides standard documentation.
- **Skill layer** ensures agent discoverability: every tool can be found, invoked, and understood by Claude Code through well-formed SKILL.md files.
- **Bridge layer** ensures platform integration: every tool uses the shared library correctly, produces structured output, and registers itself in the dotfiles ecosystem.

Tools are free to diverge on everything else -- command structure, internal architecture, feature scope. The contract covers only what matters for interoperability.
