# Plan Workflow — Idea/Architecture Phase (Phase 1.5)

Read by `/folio plan` after Phase 1 (Understand), before Phase 2 (Propose). Produces the
**idea/architecture sketch**: a single self-contained HTML page you react to *before* committing
to a design. Lead-driven and cheap — no agent team; the first reaction point, before any design or
code is committed. Direction is hardened by reacting to a visual shape, not by reading prose.

## The artifact
`folio new sketch <capability>` scaffolds `work/active/<date>-<topic>/reference/sketch/index.html`
from an embedded seed and creates the work dir (like `design`). One self-contained HTML file.

**Title** = a capability **noun-phrase** — no `type(scope)`. It must NOT parse as a
conventional-commit header (the design phase mints that later from this phrase).

## Layout catalog (fill the seed's slots)
- **One-liner** — the whole idea in one sentence.
- **Shape** — stages/pieces/flow as co-located cards (each card: what it is + what it produces).
- **At a glance** — key pieces/systems, one fact each (a table).
- **Supports / doesn't** — what's in scope vs deliberately out (the right column is where
  reactions land).
- **Open questions** — what the user's reaction resolves.

## Build conventions (HTML-first)
Adopt the self-containment + aesthetic conventions of the interactive-artifact ecosystem, minus the
interactivity (see Read-only below):
- **Single self-contained HTML file. Inline all CSS/JS. NO external deps / CDNs.**
- **HTML-first for information and visuals** — cards, tables, ladders carry the content. This is the
  default and usually sufficient.
- **A graph diagram (inline SVG) only when a visual genuinely needs 2D structure** (system map,
  dependency graph) AND it's asked for. Not by default. If a text-DSL (e.g. D2) is compiled to SVG,
  strip its opaque page background (both the background rect and the `.fill-N7`-style rule that
  overrides fills) and the root `<svg>` width/height, and cap the rendered size so it stays
  proportional to the page.
- Dark-mode via `prefers-color-scheme`; sensible defaults; scannable density.
- **No private/internal tool or project names** — the sketch is the shareable surface. Local
  markdown planning docs may name things freely.
- Roughly-right, then frozen — not iterated line-by-line (HTML diffs coarsely; that's acceptable
  for a layer meant to be socialized and frozen).

## Read-only (v1)
The sketch is **read-only**. The react loop runs conversationally: show the rendered page, the user
reacts, the lead rebuilds. That IS the round-trip, human-mediated through the session — no listener
or in-page controls (deferred; interactivity is a prompt-injection surface and redundant in-session).

## The birds-eye sign-off gate (Hard)
**Show the artifact, don't summarize it** (see the standing rule). Open/render the page for the
user, run the react→revise loop, and freeze only on explicit approval. Commit the frozen sketch via
`folio home push` **before** Phase 2 begins.

## Lightweight decision (made once, here)
At sketch-freeze, write the candidate track-title list — one `type(scope): description` line per
prospective track. Let **N** = count of distinct lines.
- **N == 1** (single clear change, single repo) → **lightweight**: skip design + brief; create the
  one `track-1.md` and go straight to implementation + commits. The sketch + that track file are the
  committed plan.
- **N ≥ 2** → **full pipeline**: Phase 2 (propose) → design doc → brief → (burndown?) → execute.

A cheap forward estimate — naming tracks, not writing briefs — co-signalled by scope clarity.

## Handoff
Full → continue to Phase 2 (`plan-design.md`), now seeded by the frozen sketch (its shape informs
lens selection). Lightweight → skip to execution (`plan-execute.md`) with the single track file.
