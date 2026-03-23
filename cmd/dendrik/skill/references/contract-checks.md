# Contract Checks Reference

All 25 checks in the dendrik tool contract. Use `dendrik lint --explain <id>` for rationale and remediation.

## Go Layer (6 checks)

### go-mod-linked (Error)

go.mod exists and go.work links this tool to pkg/dendrik.

**Fix**: Create go.mod with `go mod init` and add `./cmd/<tool>` to the `use` block in go.work.

### main-dispatch (Error)

main.go has `func main()` with at least one `os.Exit(run*(...))` call.

**Fix**: Delegate to `run*()` functions via `os.Exit(run*(...))` in main.go.

### cmd-file-exists (Error)

At least one `cmd_*.go` file exists.

**Fix**: Create a file matching `cmd_*.go` (e.g., `cmd_lint.go`) with a `run*()` function.

### makefile-targets (Error)

Makefile exists with `build`, `test`, `check` targets.

**Fix**: Create a Makefile with these three targets. See `cmd/jf/Makefile` for reference.

### readme-exists (Error)

README.md exists in the tool directory.

**Fix**: Create `README.md` in `cmd/*/`.

### readme-sections (Warning)

README.md contains `## Install`, `## Quick Start`, `## Commands`, `## Code Structure`.

**Fix**: Add the missing `##` sections. These are checked via exact string match.

## Skill Layer (9 checks)

### skill-exists (Error)

SKILL.md exists at `skill/SKILL.md`.

**Fix**: Create `skill/SKILL.md` with YAML frontmatter containing `name` and `description`.

### skill-frontmatter (Error)

Valid YAML frontmatter with `name` (1-64 chars, lowercase+hyphens) and `description` (1-1024 chars).

**Fix**: Add valid `name:` and `description:` fields.

### skill-extra-fields (Warning)

No unexpected frontmatter fields outside the Agent Skills spec.

Known spec fields: `name`, `description`, `version`, `compatibility`, `metadata`, `user_invocable`, `argument-hint`.

### skill-links (Error)

Standard markdown `[text](path)` links in SKILL.md resolve to existing files.

**Fix**: Ensure all linked files exist at the specified paths.

### ref-naming (Warning)

Reference files in `references/` follow kebab-case naming (lowercase, hyphens between words).

### skill-size (Error)

SKILL.md body does not exceed 500 lines. Token estimate warning at ~5000 tokens.

**Fix**: Move detailed content to reference files and link with arrow syntax.

### argument-hint (Error)

If `user_invocable: true`, then `argument-hint` is present.

**Fix**: Add `argument-hint: "<command> [flags]"` to frontmatter.

### arrow-refs (Error)

All arrow references (`->` Read references/...) resolve to existing files.

**Fix**: Create the referenced file in `skill/references/` or fix the path.

### activation-guidance (Warning)

Description includes routing hints like "Use when...", "For tasks that...", or "Use this...".

**Fix**: Add activation phrasing to help agents decide when to invoke the skill.

### activation-metadata (Error)

If `trigger`, `skip_when`, or `related` fields are present, they must be non-empty strings or string arrays.

**Fix**: Remove empty fields or provide valid values.

## Bridge Layer (10 checks)

### dendrik-import (Error)

At least one `.go` file imports `pkg/dendrik`.

**Fix**: Import `github.com/thebrokencube/files-with-a-dot/pkg/dendrik`.

### exit-constants (Error)

No bare integer returns (0, 1) in `cmd_*.go`. No `os.Exit()` outside `main.go`.

**Fix**: Use `dendrik.ExitOK`, `dendrik.ExitUserError`, etc. Move `os.Exit` to main.go only.

### json-output (Error)

If `--json` flag exists, at least one code path uses `dendrik.WriteResult` or `Output.Result`.

**Fix**: Add structured output calls in commands that register `--json`.

### go-work-sync (Error)

go.work `use` entries match `cmd/*/` directories with go.mod (symmetric difference).

**Fix**: Update go.work to match the filesystem. Remove stale entries, add missing ones.

### symlink-entries (Error)

`symlink_map.txt` has entries for the binary and skill directory.

**Fix**: Add `cmd/<tool>/<tool>:$HOME/.local/bin/<tool>` and `cmd/<tool>/skill:$HOME/.claude/skills/<tool>`.

### makefile-gofiles (Warning)

Makefile GOFILES find path includes `../../pkg/dendrik`.

**Fix**: Update to `$(shell find . ../../pkg/dendrik -name '*.go')`.

### no-json-encoder (Error)

No `json.NewEncoder` in `cmd_*.go` files. Bypasses the ResultEnvelope contract.

**Fix**: Use `dendrik.Output.Result()` or `dendrik.WriteResult()` instead.

### no-raw-json (Warning)

No raw `fmt.Print(string(` JSON passthrough in files with `--json` flag. Lines using `MustResult` or `.Result(` are allowed.

**Fix**: Wrap in ResultEnvelope or add `//nolint:no-raw-json` if passthrough is intentional.

### run-has-json (Warning)

All `cmd_*.go` files with `run*` functions register a `--json` flag.

**Fix**: Add `--json` to the flag set, or acknowledge the gap is intentional.
