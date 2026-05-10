# Lifecycle Derivation

Suggests next actions for a project based on `type` and `status` fields on folio.yml source entries (schema v3). Used by the bare `/folio` invocation after the user picks a project.

## Derivation Rules

Check in order. Present the first matching suggestion per active work track.

1. Active work track with spike(s) but no design → "Spike may be ready to promote to design, or archive if investigation concluded"
2. Design with `status: active` → "Design is iterating — continue or settle"
3. Design with `status: done`, no plan → "Design is settled — create execution brief"
4. Plan with `status: active` → find first track without `status: done` → "Execute Track N"
5. All tracks `status: done`, no retro in the work track → "Write retro, then archive"
6. All done + retro exists → "Archive work track (`folio archive`)"

Phrase suggestions conversationally based on context, not robotically from the list above.

## Stale Detection

Flag active work tracks whose directory mtime is older than 14 days as "stale — consider resuming or archiving."

To obtain mtime, run `stat -f '%m' <path>` (macOS) on the work track directory itself — one Bash call per active work track, not per source file. Compare against current epoch. If no active work tracks exist, skip stale detection entirely.

14 days is a starting threshold. The suggestion says "consider" — it's guidance, not enforcement.

## Schema Migration Hint

If the project's `schema` is less than 3 and `reference/spike/` or `reference/retro/` directories exist on disk, mention: "This project uses schema v2 — lifecycle artifacts in `reference/` can be migrated to `work/` (see `references/migrate.md`)."

## Fallback

If no `type`/`status` fields are present (schema 2 projects without migration), fall back to the existing behavior: suggest actions based on observations, staleness, and compose/publish readiness.
