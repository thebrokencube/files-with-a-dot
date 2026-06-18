# CLI Conventions

dendrik is the platform — a library of composable primitives for agentic tooling
— today expressed as a convention contract (`dendrik lint`) and a type-dispatching
review framework (`/dendrik`). See `cmd/dendrik/docs/00-what-is-dendrik.md`.

Conventions codified from the jf and folio implementations. This is the source
of truth for the CLI conventions enforced by the `dendrik lint` contract (see
`contract.go`) and `dendrik new` (Track 5).

Philosophy: Sinatra, not Rails. dendrik provides functions at the points where
convention matters. It doesn't own `main()`, dispatch, config schemas, help
text, or subcommand routing.

**Human-agent parity**: Every operation accessible to humans is accessible to
agents, and vice versa. Non-TTY stdout defaults to JSON. Exit codes have
defined agent actions. SKILL.md is the agent discovery surface; `--help` is the
human discovery surface. Neither should duplicate the other.

**Transport**: the CLI (text + `--json`) is the canonical agent interface. An
MCP surface is a deferred, additive option — a tool-specific adapter over the
same command functions — not a current target.

---

## Exit Codes

| Code | Constant | Meaning | Agent action |
|------|----------|---------|-------------|
| 0 | `ExitOK` | Success | Proceed |
| 1 | `ExitUserError` | Bad input, missing args, invalid flags | Fix invocation and retry |
| 2 | `ExitExternalErr` | API down, file not found, network failure | Report to user or retry later |
| 3 | `ExitConflict` | Resource conflict (e.g., sync collision) | Resolve conflict, then retry |

**Decision rules:**

- Parse errors (bad flags, missing required args) → `ExitUserError`
- File not found, YAML parse failure, validation failure → `ExitExternalErr`
- External tool failure (jf binary, API call) → `ExitExternalErr`
- Sync conflict where both sides changed → `ExitConflict`
- "Stale targets found" or "lint issues found" (exit-as-signal) → `ExitUserError`

dendrik CLIs never call `os.Exit()` directly. Commands return an int; the
dispatcher in `main()` calls `os.Exit()` once.

**Divergence from convention (intentional).** GNU/POSIX shells commonly reserve
`2` for *usage* errors; dendrik instead uses `2` for external/environment errors
and folds usage errors into `1` (`ExitUserError`). There is no universal
exit-code standard — BSD `sysexits.h` and `grep` each differ again — so dendrik
defines a small, semantically-actionable house scheme rather than conforming to a
norm that doesn't exist. The per-code **agent actions** above are the contract;
the numbers are stable and are not renumbered.

---

## Flag Conventions

### Global reservations

These short flags have fixed meaning across all dendrik CLIs:

| Flag | Long | Meaning |
|------|------|---------|
| `-h` | `--help` | Show help (provided by ff) |
| `-j` | `--json` | Machine-readable JSON output |
| `-n` | `--dry-run` | Preview without side effects |

### Common long-only flags

| Flag | Meaning | Notes |
|------|---------|-------|
| `--no-color` | Disable colored output | Respects `NO_COLOR` env var per spec |
| `--version` | Print version | Handled in `main()` dispatch; `-V` short form |

### Per-CLI flags

Short flags outside the global set are CLI-specific. The same letter can mean
different things in different CLIs. This is acceptable because CLIs are
independent binaries — users never mix flags across them.

Known collisions documented in `flag_registry.go`:

| Flag | jf | folio |
|------|-----|-------|
| `-f` | `--force` (push/sync) | `--folio` (path to folio.yml) |
| `-d` | `--dir` (working directory) | — (not used) |
| `-s` | `--subtree` (push) | `--source` / `--status` / `--scope` (varies) |
| `-m` | — | `--materialize` / `--message` (varies) |

### Flag parsing

- Use `dendrik.NewFlagSet()` + `dendrik.Parse()` — wraps ff v4
- ff v4 handles interspersed flags natively (flags work in any position)
- Short flags: `fs.String('d', "dir", ".", "description")`
- Long-only flags: `fs.StringLong("no-color", "description")`
- Parse errors return `error` — never `os.Exit()`

