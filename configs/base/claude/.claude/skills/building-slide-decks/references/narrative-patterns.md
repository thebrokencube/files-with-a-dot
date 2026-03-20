# Narrative Patterns

Lessons from building and iterating on narrative-driven slide decks. These complement the visual design system — they're about story structure, tone, and how content flows.

## The pause beat pattern

Self-deprecating asides, punchlines, and editorial commentary go AFTER the content they react to. Never between two related visual elements.

```
WRONG:  intro text → beat → SVG diagram
WRONG:  SVG diagram → beat → code block (if they show the same concept)
RIGHT:  intro text → SVG diagram → code block → beat
RIGHT:  intro text → content → beat
```

The beat is a punchline. Setup → content → punchline. Putting it in the middle breaks the visual flow and makes the slide feel disjointed.

## Slide cohesiveness

Each slide should have one visual flow, not competing elements. Signs a slide needs splitting:

- SVG diagram + code block + prose all fighting for attention
- Side-by-side layout where both sides are dense
- Story setup AND payoff on the same slide when both are substantial

When splitting: one slide for the story/setup, next slide for the evidence/payoff.

## Grounding real examples

When referencing real projects or domain-specific work:

- **Ground the audience** with 1-2 sentences of context (what the project is, what matters)
- **Focus on the concept**, not the implementation details — the audience shouldn't need to learn your domain
- **Use real identifiers** (ticket numbers, file names) to feel authentic, but don't explain them
- **Don't over-specify** — "compliance deadline requirements" is enough; "EndOfYear DueAt rounding in the compliance engine's Rounding T::Enum" is too much

## Portraying cross-team collaborators

When the narrative involves other teams catching problems or contributing expertise:

- Frame them as **partners contributing expertise**, not obstacles or error-catchers
- "Brought their domain expertise to bear" > "said 'actually, no'"
- "Asked the right questions" > "sent corrections"
- If their work is central to the story, name the team specifically

## Terminal/code mockups

- **3-4 representative items**, not a full dump
- Slightly larger font than you think — these are presentation slides, not terminal windows
- One item per category/concept is enough to show the pattern
- The comment showing "... N more" signals scale without overwhelming

## Narrative framing for personal infrastructure

When the deck is about tools the presenter built:

- **"Built from friction, not foresight"** — positions as evolved pragmatism, not architecture astronautics
- Clearly signal what's **novel** (worth stealing) vs. **plumbing** (useful to the presenter, not transferable)
- **"Steal this"** framing on close — concrete, actionable, respects the audience's time
- Self-deprecation helps relatability but shouldn't undercut the strongest evidence (dial it back on the money slides)

## Understanding vs. requirements

When discussing shifting requirements:

- **"Understanding evolves"** > "requirements aren't facts" — the latter sounds dismissive of domain expertise
- Frame it as *your* understanding shifting, not the requirements being unreliable
- Requirements are the ground truth; your interpretation is what changes
