# Plan Workflow — Idea/Architecture Phase (Phase 1.5)

Read by `/folio plan` after Phase 1 (Understand), before Phase 2 (Propose). Produces the
**idea/architecture sketch**: a single self-contained HTML page you react to *before* committing to a
design — the first reaction point, before any design or code is committed. Lead-driven and cheap: no
agent team builds it; one fresh subagent reviews it before you see it. Direction is hardened by
reacting to a visual shape, not by reading prose.

## The artifact
`folio new sketch <capability>` scaffolds `work/active/<date>-<topic>/reference/sketch/index.html`
from an embedded seed and creates the work dir (like `design`). One self-contained HTML file. The
seed is a **minimal skeleton** — replace its card/table slots with the visuals below; do not treat
the seed's shape as the target.

**Title** = a capability **noun-phrase** — no `type(scope)`. It must NOT parse as a
conventional-commit header (the design phase mints that later from this phrase).

## The one rule: lead with the strongest visual, cut paragraphs
A sketch is a *picture of the data model and its flow*, not an essay about it. Before writing a
sentence, ask what diagram it would be instead. The mechanical form that enforces this: **each
`<h2>` section is one visual plus at most a one-line caption; the only multi-line prose on the whole
page is the single `.one` one-liner.** First drafts fail by being wordy — and the fix is never a
shorter paragraph, it's a diagram where the paragraph was.

