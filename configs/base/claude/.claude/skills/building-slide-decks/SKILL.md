---
name: building-slide-decks
description: Build self-contained HTML slide decks with keyboard navigation, speaker notes, and responsive typography. Use when the user asks for a presentation, slide deck, or slides. Produces a single HTML file — no external dependencies.
user_invocable: true
---

# Building Slide Decks

Build self-contained HTML presentations. Single file, inline CSS/JS, no dependencies. Arrow keys navigate, N toggles speaker notes, progress bar tracks position.

## When to use

When the user asks for a presentation, slide deck, or slides about any topic. NOT for interactive playgrounds or explorers (use playground skill instead).

## Workflow

1. **Gather content** — Understand the topic. Read source material. Identify the audience.
2. **Plan structure** — Outline slides (aim for 5-8 for a 3-5 min deck). Each slide needs a clear purpose.
3. **Pick a palette** — Two accent colors (primary + highlight) on a light (#F5F5F5) background. Dark-bg slides use the primary color as background for emphasis/section breaks.
4. **Generate HTML** — Read `references/css-system.md` for the design system and `references/js-engine.md` for the navigation engine. Build the deck.
5. **Open in browser** — `open <filename>.html` to preview.
6. **Iterate** — Fix issues the user flags. Re-read reference files as needed.
7. **Review** — Run persona reviews when the deck is near-final. Read `references/review-gates.md`.
8. **Commit** — Use the project's commit conventions.

## Core requirements

Every deck MUST have:

- **Single HTML file** — all CSS and JS inline. No CDN, no external fonts.
- **Keyboard navigation** — Arrow keys, Home/End, click left/right halves.
- **Speaker notes** — `data-notes` attribute on every slide. N key toggles notes panel.
- **Progress bar** — fixed bottom, accent-colored, auto-advancing.
- **Slide counter** — fixed bottom-right, monospace.
- **Responsive typography** — all text sizes use `clamp()`. No fixed px for body text.
- **Links don't navigate** — clicking `<a>` tags must not advance slides.
- **Content centering** — `.slide-inner` wrapper, `max-width: 960px`.

## Slide density

- **3-5 major elements per slide** — title + 2-3 content blocks. If you need more, split.
- **~500 word budget** for speaker notes per slide. Notes carry the depth; slides carry the signal.
- **Every slide answers**: "why are we talking about this now?" Bridge from the previous slide.

## Visual elements

| Element | When to use | Reference |
|---------|-------------|-----------|
| Bullet list | 3-5 items with labels | `references/css-system.md` § bullet-list |
| Code block | CLI commands, file content, config | `references/css-system.md` § code-block |
| Comparison table | Side-by-side options | CSS grid, 3-col layout |
| SVG diagram | Architecture, flow, relationships | `references/svg-patterns.md` |
| Two-column layout | Code + explanation, list + list | Flexbox with gap |

## Color system

The deck uses CSS custom properties. Define these in `:root`:

- `--primary` — dark accent (headings, emphasis, dark-bg background)
- `--primary-light` — lighter variant for SVG/diagram elements
- `--highlight` — contrasting accent (CTAs, labels, key terms)
- `--highlight-light` — lighter variant for dark-bg contexts
- `--bg: #F5F5F5` — page background (always light)
- `--gray`, `--dark-gray`, `--light-gray` — neutral tones (unchanged across decks)
- `--code-bg` — dark code block background

Previous decks used: navy/amber, teal/coral. Pick colors that fit the topic.

## Slide types

- **Light** (default) — light background, primary-colored headings
- **Dark** (`.dark-bg`) — primary-colored background, light text. Use for emphasis, section breaks, architecture diagrams. Max 2-3 per deck.

## Links

Add links where they provide value — repo URLs, documentation, Slack channels, specific files referenced. Style with `a.deck-link` (underline in highlight color). The JS engine must skip `<a>` clicks for slide navigation.

## Common mistakes

- **SVG viewBox too tight** — add 10-20px padding beyond content bounds
- **No speaker notes** — every slide needs `data-notes`
- **Missing links** — if you reference a repo, doc, or channel, link it
- **Too much text** — slides are visual anchors, not documents
- **Inconsistent slide density** — one slide with 2 items, next with 8. Redistribute.
- **Closing slide undercooked** — closing slides need the most iteration. Budget for it.

## Reference files

- **references/css-system.md** — Full CSS design system (typography, utilities, components)
- **references/js-engine.md** — Navigation JS (keyboard, click, notes panel, progress bar)
- **references/svg-patterns.md** — Reusable SVG diagram patterns (flow, architecture, comparison)
- **references/review-gates.md** — Persona review prompts and quality checklists
