# CLI Conventions

Canonical source: `pkg/dendrik/conventions/cli.md`

Covers exit codes, flag conventions (global reservations, per-CLI flags, parsing), output conventions (JSON envelope, color, error output), and the FC;IS command structure pattern.

Key points for quick reference:

- **Exit codes**: 0=OK, 1=UserError, 2=ExternalErr, 3=Conflict
- **Global flags**: `-h` (help), `-j` (json), `-n` (dry-run)
- **JSON envelope**: `{"data": ...}` on success, `{"error": "...", "detail": "..."}` on failure
- **Output mode**: `dendrik.NewOutput(jsonFlag, noColorFlag)` → auto-detects TTY
- **Command pattern**: Pure function returns data, imperative shell handles I/O at edges

See the canonical source for full details, flag collision table, and code examples.
