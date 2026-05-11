# Burndown

Read by plan-execute.md when the work plan specifies `shape: burndown`. Replaces the
per-track step sequence (Steps 0-7 within Phase 7) with a wave-based batch loop.
Validation and review gates from plan-execute.md still apply per-group.

## When to Use

A track is burndown-shaped when it has homogeneous items with measurable completion
criteria, and each iteration improves context for the next. Examples: migrating N files,
fixing N lint violations, converting N models. Declare `shape: burndown` in the track
header of the execution brief.

## Concepts

| Term | Definition |
|------|-----------|
| **Wave** | A batch of groups executed with the same frozen approach. Strategy evolves between waves, not during. |
| **Group** | Atomic unit of work. One or more coupled items that produce one commit. |
| **Ratchet** | A domain-specific shell command returning a number that must monotonically improve. Defined in the execution brief. |
| **Ratchet record** | The `waves.md` table that tracks the ratchet command's output over time. |
| **Next-task** | Within a wave: pick the next unresolved group. |
| **Advance** | Between waves: run the wave checkpoint, then start the next wave. |
| **Wave checkpoint** | Toyota Kata at the wave boundary: target → actual → gap → adjustments. |

Groups are the primitive. A wave is a set of groups. Serial execution is a wave with one group.

## Wave Ordering

Order waves by ascending complexity — simplest first to learn the pattern, hardest last
when the approach is proven. Each wave definition includes:

- **Group list** — which groups are in this wave
- **Expected ratchet delta** — how much the ratchet should improve
- **"What we expect to learn"** — Toyota Kata pre-commitment (prediction before execution)

Wave ordering is a starting hypothesis. Evaluate it at every checkpoint — see
**Rebalancing** for triggers and mechanics.

## Execution Loop

For each group in the current wave:

1. **Next-task** — read group files in `progress/groups/`, pick first group without a
   completed state that isn't deferred
2. **Execute** — implement the group's work
3. **Validate** — `folio validate` (or brief-specified validation commands)
4. **Record** — write per-group file with full Kata: target condition, actual condition,
   gap, adjustments
5. **Commit** — `folio home push`

Plan-execute.md gates carry forward: validation gate and review gate apply per-group.
Burndown replaces the *selection and iteration* pattern, not the safety checks.

**Resuming**: read `waves.md`, find the ACTIVE wave, continue from first unresolved group.

## Wave Checkpoint

Run at wave boundary. Synthesizes across per-group findings. Write to
`progress/groups/wave-NN-checkpoint.md`:

1. **Target condition** — what did we expect from this wave? (from the wave definition)
2. **Actual condition** — run the ratchet command. Report the number.
3. **Patterns** — what systemic findings emerged across groups?
4. **Adjustments** — what changes for the next wave? Re-order groups, update reference
   docs, defer items. For structural changes (splitting, merging, reordering waves),
   run the **Rebalancing** protocol.

Per-group files use the same Kata format. The wave checkpoint synthesizes across
groups — patterns, systemic findings, ordering changes.

## Rebalancing

Wave structure is a hypothesis — the initial decomposition reflects what you know at
planning time. Checkpoints are where you learn whether it still holds. Rebalancing
means splitting, merging, or reordering remaining waves based on what execution revealed.

### Triggers

Evaluate rebalancing at every wave checkpoint. These signals suggest the current
decomposition is wrong:

| Signal | Example | Likely action |
|--------|---------|---------------|
| **Size imbalance** — a remaining wave is >2x the average of completed waves | Waves 1-5 averaged 15 items; Wave 6 has 105 | Split by risk tier or dependency boundary |
| **Risk heterogeneity** — a wave mixes items with different risk profiles | Low-risk retro moves + high-risk cluster moves in one wave | Separate into distinct waves ordered by risk |
| **Dependency tiers** — items within a wave have ordering constraints on each other | Independent items must finish before dependent items can move | Split into waves that respect the dependency order |
| **Approach divergence** — different items in a wave need different execution strategies | Some items are mechanical moves, others need classification judgment | Group by strategy, execute simpler strategy first |

### Mechanics

1. **Assess** — at the checkpoint, compare remaining work against completed waves.
   Look for the triggers above.
2. **Propose** — draft new wave boundaries. Each new wave must have: group list,
   expected delta, "what we expect to learn", and a clear dependency ordering
   relative to other new waves.
3. **Validate the split** — check three properties:
   - **Dependency ordering**: no wave depends on artifacts that move in a later wave
   - **Atomic clusters**: items connected by `depends_on` chains stay in the same wave
   - **Size balance**: no single wave is >2x the median of all waves (including completed)
4. **Update tracking** — revise the wave table in `waves.md` with a restructure note
   explaining why. Old wave numbers stay; new waves get the next sequential numbers.

### Anti-patterns

- **Splitting for splitting's sake** — if a wave is large but homogeneous and low-risk,
  don't split it. Split when risk tiers or dependencies demand it.
- **Merging independent projects into one wave** — projects with no shared dependencies
  should be separate waves even if individually small. Keeps blast radius contained and
  checkpoints meaningful.

## Advance & Re-scope

**Advance** has two parts:

1. **Mechanical gate** — ratchet improved (or held steady with deferrals), all groups
   resolved (done or deferred), `folio validate` passes
2. **Skill checkpoint** — the Kata questions above. Update `waves.md` with results.
   Commit via `folio home push`.

**When the ratchet can't improve**, classify:

| Classification | Signal | Action |
|---------------|--------|--------|
| Execution problem | Mistakes made, fixable | Retry within wave, apply learnings |
| Scope discovery | New items found, count increased | Update ratchet baseline, document why, continue |
| Scope change | Approach itself is wrong | Stop, consult user. May require plan re-run. |

**Flywheel composition**: ratchet tells you IF progress happened. The checkpoint tells
you WHY. The ordering update tells you WHAT comes next.

## Scale Guidance

Always start with a pilot wave (1-2 groups) to learn the pattern before batching.
Wave count is a judgment call at planning time — more waves means more checkpoints
and more flywheel turns, fewer waves means less overhead. Let complexity and
uncertainty guide the split, not item count.

## State Tracking

All state lives in the work track as plain files:

```
work/active/<slug>/
  progress/
    waves.md                       # wave table with status, counts, ratchet
    groups/
      wave-NN-<group-slug>.md      # per-group: full Kata (target/actual/gap/adjustments)
      wave-NN-checkpoint.md        # wave checkpoint (full Kata)
```

### Required Elements

Every burndown must define before starting:

1. **Ratchet** — shell command + direction (decreasing or increasing)
2. **Wave ordering** — initial waves with group assignments and expected deltas
3. **Pilot wave** — first wave with 1-2 groups to validate the approach
