# Adversarial Review Principle

Cross-cutting principle applied wherever an agent exercises subjective judgment.
Referenced by plan-design, plan-brief, plan-execute, compose, gather, and observe workflows.

## Principle

Every subjective judgment needs a pushback mechanism. The mechanism scales with stakes:

| Tier | Mechanism | Cost | When to use |
|------|-----------|------|-------------|
| **Self-challenge** | Agent states what it rejected and why before proceeding | Zero — prompt language | Any unchecked judgment: context compilation, threshold decisions, classification calls, sufficiency assessments |
| **Adversarial prompt** | Single agent gets pushback instructions: "Challenge your own conclusion. What's the strongest argument against?" | Low — added instructions | Medium-stakes judgment within a review step: brief reviews, code reviews, compose reviews |
| **Parallel adversarial** | Dedicated agents reviewing from opposing perspectives | High — 2+ agents | High-stakes judgment that shapes downstream work: design reviews |

## Self-Challenge Format

When an agent makes a subjective judgment at the self-challenge tier, it must state:

```
Judgment: [what was decided]
Rejected: [the alternative(s) considered]
Why: [one sentence on why the chosen option wins]
```

This makes reasoning visible without adding cost. Downstream reviewers or the user can
spot bad judgment from the visible reasoning.

## Application

The tiers are applied per-workflow in the workflow reference files themselves. The key
structural applications:

- **Phase 4b design review** (plan-design.md): Parallel adversarial (2 agents). In
  lightweight mode, falls back to adversarial prompt (1 agent).
- **Phase 6 brief review** (plan-brief.md): Adversarial prompt dimension added.
- **Phase 7 code review** (plan-execute.md): Adversarial prompt dimension added.
- **Compose review** (compose.md): Adversarial prompt dimension added.

Self-challenge is a general expectation for any subjective judgment — agents should state
what they rejected and why when making non-obvious decisions. This is a behavioral norm,
not a per-judgment-point checklist.
