# Alignment Protocol

Read by host workflows (plan, compose, observe) at their alignment points.
Not a standalone skill — always invoked with parameters from the calling workflow.

## Invocation Contract

The calling workflow provides four parameters as prose context in its step instruction:

| Parameter | Type | Description |
|-----------|------|-------------|
| budget | int | Default question count (caller sets; user can exit early with "enough" or extend with "keep going") |
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
- **"Enough"** — exit the alignment, remaining questions dropped

## Behavioral Rules

1. **Resolve factual questions silently.** Look up facts from the codebase, vault, and
   folio sources yourself. Only surface judgment questions to the user — scope decisions,
   trade-offs, term definitions, framing choices.
2. **One question at a time.** Present one claim, wait for the user's response before
   presenting the next. Do not batch questions.
3. **Inline materialization.** Each resolved question updates the target immediately.
   Do not accumulate decisions for a batch update at the end.
4. **Budget pause.** At the budget count, pause and ask "Continue or wrap up?" Do not
   silently stop asking questions. Do not silently continue past the budget.

## Budget Mechanics

The budget is a default, not a hard cap. The user controls the actual duration:

- **"Enough"** at any point exits early — remaining questions are dropped
- **"Keep going"** after the budget pause extends the alignment
- At the budget count, always pause: "Continue or wrap up?"

The calling workflow sets the budget based on alignment depth needed (e.g., 7 for plan,
4 for compose, 2 for observe). The user's response overrides this in either direction.

## AI-DECIDED Tagging

When the user responds with "Figure it out", the AI adopts its own recommendation and
tags the decision `[AI-DECIDED]` in the target. This tag serves two purposes:

1. Review agents (e.g., Phase 4b design review) can flag AI-decided items for
   re-examination if they look questionable
2. The user can scan the target for `[AI-DECIDED]` to see which decisions they
   delegated vs confirmed explicitly

The tag is lightweight metadata, not a quality judgment — it simply marks provenance.
