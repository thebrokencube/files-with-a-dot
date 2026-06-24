# Contract Checks Reference

The enumerated catalog of the dendrik tool contract. This is the one derived source for check IDs and remediation; it is kept honest against the canonical `Contract` slice in `pkg/dendrik/conventions/contract.go` by a count-assert test. Use `dendrik lint --explain <id>` for rationale and remediation.

## Go Layer

### go-mod-linked (Error)

go.mod exists and go.work links this tool to pkg/dendrik.

**Fix**: Create go.mod with `go mod init` and add `./cmd/<tool>` to the `use` block in go.work.

### main-dispatch (Error)

main.go has `func main()` with at least one `os.Exit(run*(...))` call.

**Fix**: Delegate to `run*()` functions via `os.Exit(run*(...))` in main.go.

### core-in-pkg (Error)

Verb domain types live in the importable `pkg/dendrik/<verb>` core, not in `package main`. A `cmd_<verb>.go` that declares a top-level type without importing `pkg/dendrik/<verb>` traps the logic in `package main`.

**Fix**: Move the type into `pkg/dendrik/<verb>` and import it; keep `cmd_<verb>.go` a thin shell (parse flags → call the core → render output).

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

### claude-md-exists (Warning)

CLAUDE.md exists in the tool directory.

**Fix**: Create CLAUDE.md with standardized skeleton: Build, Test, Binary Distribution, Code Conventions, Deep Context.

### docs-naming (Error)

All files in `docs/` match numbered kebab-case pattern (`NN-name.md`, e.g., `01-getting-started.md`).

**Fix**: Rename to match the pattern. Only fires if `docs/` directory exists.

### docs-getting-started (Warning)

`docs/01-getting-started.md` exists when `docs/` directory is present.

**Fix**: Create `docs/01-getting-started.md` as the entry point for human-progressive documentation.

### readme-doc-links (Error)

Links in README.md `## Documentation` section resolve to existing files.

**Fix**: Ensure all `[text](path)` links in the Documentation section point to existing files.

### version-flag (Warning)

main.go handles a `--version` flag (with `-V`), distinct from the `version` subcommand.

**Fix**: In `main()`'s dispatch, fold the flag forms into the version case: `case "version", "--version", "-V":` printing the version and exiting 0.

## Skill Layer

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

## Bridge Layer

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

`symlink_map.txt` has an entry for the skill directory. (Binaries are installed from GitHub Releases by `dot sync`, not symlinked — see `pkg/dendrik/conventions/release.md`.)

**Fix**: Add `cmd/<tool>/skill:$HOME/.claude/skills/<tool>`.

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
