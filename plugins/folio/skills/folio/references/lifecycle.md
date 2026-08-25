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

## Stale Detection

Flag active work tracks whose directory mtime is older than 14 days as "stale — consider resuming or archiving."

To obtain mtime, run `stat -f '%m' <path>` (macOS) on the work track directory itself — one Bash call per active work track, not per source file. Compare against current epoch. If no active work tracks exist, skip stale detection entirely.

14 days is a starting threshold. The suggestion says "consider" — it's guidance, not enforcement.

## Schema Migration Hint

If the project's `schema` is less than 3 and `reference/spike/` or `reference/retro/` directories exist on disk, mention: "This project uses schema v2 — lifecycle artifacts in `reference/` can be migrated to `work/` (see `references/migrate.md`)."

## Session Entry Display

Used by the bare `/folio` invocation to present a recency-ranked project list instead of an alphabetical dump. This replaces the default project list — it is not a second list on top.

### Implementation

1. Run `folio home list` to get active project paths
2. For each project, find active work track directories: `find $FOLIO_HOME/active/<project-path>/work/active -maxdepth 1 -mindepth 1 -type d 2>/dev/null`
3. Get mtime for each work track dir: `stat -f '%m' <dir>` (macOS). Use `FOLIO_HOME` when a workspace is active, else `~/.folio/active/`.
4. Sort all work tracks across all projects by mtime descending
5. For each entry, derive the lifecycle stage by reading the project's folio.yml and applying the Derivation Rules above

### Display Format

```
Recently active:
  1. SRM — "Legacy Launch Prep" (design, 2h ago)
  2. Folio — "Session Handoff" (sketch, 1d ago)
Also active but stale:
  3. dot — 3 observations (14d+)
Pick up where I left off — or name a project/command.
```

- Cap displayed entries at 5; overflow as "N more — name them to see"
- Projects with no active work tracks: show project name + observation count only
- Group into "Recently active" (< 14 days) and "Also active but stale" (≥ 14 days)
- Recency labels: "today", "1d ago", "3d ago", "1w ago", "2w ago", etc.

## Fallback

If no `type`/`status` fields are present (schema 2 projects without migration), fall back to the existing behavior: suggest actions based on observations, staleness, and compose/publish readiness.
