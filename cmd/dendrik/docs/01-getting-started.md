# Getting Started with dendrik

dendrik is the platform — a library of composable primitives for agentic tooling — today expressed as a convention contract (`dendrik lint`) and a type-dispatching review framework (`/dendrik`). See [00-what-is-dendrik.md](00-what-is-dendrik.md). This guide covers the contract surface (`dendrik lint`).

## What dendrik Does

dendrik enforces structural conventions across CLI tools in the dotfiles repo. It validates checks across three layers (Go, Skill, Bridge) so that tools stay consistent as they evolve independently.

dendrik is a linter for tool structure, not code style. It checks that your tool has the right files, follows naming conventions, wires up the shared library correctly, and provides skill documentation discoverable by Claude Code (the AI coding agent).

## Installing

dendrik is already on PATH if dotfiles are installed (`dot sync`). Verify:

```bash
dendrik version
```

If missing, check that `symlink_map.txt` has the `cmd/dendrik/dendrik -> ~/.local/bin/dendrik` entry and re-run `dot sync`.

## Your First Lint Run

Point dendrik at any tool directory:

```bash
dendrik lint cmd/jf
```

Sample output:

```
  E [go-mod-linked] go.work does not link this tool (go.work)
    Add `./cmd/jf` to the `use` block in go.work.
  W [claude-md-exists] CLAUDE.md not found (CLAUDE.md)
    Create CLAUDE.md with standardized skeleton: Build, Test, Binary Distribution, Code Conventions, Deep Context.

jf: 1 error(s), 1 warning(s)
```

Each line shows:
- **Severity icon**: `E` (error, red) or `W` (warning, yellow)
- **Check ID** in brackets: the unique identifier for this check
- **Message**: what's wrong, with the file location in parentheses
- **Remediation** (indented): what to do about it

## Reading Results

**Errors** are contract violations. They indicate something that will break builds, agent discovery, or platform integration. Fix these before tagging a release.

**Warnings** are convention gaps. They indicate missing documentation or best practices that should be addressed but don't block functionality.

Use `--strict` to promote all warnings to errors -- useful for pre-release validation:

```bash
dendrik lint cmd/jf --strict
```

## The --explain Flag

Get detailed rationale for any check:

```bash
dendrik lint --explain go-mod-linked
```

Output:

```
go-mod-linked [error] go.mod exists and go.work links this tool to pkg/dendrik

Rationale: Every dendrik tool is a Go module that imports the shared library.
Without go.mod the tool can't build; without the go.work link it can't resolve
pkg/dendrik locally.

Remediation: Create go.mod with `go mod init` and add a `use` entry for this
tool in the root go.work file.
```

## Fixing Violations

1. Start with errors (they block tagging)
2. Read the remediation hint -- it tells you exactly what to do
3. Make the change
4. Re-run `dendrik lint` to verify
5. Move to warnings

For structured output (useful in scripts or agent workflows):

```bash
dendrik lint cmd/jf --json
```

## What's Next

- [03-reference.md](03-reference.md) -- full check catalog by layer, output formats, exit codes
- [README.md](../README.md) -- code structure and contract layer overview
