# dendrik — blueprint

dendrik itself, as a system — both the owner of these concepts and a composer of them (its own
dogfood). Concept ids resolve in `building-blocks.md`.

- composes: result-envelope, exit-code, contract, core-shell, agentskills, flag-registry
- because: dendrik is built on the same primitives it defines — it eats its own dog food, so this
  blueprint doubles as the record of the patterns already extracted into `pkg/dendrik`.
- non-goals: nothing composed that it doesn't own; this is the reference instance, not a
  propagation case (folio/jf are the propagation cases).
- status: stable

## Composed concepts
| concept | role in dendrik | conforms? |
|---|---|---|
| result-envelope | lint/build CLI output | ✓ json-output, no-json-encoder, no-raw-json, run-has-json |
| exit-code | typed exit status | ✓ exit-constants |
| contract | the lint-check registry it owns | ✓ (owns it) |
| core-shell | dendrik is itself a wired skill⋈CLI | ✓ dendrik-import, go-work-sync, symlink-entries, core-in-pkg, main-dispatch |
| agentskills | validates its own skill | ✓ skill-frontmatter, skill-links, skill-size |
| flag-registry | its CLI flag vocabulary | ✓ (owns it) |

Fully conformant — `dendrik lint cmd/dendrik` is 32/32. The clean reference example.
