# CSS Design System

Extracted from two production decks. Copy and adapt — colors are parameterized via CSS custom properties.

## Base reset and layout

```css
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

:root {
  /* Replace these per deck */
  --primary: #1B2A4A;
  --primary-light: #2D4470;
  --highlight: #E8943A;
  --highlight-light: #F0AD60;
  /* These stay constant */
  --bg: #F5F5F5;
  --gray: #6B7280;
  --dark-gray: #4B5563;
  --light-gray: #E5E7EB;
  --code-bg: #141E30;
  --code-comment: #6B7994;
  --font: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  --mono: 'SF Mono', 'Fira Code', 'Fira Mono', Menlo, Consolas, monospace;
}

html, body { height: 100%; background: var(--bg); color: var(--primary); font-family: var(--font); overflow: hidden; }
.deck { position: relative; width: 100%; height: 100%; }
```

## Slide structure

```css
.slide {
  position: absolute; inset: 0;
  display: flex; flex-direction: column; justify-content: center; align-items: center;
  padding: 48px 8vw 64px;
  visibility: hidden; opacity: 0; transition: opacity 0.35s ease;
}
.slide.active { visibility: visible; opacity: 1; }
.slide.dark-bg { background: var(--primary); color: #E0E4EA; }
.slide-inner { max-width: 960px; width: 100%; }
```

Use `visibility: hidden/visible` + opacity for transitions. No transforms — they cause SVG rendering issues.

## Typography

All sizes use `clamp(min, preferred, max)` for responsive scaling.

```css
.t-hero   { font-size: clamp(2.4rem, 5vw, 3.6rem); font-weight: 800; color: var(--primary); line-height: 1.08; letter-spacing: -0.025em; }
.t-title  { font-size: clamp(1.6rem, 3vw, 2.2rem); font-weight: 700; color: var(--primary); line-height: 1.15; letter-spacing: -0.015em; margin-bottom: 0.6em; }
.t-large  { font-size: clamp(1.15rem, 1.8vw, 1.4rem); line-height: 1.5; color: var(--primary); }
.t-body   { font-size: clamp(0.95rem, 1.4vw, 1.15rem); line-height: 1.65; color: var(--dark-gray); }
.t-small  { font-size: clamp(0.8rem, 1.05vw, 0.92rem); color: var(--gray); line-height: 1.5; }
.t-mono   { font-family: var(--mono); font-size: 0.88em; }
```

## Color utilities

```css
.c-primary { color: var(--primary); }
.c-highlight { color: var(--highlight); }
.c-gray { color: var(--gray); }
.bold { font-weight: 700; } .semi { font-weight: 600; } .italic { font-style: italic; }
.centered { text-align: center; }
```

## Spacing utilities

```css
.mb-xs { margin-bottom: 0.3em; }
.mb-sm { margin-bottom: 0.6em; }
.mb-md { margin-bottom: 1em; }
.mb-lg { margin-bottom: 1.6em; }
.mb-xl { margin-bottom: 2.4em; }
.mt-sm { margin-top: 0.5em; }
.mt-md { margin-top: 1em; }
.mt-lg { margin-top: 1.6em; }
```

## Dark-bg overrides

```css
.dark-bg .t-hero, .dark-bg .t-title { color: #F0F2F5; }
.dark-bg .t-large { color: #D4D8E0; }
.dark-bg .t-body  { color: #A8B0BE; }
```

## Section label

Small uppercase label above slide titles (e.g., "WHY", "HOW", "ARCHITECTURE").

```css
.section-label {
  font-size: clamp(0.72rem, 0.9vw, 0.82rem);
  font-weight: 700; letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--highlight); margin-bottom: 0.6em;
}
.dark-bg .section-label { color: var(--highlight-light); }
```

## Bullet list

Separated items with bottom borders. Use `.bl` for primary-colored labels, `.ba` for highlight-colored labels.

```css
.bullet-list { list-style: none; padding: 0; }
.bullet-list li {
  padding: 12px 0; border-bottom: 1px solid var(--light-gray);
  font-size: clamp(0.92rem, 1.25vw, 1.05rem); line-height: 1.55;
  color: var(--dark-gray);
}
.bullet-list li:last-child { border-bottom: none; }
.bullet-list .bl { font-weight: 700; color: var(--primary); }
.bullet-list .ba { font-weight: 700; color: var(--highlight); }
```

## Code block

Dark-background code display with syntax highlighting spans.

```css
.code-block {
  background: var(--code-bg); border-radius: 10px;
  padding: 24px 28px;
  font-family: var(--mono);
  font-size: clamp(0.78rem, 1.1vw, 0.92rem);
  line-height: 1.75; overflow-x: auto; white-space: pre;
  color: #CBD5E1;
}
.code-block .ck { color: var(--highlight); }        /* keyword/command */
.code-block .cv { color: #E0E4EA; }                 /* value */
.code-block .cc { color: var(--code-comment); }      /* comment */
.code-block .ca { color: var(--highlight); font-weight: 600; } /* accent */
```

## Links

```css
a.deck-link {
  color: inherit;
  text-decoration: underline;
  text-decoration-color: var(--highlight);
  text-underline-offset: 3px;
  text-decoration-thickness: 2px;
  transition: text-decoration-color 0.2s;
}
a.deck-link:hover { text-decoration-color: var(--primary-light); }
.dark-bg a.deck-link { text-decoration-color: var(--highlight-light); }
.dark-bg a.deck-link:hover { text-decoration-color: #fff; }
```

## Chrome (progress bar, counter, key hint, notes panel)

```css
.progress-bar {
  position: fixed; bottom: 0; left: 0; height: 3px;
  background: var(--highlight); transition: width 0.3s ease; z-index: 100;
}
.slide-counter {
  position: fixed; bottom: 12px; right: 20px;
  font-size: 0.78rem; color: var(--gray); font-family: var(--mono); z-index: 100;
}
.key-hint {
  position: fixed; bottom: 12px; left: 20px;
  font-size: 0.72rem; color: var(--gray); z-index: 100; opacity: 0.4;
}
.notes-panel {
  position: fixed; bottom: 0; left: 0; right: 0; max-height: 35vh;
  background: var(--primary); color: #e0e0e0;
  padding: 20px 8vw; font-size: 0.92rem; line-height: 1.6;
  overflow-y: auto; transform: translateY(100%);
  transition: transform 0.25s ease; z-index: 200;
}
.notes-panel.visible { transform: translateY(0); }
.notes-panel .notes-label {
  font-size: 0.68rem; text-transform: uppercase;
  letter-spacing: 0.1em; color: var(--highlight); margin-bottom: 6px;
}
```

## SVG text

```css
svg text { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
```
