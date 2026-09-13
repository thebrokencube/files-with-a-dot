# Lifecycle Derivation

Suggests next actions for a project based on `type` and `status` fields on folio.yml source entries (schema v3). Used by the bare `/folio` invocation after the user picks a project.

## Derivation Rules

Check in order. Present the first matching suggestion per active work track.

1. Active work track with spike(s) but no sketch or design → "Spike may be ready to lay out as a sketch, or archive if investigation concluded"
   When the user brings work items at this stage, route to: `/folio observe` or `/folio gather`
2. Frozen sketch but no design → "Sketch is frozen — harden into a design (N≥2 tracks) or go straight to implementation (lightweight, N==1)"
   When the user brings work items at this stage, route to: `/folio plan` or `track-1.md` content
3. Design with `status: active` → "Design is iterating — continue or settle"
   When the user brings work items at this stage, route to: design doc content or `/folio compose`
4. Design with `status: done`, no plan → "Design is settled — create execution brief"
   When the user brings work items at this stage, route to: execution brief content or `/folio plan`
5. Plan with `status: active` → find first track without `status: done` → "Execute Track N"
   When the user brings work items at this stage, route to: `/folio observe` or track status update
6. All tracks `status: done`, no retro in the work track → "Write retro, then archive"
   When the user brings work items at this stage, route to: `folio new retro` or observation
7. All done + retro exists → "Archive work track (`folio archive`)"
   When the user brings work items at this stage, route to: `/folio wrap-up`

Phrase suggestions conversationally based on context, not robotically from the list above.

## Freshness and Work-Track Age

Folio does not infer work-track age from filesystem timestamps. Target freshness is
content-based: a complete matching composition digest is `clean`; changed inputs are
`stale`; missing outputs are `missing`; and an existing output without a complete
snapshot is `unknown`. Unknown is an honest signal, not a failure.

Use authored artifacts and the manifest's declaration order for lifecycle guidance.
Terminal targets marked `final: true` are omitted from the stale queue only while their
local outputs are clean. If a final target becomes stale or missing, report it; do not
accept a final flag as proof that the output still matches its inputs.

## Schema Migration Hint

If the project's `schema` is less than 3 and `reference/spike/` or `reference/retro/` directories exist on disk, mention: "This project uses schema v2 — lifecycle artifacts in `reference/` can be migrated to `work/` (see `references/migrate.md`)."

## Session Entry Display

Used by the bare `/folio` invocation to present a deterministic active-work list
instead of an alphabetical dump. This replaces the default project list — it is
not a second list on top.

### Implementation

1. Run `folio home list` to get active project paths and the selected content work root.
2. Read each project's manifest and collect active work tracks referenced by local
   source declarations. Preserve the manifest declaration order; append unreferenced
   active tracks in lexical path order.
3. Derive each track's lifecycle stage from authored artifact presence: spike, sketch,
   design, plan, implementation, or retro. Do not infer recency from directory metadata.
4. Show each project and its observation count; when no active track is declared,
   show the project without inventing recency.

### Display Format

```
Active work:
  1. SRM — "Legacy Launch Prep" (design; authored design doc)
  2. Folio — "Session Handoff" (sketch; authored sketch)
  3. dot — 3 observations (no active track)
Pick up a project or command.
```

- Cap displayed entries at 5; overflow as "N more — name them to see"
- Preserve manifest declaration order, with lexical order only for unreferenced tracks
- Show authored artifact evidence instead of age labels
- Do not group entries into recent/stale buckets

## Fallback

If no `type`/`status` fields are present (schema 2 projects without migration), derive
the suggestion from authored observations and available lifecycle artifacts. If the
manifest cannot establish a touched project or track, ask the user rather than infer it
from filesystem timestamps.

