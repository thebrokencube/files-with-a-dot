# folio — blueprint

Which dendrik concepts folio composes, and why. Concept ids resolve in dendrik's
`building-blocks.md` (the generated map, shipped with the dendrik plugin).

- composes: result-envelope, exit-code, core-shell
- because: folio is a Go consumer — it imports dendrik's output + exit primitives and is wired as a
  core-shell, rather than carrying its own output/exit conventions.
- non-goals: no folio-local JSON envelope or exit scheme.
- status: evolving

## Composed concepts
| concept | role in folio | conforms? |
|---|---|---|
| result-envelope | CLI output | ⚠ partial — `no-json-encoder` (1), `run-has-json` (6): raw JSON / missing `--json` |
| exit-code | typed exit status | ⚠ partial — `exit-constants` (4): bare exit ints instead of the named constants |
| core-shell | wired skill⋈CLI (go.work use, symlink_map, importable core) | ✓ dendrik-import, go-work-sync, symlink-entries |

`[NEEDS CLARIFICATION]` result-envelope + exit-code gaps: real debt to close, or intentional?
(Separately, folio's skill carries a `skill-size` warning — skill quality, not a composition concern.)