## Two archetypes (pick the lede)
The two exemplar sketches — **per-repo-warden-capabilities** and **github-anchored-review-dashboards**
(each at a project's `work/active/<date>-<topic>/reference/sketch/index.html`) — are the bar. They
show the two shapes a sketch takes; pick by what the idea is:
- **Data model** (default, for a schema / record change) — ER entity boxes + relationships as the
  lede, a concrete record, and a data-flow SVG. **Bias hard toward this style.**
- **Narrative of models** (for a reframe or architecture shift) — 2-4 numbered `<h2>` sections, each
  one SVG, sequencing today → the idea → what it unlocks → the invariant. No ER boxes; the payoff is
  the sequence, anchored by an invariant callout.

## The vocabulary
Reach for these first. Everything data-shaped is monospace; everything else is a one-line caption.
The named classes come from the two exemplar sketches (the seed ships only the skeleton's classes) —
reuse them so every sketch reads as one system.

| Element | What it is | When |
|---|---|---|
| **ER entity boxes** | one monospace box per entity: a title row tagged `new` \| `existing` \| `code-registry`, then `name : type` field rows. PK/FK marked; removed fields shown with ~~strikethrough~~ (`.del`); `uniq(...)` noted; relationship lines (`repos 1 ──< N configs`) drawn between boxes. `new` boxes highlighted (`.ent.new`), `code-registry` boxes dashed (`.ent.code`). | data-model lede |
| **Data-flow SVG** | inline `<svg>` of boxes + arrows: a record fanning into the gates or surfaces that read each field, one arrow per consumer | a record drives behavior at 2+ points |
| **Concrete example** | one real record as a syntax-colored `<pre>` (key=accent, string=green, number=amber, comment=muted) — optionally beside a small reference table with an italic `future`/drop-in row for extensibility — or one CLI call + its output | data-model / record-shaped sketches |
| **Open-forks cards** (`.fork`) | compact cards, one per unresolved decision; each states the options with the picked one marked `▸` (`.pick`) | the reaction surface — where the user's call lands |
| **Invariant callout** (`.inv`) | one centered, accent-bordered line stating a law the design must hold (counts match, never-drop) | when the idea hinges on a constraint |
| **Monospace** | field names, types, slugs, paths, values — anything data-shaped | pervasive |

Cards and tables are the fallback only for content with genuinely no 2D structure — never the
opening move.

## Layout catalog (fill the seed's slots)
Each slot is a **visual** wherever the content admits one; prose drops to a one-line caption beneath
it. For a **data-model** sketch:
- **One-liner** — the whole idea in one sentence (the `.one` callout).
- **Data model** — ER entity boxes + relationship lines (the lede).
- **One row / call, concretely** — the concrete example.
- **Data flow** — the data-flow SVG.
- **Open forks** — open-forks cards, the picked option marked `▸`. This is where reactions land.

For a **narrative-of-models** sketch the slots become the sequence: the one-liner, then 2-4 numbered
`<h2>` model sections (each one SVG + caption), an invariant callout, and a closing "Decide" block
of open-forks cards.

## Build conventions (visual-first, self-contained)
- **Single self-contained HTML file. Inline all CSS / JS / SVG. NO external deps / CDNs.**
- **Derive every tint from palette vars via `color-mix(in srgb, var(--x) N%, transparent)`** —
  callout backgrounds, entity highlights, card fills, and SVG box fills. Never hardcode a fill, so
  light and dark both hold.
- **Inline SVG is the default medium, not an exception.** Hand-author it, or compile a text-DSL
  (e.g. D2) to SVG; if compiled, strip its opaque page background (both the background rect and the
  `.fill-N7`-style rule that overrides fills) and the root `<svg>` width/height (**but keep the
  `viewBox`** — see the reviewer gate), and cap the rendered size so it stays proportional to the
  page. For repeated iconography, inline a `<symbol>` sprite and reference it with `<use>`.
- **`<defs>` holds only `<marker>`, `<symbol>`, gradients, and clip paths** — never renderable
  shapes (a shape there renders blank; the reviewer gate verifies this). Arrow endpoints must land on
  the boxes they connect, and all content stays inside the `viewBox`.
- When the content isn't a data model — a UI, a reframe, components/ownership, a hierarchy — reach
  for the matching visual: **wireframe** (app frame mocking real screens), **system map** (SVG boxes
  + arrows), **pipeline / fan-out** diagram, **positioning chart** (2 labeled axes + plotted points
  + a move-arrow, for "these properties are independent" or "we're moving up this axis" arguments),
  or **monospace tree** (URL schemas, directory layouts).
- Dark-mode via `prefers-color-scheme`; drive the palette from CSS vars so SVG and HTML share it.
  Keep it scannably dense.
- **No customer data or secrets.** Code identifiers — table / class / field / slug names — are
  expected; they're what make the model concrete (the exemplars use them freely). Scrub only if the
  sketch will be shared outside the team; local markdown planning docs may name anything.
- Roughly-right, then frozen — not iterated line-by-line (HTML/SVG diff coarsely; that's acceptable
  for a layer meant to be socialized and frozen).

## Read-only (v1)
The sketch is **read-only**. The react loop runs conversationally: show the rendered page, the user
reacts, the lead rebuilds. That IS the round-trip, human-mediated through the session — no listener
or in-page controls (deferred; interactivity is a prompt-injection surface and redundant in-session).

## Reviewer gate (fresh subagent — before the sign-off gate)
After building or editing the sketch, and **before rendering it for the user**, run one fresh
reviewer subagent over the HTML. It shares Phase 4b's review *structure* (a fresh subagent, a
fix-loop, run before the artifact reaches the user) but is a **defect / rendering QA pass — not one
of the adversarial tiers** in `references/adversarial-review.md`; it hunts for bugs, not
counter-arguments.
Use `subagent_type: general-purpose`, session default model; pass it the sketch path, the repo
root(s), and the local planning notes.

Four dimensions — return concrete defects (what's wrong + the fix):

1. **SVG blank-render bugs (highest priority — these silently ship an invisible diagram; every one
   is statically checkable).**
   - **Shapes in `<defs>`** — a `<rect>`/`<circle>`/`<path>`/`<text>`/`<line>` inside `<defs>` never
     paints (`<defs>` is markers/symbols/gradients/clip-paths only). **This exact bug has shipped a
     blank diagram.**
   - **Dangling refs** — every `url(#id)` (marker / fill / stroke / clip) and `<use href="#id">`
     must resolve to an element with that `id` in the same file, or it silently fails to paint.
   - **Missing `viewBox`** — every inline `<svg>` needs a `viewBox`; a compile-and-stripped SVG with
     `height:auto` and no `viewBox` collapses to zero height (whole diagram blank).
   - **Unset `<text>` fill** — every `<text>`/`<tspan>` needs an explicit `fill` that isn't the
     background var; the default is black and vanishes on the dark-mode `--bg`.
   - **Undefined `var(--…)`** — a `var()` fill/stroke must be declared in `:root`, the dark-mode
     media block, or an inline `style`.
   - **Disconnected arrows** — endpoints must land on the source/target `<rect>`. For `<line>` use
     `(x1,y1)/(x2,y2)`; for a `<path>` connector the endpoint is the **last coord pair in `d`**
     (`M … C x1 y1, x2 y2, x y` → `x y`). Confirm each endpoint sits within ~`markerWidth` of a
     rect's edge, not floating in whitespace.
   - **Out-of-viewBox / hidden** — coords outside the `viewBox`, or an opaque rect fully covering an
     earlier element (later paint wins). Fine legibility/overflow can't be judged statically — leave
     it to the human render gate.
2. **Prose bloat.** Flag any `<h2>` section with no visual, and any caption longer than one sentence
   (the mechanical rule above makes this countable, not a judgment call).
3. **Missing data-model / relationships.** For a data-model sketch: are the entities, their
   `new|existing|code-registry` tags, and the relationship lines actually *drawn* — not just
   described in prose?
4. **Codebase contradictions.** Treat each ER box title as a table/class, each field row as a
   column + type, each slug / value / captioned behavior / PR number as a claim; grep the codebase
   (schema / migration, class / module, constant, PR) and flag anything that doesn't exist or
   differs. A wrong premise wastes the whole round.

**Loop:** blank-render (dimension-1) defects are mechanical and objective — keep fixing them; never
surface-and-proceed on a broken diagram, and escalate to the user if one truly won't resolve. For
judgment defects (dimensions 2-4), cap at 2 iterations — lower than Phase 4b's 3, since the sketch is
a single roughly-right file — then surface any remainder alongside the sketch. A clean gate means no
*statically detectable* defect, not a verified render; the birds-eye gate below is the human render
check.

## The birds-eye sign-off gate (Hard)
Runs after the reviewer gate passes. **Show the artifact, don't summarize it** (see the standing
rule). Open/render the page for the user, run the react→revise loop (each revise re-runs the
reviewer gate), and freeze only on explicit approval. Commit the frozen sketch via `folio home push`
**before** Phase 2 begins.

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
