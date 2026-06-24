# Starting a new tool

dendrik does not generate tools — it **provides** the bridge two ways: a worked exemplar to copy, and
`dendrik lint --fix` to mechanically wire the new tool into the repo. (A generator is deferred until the
blueprint engine exists; copying a live, daily-exercised tool never drifts from the contract the way a
frozen template would.)

## The core you build on

A dendrik tool is a thin `cmd/<tool>/` shell over the shared importable core in `pkg/dendrik`:

- `pkg/dendrik` (the root package) — exit-code constants (`dendrik.ExitOK`, …), the `ResultEnvelope`
  output contract (`dendrik.NewOutput`), and flag parsing (`dendrik.NewFlagSet`, `dendrik.ParseCheck`).
- `pkg/dendrik/conventions` — the lint `Contract` (the bridge the tool conforms to).
- Verb logic, when a tool grows it, lives in its own importable subpackage (e.g. `pkg/dendrik/lint`,
  `pkg/dendrik/build`) — the `cmd_<verb>.go` file stays a thin shell that imports it.

## Recipe

1. **Copy a live tool.** `cmd/jf` is the smallest full exemplar — copy it to `cmd/<tool>/`.
2. **Rename.** Update the module path in `cmd/<tool>/go.mod`, the binary name, and the skill `name:`.
   Strip the copied tool's domain content (commands, README, SKILL body).
3. **Wire it in.** Run `dendrik lint --fix cmd/<tool>`. This mechanically resolves the bridge checks it
   can: adds the `./cmd/<tool>` entry to `go.work`, and the skill line to `symlink_map.txt`.
4. **Fill the judgment.** `--fix` never invents content. Lint again (`dendrik lint cmd/<tool>`) and address
   what it reports — README sections, `docs/01-getting-started.md`, the skill description — by hand.

`--fix` only applies edits with a single mechanically-correct form. Content and ordering decisions
(e.g. `docs-naming`'s `NN-` prefix) are reported, never guessed.
