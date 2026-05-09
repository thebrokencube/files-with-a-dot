# Compose Workflow

Read by `/folio compose [target]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

> Schema reference: see references/schema.md for the full folio.yml structure.

## Compose Steps

1. Run `folio validate`. Stop if invalid.
2. Run `folio status --json`. Identify stale/missing/unknown targets (or filter to the specific target requested).
3. Read `blocked_by` edges from folio.yml (not in status JSON). Resolve composition order: upstream targets first.
4. For each target, in DAG order:
   a. Read source files from target's `sources`
   b. Read `how` from folio.yml
   c. Compose per the target's `how` field. **For targets with external outputs** (Jira, Slack,
      Google Docs): if `how` does not specify audience or tone, read `references/alignment.md` and
      run the alignment protocol with:
      - Budget: 4
      - Grounding: target's source files, `how` field
      - Target: ephemeral how-amendment (session-scoped annotation to `how`, not persisted)
      - Hard constraints: any explicit framing already in `how`
      For multi-target DAGs: alignment fires once at the batch start (first external target
      missing audience/tone), not per-target. Local-only targets (`path:` outputs) skip
      the alignment entirely.
   d. **Local outputs** (`path:`): write compiled file
   e. **External outputs** (`external:`): resolve push method from tooling.yml
5. **Review gate (soft)**: Present targets composed (cap at 5), output paths, and file sizes. "Review outputs? (y to review, n to continue)" — if yes, show first 10 lines of each output.
6. Run `folio status` again to report final state.

**Code references**: Use `repositories` URL patterns from folio.yml for clickable links in targets that support them.

## Forest Target Composition

Forest targets are jf-managed Jira hierarchies. Each node composes independently from its own file. Nodes do NOT consume child outputs.

1. Run `jf status --json` in the forest root to get per-node staleness
2. Read `how_default` and `how_overrides` from the target's `forest:` block in folio.yml
3. For each stale/missing node:
   a. Read the node's markdown file
   b. Apply the node's `how_overrides` entry (fall back to `how_default` if none)
   c. Use `jf push <KEY> <FILE>` to compile and push (see references/publish.md)
4. Touch the target's local `path:` output to update mtime (if one exists)

## Batch Target Composition

Multiple items sharing one `how` directive. Each item has its own source and output.

1. Get per-item status from `folio status --json` (`batch_items` field)
2. For each stale/missing item:
   a. Read item's `source` file
   b. Apply target-level `how`
   c. Push via tooling.yml. Output: `{external: batch.system, id: item.output.id, field: item.output.field || batch.field}`
3. Touch the target's local `path:` output (if one exists)

## Iteration

Composition is rarely one-shot. The compose-review-re-compose loop handles two distinct iteration types:

- **New source** (changes the DAG): gather additional source material, then re-compose. Staleness tracking handles this automatically — new/updated sources make targets stale.
- **Reframe** (same DAG, different lens): update the target's `how` field, then re-compose with `--force`. `how` isn't tracked for staleness, so reframes require an explicit force flag.

After composing, review the output — run `folio validate` and `folio status` to check for issues. If the output needs work, determine which type of iteration applies and loop back.

## Error Handling

- **Target fails mid-DAG**: Continue composing independent targets. Report partial results — which targets succeeded, which failed, and why.
