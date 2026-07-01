# jf — blueprint

Which dendrik concepts jf composes, and why. Concept ids resolve in `building-blocks.md` (the
generated map shipped alongside this file).

- composes: result-envelope, exit-code, core-shell
- because: jf wraps its own `NodeResult`/`StatusResult` inside the result-envelope's `Data` rather
  than defining a second output contract — one JSON shape across the fleet.
- non-goals: no jf-local exit scheme or output encoder; jf defers to dendrik.
- status: evolving

## Composed concepts
| concept | role in jf | conforms? |
|---|---|---|
| result-envelope | CLI output; `NodeResult`/`StatusResult` ride in `.Data` | ⚠ partial — `no-raw-json` (2), `run-has-json` (6): some commands emit raw JSON or lack `--json` |
| exit-code | typed exit status on commands | ✓ exit-constants |
| core-shell | wired skill⋈CLI (go.work use, symlink_map, importable core) | ✓ dendrik-import, go-work-sync, symlink-entries |

`[NEEDS CLARIFICATION]` result-envelope: are the `run-has-json` / `no-raw-json` gaps intentional
(commands that legitimately don't emit JSON) or debt to close? (Separately, jf's skill carries 7
`work-specific-content` lint errors — a skill-quality issue, not a composition concern.)
