# Alignment Protocol

Read by host workflows (plan, compose, observe) at their alignment points.
Not a standalone skill — always invoked with parameters from the calling workflow.

## Invocation Contract

The calling workflow provides five parameters as prose context in its step instruction:

| Parameter | Type | Description |
|-----------|------|-------------|
| minimum | int | Minimum questions before exit is eligible (caller sets) |
| categories | list | Decision categories that must reach CONFIRMED before exit |
| grounding | list | Sources to ground against (file paths, vault refs, folio sources) |
| target | string | What to update with decisions (e.g., "context summary", "ephemeral how-amendment") |
| hard_constraints | list | User-stated decisions that are non-negotiable (alignment must not re-evaluate) |

## Question Format

Each question uses the claim-first format:

```
**Claim**: [What the AI believes]
**Evidence**: [File path, line reference, or vault entry that supports the claim]
**Recommended**: [What the AI recommends based on the evidence]
```

Every claim MUST cite a file path, code reference, or vault entry. "Based on context"
is not valid evidence. If you cannot ground a claim in a specific source, resolve the
factual question yourself by reading the codebase — do not surface it to the user.

## User Responses

Five response types, each with a specific outcome:

- **Confirm** — claim becomes a decision, materialized into the target immediately
- **Override** — user's version replaces the claim and becomes the decision
- **"Figure it out"** — AI adopts its own recommendation, tagged `[AI-DECIDED]` in the target
- **"Skip"** — question dropped, no decision recorded
- **"Enough"** — exit the alignment (only after minimum count reached), remaining categories marked UNRESOLVED

## Behavioral Rules

1. **Resolve factual questions silently.** Look up facts from the codebase, vault, and
   folio sources yourself. Only surface judgment questions to the user — scope decisions,
   trade-offs, term definitions, framing choices.
2. **One question at a time.** Present one claim, wait for the user's response before
   presenting the next. Do not batch questions.
3. **Inline materialization.** Each resolved question updates the target immediately.
   Do not accumulate decisions for a batch update at the end.
4. **Confidence checkpoint.** After the minimum count, check category coverage. When all
   categories are CONFIRMED (or AI-DECIDED), present the Confidence Checkpoint and wait
   for "Proceed" before continuing. Do not silently move on.
5. **Gap recovery.** If the user identifies an unresolved category after the checkpoint
   ("wait, we didn't cover X"), resume alignment from that category without restarting.
   Back-reference the checkpoint in the updated summary.

## Confidence-Based Exit

The alignment runs until all decision categories reach CONFIRMED or AI-DECIDED status —
not until a fixed count is reached. The minimum prevents premature exit.

- After minimum count: check category coverage. If all categories resolved, present
  the Confidence Checkpoint. If categories remain, continue questioning.
- **"Enough"** after the minimum exits early — the checkpoint shows unresolved categories
  as `[UNRESOLVED — user exited]`. "Enough" is not available before the minimum.
- There is no "Keep going" mechanic. If categories remain unresolved, the alignment
  continues naturally. If all are resolved and the user wants to go deeper, gap recovery
  handles it.

## Confidence Checkpoint

Present this when all categories are resolved (or after "Enough"):

```
Alignment checkpoint:
- [Category 1]: CONFIRMED (Q[N] — [one-line rationale])
- [Category 2]: CONFIRMED (Q[N] — [one-line rationale])
- [Category 3]: [AI-DECIDED] (Q[N] — [one-line rationale])
- Uncovered: [category] ([reason — e.g., "not applicable at this scope"])

I believe we're aligned on the key decisions. Proceed?
```

Each CONFIRMED line MUST back-reference the question number that resolved it.
Uncovered categories are explicit acknowledgments, not failures. `[AI-DECIDED]`
entries display distinctly so the user can see which decisions they delegated.

## AI-DECIDED Tagging

When the user responds with "Figure it out", the AI adopts its own recommendation and
tags the decision `[AI-DECIDED]` in the target. This tag serves two purposes:

1. Review agents (e.g., Phase 4b design review) can flag AI-decided items for
   re-examination if they look questionable
2. The user can scan the target for `[AI-DECIDED]` to see which decisions they
   delegated vs confirmed explicitly

The tag is lightweight metadata, not a quality judgment — it simply marks provenance.
