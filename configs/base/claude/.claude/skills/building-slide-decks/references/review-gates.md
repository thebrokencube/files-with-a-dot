# Review Gates

Run these checks before finalizing a deck. The persona reviews are optional but strongly recommended for any deck that will be presented to an audience.

## Pre-review checklist (run every time)

Before persona reviews, verify:

- [ ] **Every slide has `data-notes`** — no empty or missing notes
- [ ] **Links exist where referenced** — repos, docs, channels, tools mentioned by name should be `<a class="deck-link">` links
- [ ] **Link audit** — open each link target mentally: is the URL correct? Does it point to the right branch/file?
- [ ] **Slide density** — no slide has more than 5 major elements; no slide has fewer than 2
- [ ] **Content accuracy** — read the actual source material (repo, docs, code), not design docs or memory
- [ ] **Typography hierarchy** — each slide uses section-label → t-title → content. Not all t-body.
- [ ] **Dark-bg distribution** — max 2-3 dark-bg slides per 7-slide deck. Used for emphasis, not decoration.
- [ ] **SVG viewBox padding** — verify no text is clipped at edges
- [ ] **Closing slide** — has clear next-steps or takeaway, not just a list of links

## Link review (run every time)

Scan every slide for references that should be linked but aren't:

- Repository names (e.g., "guideline-app/radr") → GitHub URL
- File names (e.g., "CONTRIBUTING.md") → GitHub blob URL
- Slack channels (e.g., "#retirement-architecture") → Slack deep link
- Tools or skills mentioned (e.g., "/write-radr") → documentation URL if available
- People's names → skip (don't link to profiles)

For each link, verify the `<a>` tag:
- Has `class="deck-link"` for consistent styling
- Has `target="_blank"` to open in new tab
- Is correctly excluded from slide navigation (JS engine handles this)

## Persona reviews (near-final decks)

Launch 3-5 parallel subagent reviews. Each agent reads the HTML file and provides feedback from their persona's perspective. Tailor personas to the specific audience.

### Default personas

For a technical team introduction deck:

1. **New team member** — "I just joined. Do I understand what this is and how to get started?"
2. **Skeptical senior IC** — "Why should I add this to my workflow? What's the overhead?"
3. **Engineering manager** — "Can I coach my team on this? Is the process clear?"
4. **Adjacent-team engineer** — "How does this relate to what I already know?"

### Persona review prompt template

```
You are reviewing a slide deck as a **{persona description}**.

Read the slide deck at: {file_path}

Then provide a review covering:
1. **Clarity**: After reading all slides, do you understand the main message? What's confusing?
2. **Actionability**: Could you take the next step after seeing this? What's missing?
3. **Density**: Too much on any slide? Too little?
4. **Gaps**: What questions would you still have?
5. **Specific suggestions**: Concrete changes to improve the deck for this persona.

Do NOT write any code or edit any files. Just provide the review.
```

### Synthesizing reviews

After all persona reviews return:
1. **Convergent findings** (multiple personas flagged the same thing) = high priority, fix these
2. **Divergent findings** (one persona's concern) = consider but don't over-rotate
3. **"Best slide" consensus** = the strongest proof point, protect it during edits
4. Present a synthesis to the user with recommended changes, ranked by impact
