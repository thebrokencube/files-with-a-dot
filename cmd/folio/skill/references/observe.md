# Observe

Manage observations in folio.yml. Observations are an **open-items queue** — things that need attention (bugs, gaps, ideas, debt, tasks). Resolve them when addressed. Do not add observations to record completed work; that's what retros and commit history are for.

**Do NOT edit the `observations:` list in folio.yml by hand** — not to add, remove, reorder, or reformat entries. All mutations go through CLI commands, which enforce format validation. `folio home push` runs lint as a gate and will reject malformed observations.

## Workflow

1. `folio observe types` — get valid types and descriptions
2. `folio observe list --json` — check existing scopes (avoid near-duplicates)
3. `folio observe 'type(scope): description'` — append (validates format)
4. `folio observe resolve <#N|substring> [#N2 ...]` — delete resolved items. **Always batch multiple resolves in a single call** to avoid index shifting between calls.
5. `folio observe lint` — check format + inline path refs

If no text provided with the command, ask the user for the observation.

## Type Disambiguation

When the observation text has investigation depth — multiple sentences, describes a problem space, mentions alternatives or approaches — read `references/alignment.md` and run the alignment protocol before routing:

- Budget: 2
- Grounding: the observation text, existing observations (`folio observe list --json`)
- Target: routing decision (observation vs spike)
- Hard constraints: none

If the alignment routes to spike, derive the topic from the alignment's claim/recommendation (the problem space identified during questioning) and use the alignment's full output as the spike's initial content via `folio new spike <topic>`. One-liner observations pass through to `folio observe` untouched — no alignment needed.
