# Progressive Disclosure

Cross-cutting principle for all folio artifacts that hand off context: briefs, handoff
documents, design docs, compose outputs, handoff prompts.

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

## Three-Layer Structure (Handoff Documents)

Session handoffs use three layers:

1. **Layer 1 — TL;DR + Start Here** (always read): 3-sentence summary + copy-paste prompt
   for the next session. Self-contained — works without reading anything else.
2. **Layer 2 — Context** (read if the session needs to understand why): Open questions,
   key decisions from this session only, exit criteria. Under 30 lines. Link to design
   doc for full context rather than restating it.
3. **Layer 3 — Reference** (skim or skip): Artifact file tree, folio project path, temp
   files. Compact — no descriptions unless the path isn't self-explanatory.

## Application

This principle applies to any artifact that crosses a session boundary or agent boundary.
When writing or reviewing a handoff artifact, check: can the consumer act from the top
section alone? If not, promote the missing information upward.
