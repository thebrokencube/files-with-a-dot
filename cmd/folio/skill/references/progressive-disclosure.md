# Progressive Disclosure

Cross-cutting principle for folio artifacts that cross an agent or session boundary:
briefs, design docs, compose outputs.

## Principle

**Action first, context second, history last.** A reader who knows what to do reads 10
lines and starts working. A reader who needs to understand why reads deeper. No reader
should need to read the whole document before acting.

## Test

If someone can't act from the top 10 lines of the document, the structure is wrong.

## Two-Layer Structure (Briefs)

Briefs use two layers:

1. **Action layer** (top): Track instructions — what to change, in what order, commit
   sequence. The execution agent reads this and can start working.
2. **Reference layer** (bottom): Key Reference sections — design decisions, templates,
   protocols, escalation tables. The agent drills into these when a track instruction
   says "see Key Reference" or when judgment is required.

The action layer should be self-sufficient for mechanical steps. The reference layer provides
context for judgment calls. If a track step can't be executed without reading a Key Reference,
the track instruction is underspecified — add more inline detail.

## Session Handoffs Are Folio State

Folio does NOT use handoff documents. Cross-session continuity is provided by the design
doc + observations + work-track filesystem. The next session resumes by running
`/folio <project>`, which surfaces the current design state inline.

When designing a doc that will need to survive a session boundary, optimize the design
doc itself for progressive disclosure: Pinned Constraints + Open Questions + Convergence
Status at the top (always read), the Direction / Interfaces / Testing Strategy bodies in
the middle (read if rationale is needed), Design Provenance at the bottom (history). See
plan-design.md "Phase 4a: Fill Design Doc" for the layer-tagged template.

## Application

This principle applies to any artifact that crosses an agent boundary or session
boundary. When writing or reviewing such an artifact, check: can the consumer act from
the top section alone? If not, promote the missing information upward.