### Env var fallback

ff supports `ff.WithEnvVarPrefix("JF")` for automatic env→flag binding.
Convention: prefix matches the CLI name in uppercase. Currently:

| CLI | Prefix | Status |
|-----|--------|--------|
| jf | `JF_` | Planned, not yet wired |
| folio | `FOLIO_` | Planned, not yet wired |

`FOLIO_HOME` is a standalone env var (not a flag fallback) used to locate
the folio home directory.

---

## Output Conventions

### Output mode

`dendrik.OutputMode(jsonFlag, plainFlag)` returns one of:

| Mode | When | Format |
|------|------|--------|
| `"json"` | `--json` flag or non-TTY stdout | Structured JSON via `dendrik.WriteResult()` |
| `"plain"` | `--plain` flag | Undecorated text (no ANSI, no JSON) |
| `"human"` | TTY stdout, no flags | Colored, formatted for humans |

Non-TTY defaults to JSON so agents get structured output automatically.

### JSON envelope

All JSON output uses `dendrik.ResultEnvelope`:

```go
type ResultEnvelope struct {
    Data     any    `json:"data"`
    Error    string `json:"error,omitempty"`
    Detail   string `json:"detail,omitempty"`
    ExitCode *int   `json:"exit_code,omitempty"`
}
```

**Success** — data present, no error:

```json
{"data": {"tool": "jf", "errors": 0, "warnings": 3, "results": [...]}}
```

**Error** — error present, no data:

```json
{"error": "forest.yml not found", "detail": "run jf init to create one"}
```

**Exit-as-signal** — data present with non-zero exit (e.g., `folio stale`):

```json
{"data": {"stale": ["target-1"]}, "exit_code": 1}
```

`exit_code` is only present when non-zero. Agents should check `error` first,
then `exit_code`, then treat presence of `data` as success.

### Error output

- Errors go to stderr, data goes to stdout
- Human errors use `output.Errf()` (colored "Error:" prefix)
- Structured errors use `dendrik.WriteError()` (JSON envelope)

### Color

- `dendrik.ColorEnabled(noColorFlag)` checks: flag → `NO_COLOR` env → TTY
- `dendrik.Palette` provides ANSI codes with zero-value fallback for no-color
- Commands that render colored output accept `--no-color` (long-only)

---

## Command Structure

### What dendrik owns

- Flag parsing (`NewFlagSet`, `Parse`)
- Exit codes (`ExitOK`, `ExitUserError`, `ExitExternalErr`, `ExitConflict`)
- Output formatting (`WriteResult`, `WriteError`, `OutputMode`)
- Terminal detection (`IsTerminal`, `ColorEnabled`)
- Color palette (`Palette`, `NewPalette`)

### What authors own

- Command dispatch (`switch` on args — too simple to abstract)
- Config schemas (folio.yml, forest.yml are domain-specific)
- Subcommand routing
- Help text
- Domain data types
- External tool wrappers

### FC;IS pattern

Commands are pure functions returning data structures. The dispatch layer
handles output mode:

```go
func statusCommand(fs *ff.FlagSet) (StatusData, error) {
    // Pure — no I/O
    return StatusData{...}, nil
}

func runStatus(args []string) int {
    // Imperative shell — handles I/O at the edge
    fs := dendrik.NewFlagSet("status")
    jsonFlag := fs.Bool('j', "json", "JSON output")
    if err := dendrik.Parse(fs, args); err != nil {
        fmt.Fprintln(os.Stderr, err)
        return dendrik.ExitUserError
    }
    data, err := statusCommand(fs)
    if err != nil { return dendrik.ExitExternalErr }
    switch dendrik.OutputMode(*jsonFlag, false) {
    case "json":  dendrik.WriteResult(os.Stdout, data)
    case "human": renderHumanStatus(data)
    }
    return dendrik.ExitOK
}
```

This separation means commands are testable without capturing stdout, and
TUI (future) renders the same data structures.
