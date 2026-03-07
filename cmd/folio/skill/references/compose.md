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
      Google Docs): if `how` does not specify audience or tone, ask the user before composing.
      Framing matters — exploratory vs definitive vs FYI produces very different output. One
      prompt at the start of composition saves multiple correction rounds.
   d. **Local outputs** (`path:`): write compiled file
   e. **External outputs** (`external:`): resolve push method from tooling.yml
5. **Review gate (soft)**: Present targets composed (cap at 5), output paths, and file sizes. "Review outputs? (y to review, n to continue)" — if yes, show first 10 lines of each output.
6. Run `folio status` again to report final state.

**Code references**: Use `repositories` URL patterns from folio.yml for clickable links in targets that support them.

## Tree Target Composition

Each node composes independently from its own file. Nodes do NOT consume child outputs. Compose bottom-up (children before parents).

1. Get per-node status from `folio status --json` (the `tree` field in target)
2. Read tree definition from folio.yml for per-node `file` and `how`
3. Walk bottom-up. For each stale/missing node:
   a. Read the node's `file`
   b. Apply node's `how` (fall back to target-level if none)
   c. If `compiled_dir`/`compiled_ext` set: use Jira Push Pipeline (see references/publish.md)
   d. Otherwise: compose and push via tooling.yml method for `tree.system`
   e. Skip push for non-system-ID nodes (descriptive slugs — report as "no external target")
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

After composing, review the output — run `/folio review local` explicitly, not as a suggestion. If the output needs work, determine which type of iteration applies and loop back.

## Error Handling

- **Target fails mid-DAG**: Continue composing independent targets. Report partial results — which targets succeeded, which failed, and why.
