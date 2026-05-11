# Migrate

Reference for migrating lifecycle artifacts from `reference/` into `work/` directories. This guide was crystallized from the lifecycle-evolution burndown (199 artifacts, 17 projects, 9 waves).

## When to Use

When a folio project has lifecycle artifacts (spikes, designs, retros) living in `reference/spike/`, `reference/design/`, or `reference/retro/` and they need to move into `work/` tracks. This applies when upgrading to schema v3 or standardizing an older project.

## Artifact Tiers

| Tier | Types | Fields | Lives in |
|------|-------|--------|----------|
| 1 | design, plan, track | `type` + `status` (active/done) | `work/` |
| 2 | spike, retro | `type` only | `work/` |
| 3 | research, domain, guide, insight, review | none | `reference/` or vault |

## Pre-Flight

Run once before any moves.

### 1. Ratchet Baseline

Count lifecycle artifacts to establish the starting point:

```bash
find $FOLIO_HOME/active/<project>/reference/spike \
     $FOLIO_HOME/active/<project>/reference/design \
     $FOLIO_HOME/active/<project>/reference/retro \
     -name '*.md' -type f 2>/dev/null | wc -l
```

Target: 0.

### 2. Dependency Index

For projects with `depends_on` entries referencing lifecycle artifacts:

1. Read folio.yml — extract every source with `depends_on`
2. For each artifact in `reference/spike|design|retro/`, list all sources that depend on it
3. Identify clusters: groups of files connected by `depends_on` chains
4. Classify artifacts as **independent** (no depends_on exposure) or **dependent** (in a cluster)

### 3. Orphan Scan

Compare filesystem against folio.yml to find unregistered files. Register or remove before migrating. Orphan rate varies by project maturity (~0% for stable projects, ~10-25% for active investigation projects).

### 4. Survey Directory Check

If `reference/survey/` exists, plan a rename to `reference/research/` during wrap-up.

## Classification Framework

For each artifact, apply signals in order:

1. **`depends_on` chains** — colocate with the work track that depends on it
2. **Slug/topic match** — `reference/spike/2026-03-22-knowledge-residency.md` → `work/archive/2026-03-22-knowledge-residency/`
3. **Temporal proximity** — same date prefix, same topic → likely related
4. **Content references** — read the summary for explicit work track references
5. **Fallback** — create standalone lightweight archived work track

Binary rule: every artifact gets **Path A** (colocate with existing work track) or **Path B** (new lightweight archived work track). No artifact stays in `reference/`.

## Per-Artifact Checklist

1. Classify destination (Path A or B)
2. Create target directory if needed:
   - Spikes: `work/archive/YYYY-MM-DD-<topic>/spike/`
   - Designs: `work/archive/YYYY-MM-DD-<topic>/reference/design/`
   - Retros: `work/archive/YYYY-MM-DD-<topic>/retro/`
3. Move file to target directory
4. Update folio.yml source `path:`
5. Add `type:` field (spike, design, or retro)
6. Add `status: done` if Tier 1 (designs in archived tracks are always `done`)
7. Update all `depends_on` references pointing to old path — atomic for cluster moves
8. Scan `cross_references` section for old paths
9. Run `folio validate` — must pass
10. Grep folio.yml for old path — zero matches

## Execution Order

Migrate in ascending risk order:

1. **Zero-artifact projects** — schema bump only, validates the tooling
2. **Independent artifacts** — no depends_on exposure, simple moves
3. **Dependent artifacts** — cluster moves requiring atomic path updates

Within each tier, further order by project size (smallest first). This builds process confidence before tackling complex projects.

### Sub-grouping

Split large batches into homogeneous sub-groups. Push after each sub-group:

- Path A spikes (colocate into existing work tracks)
- Path B spikes (create standalone work tracks)
- Path A designs (colocate)
- Path B designs (standalone)
- Dependent clusters (one push per cluster, smallest clusters first)

### Cluster Moves

For artifacts connected by `depends_on` chains:

1. Move ALL files in the cluster before running validate
2. Update ALL `depends_on` references (both the moved artifact's entry and downstream consumers)
3. Grep for every old path in the cluster — zero matches across entire folio.yml
4. Push the cluster as one atomic commit

## Per-Project Wrap-up

After all artifacts moved:

1. Rename `reference/survey/` → `reference/research/` if it exists (update folio.yml paths)
2. Bump `schema: 3` in folio.yml
3. Remove empty `reference/spike/`, `reference/design/`, `reference/retro/` directories
4. Check for stray non-.md files in emptied directories (HTML renders, etc.)
5. Run `folio validate` and `folio health`
6. Push

## Invariants

At completion:

- Primary ratchet = 0 (no lifecycle artifacts in `reference/`)
- All lifecycle source entries have `type:` fields
- Tier 1 entries have `status:` fields
- `folio validate` passes for all projects
- `folio health` reports Good for all projects
- No orphaned empty `reference/spike|design|retro/` directories

## Pitfalls

- **Colocated designs are not migration targets.** Designs already inside `work/.../reference/design/` are correctly placed — don't re-move them.
- **cross_references hold old paths too.** Any folio.yml section that references paths must be scanned, not just `sources` and `depends_on`.
- **Orphans appear during moves.** Files on disk but missing from folio.yml. Register them before moving to avoid validation failures.
- **Non-.md files survive directory cleanup.** The ratchet counts `.md` only. Use `folio health` as the authoritative end-state check.
- **Agent prompts for batch moves must include push frequency.** "Push after each sub-group" is easy to omit from ad-hoc prompts. Include it explicitly.
