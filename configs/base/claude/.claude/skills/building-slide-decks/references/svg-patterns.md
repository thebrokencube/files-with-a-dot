# SVG Diagram Patterns

Reusable patterns for common diagram types. All SVGs use the deck's CSS custom properties via `var()`.

## General rules

- **ViewBox padding**: Add 10-20px beyond content bounds. Text clipping at edges is the #1 SVG bug.
- **Font**: SVG text inherits from `svg text` CSS rule (system font stack). Don't set font-family inline.
- **Colors**: Use `var(--primary)`, `var(--highlight)`, etc. They work in SVG `fill` and `stroke`.
- **Responsive**: Set `style="width:100%; max-width:NNNpx; display:block;"` on the `<svg>`.
- **Dark-bg slides**: Use lighter colors — `rgba(255,255,255,0.08)` for subtle fills, `#D4D8E0` / `#A8B0BE` for text.

## Arrow markers

Define once in `<defs>`, reuse with `marker-end="url(#arrID)"`.

```html
<defs>
  <!-- Right-pointing arrow -->
  <marker id="arrR" markerWidth="6" markerHeight="8" refX="6" refY="4" orient="auto">
    <polygon points="0 0, 6 4, 0 8" fill="#5BA3A2"/>
  </marker>
  <!-- Down-pointing arrow -->
  <marker id="arrD" markerWidth="8" markerHeight="6" refX="4" refY="6" orient="auto">
    <polygon points="0 0, 8 0, 4 6" fill="#5BA3A2"/>
  </marker>
</defs>
```

## Flow diagram (left to right)

Pill-shaped stages connected by arrows. Good for workflows.

```html
<svg viewBox="0 0 600 50" style="width:100%; max-width:600px; display:block;">
  <rect x="0" y="8" width="120" height="34" rx="17" fill="rgba(var-primary,0.2)" stroke="var(--primary-light)" stroke-width="1"/>
  <text x="60" y="30" text-anchor="middle" fill="#D4E0E0" font-size="11" font-weight="600">Step 1</text>

  <line x1="124" y1="25" x2="148" y2="25" stroke="var(--primary-light)" stroke-width="1.5" marker-end="url(#arrR)"/>

  <rect x="152" y="8" width="120" height="34" rx="17" fill="rgba(var-primary,0.2)" stroke="var(--primary-light)" stroke-width="1"/>
  <text x="212" y="30" text-anchor="middle" fill="#D4E0E0" font-size="11" font-weight="600">Step 2</text>
  <!-- repeat pattern -->
</svg>
```

## Architecture diagram (layered)

Top-down layers showing orchestration relationships. Used in the RADR deck for /write-radr → CLI → artifacts.

Pattern:
1. **Top layer**: Primary entry point (pill, prominent fill)
2. **"orchestrates" label + arrow**
3. **Middle layer**: Dashed border container with internal pipeline stages (pills connected by arrows)
4. **"produces" label + arrow**
5. **Bottom layer**: Artifact cards (rectangles with colored top-bar accent)
6. **Side element**: Automation / on-push action (dashed connection)

Key technique — container with dashed border groups related elements:
```html
<rect x="30" y="96" width="640" height="66" rx="10"
      fill="rgba(42,157,143,0.08)" stroke="var(--primary-light)"
      stroke-width="1" stroke-dasharray="4,3"/>
```

## Comparison table (CSS grid, not SVG)

For side-by-side comparisons, use CSS grid instead of SVG:

```html
<div style="display: grid; grid-template-columns: auto 1fr 1fr;
     border: 1px solid var(--light-gray); border-radius: 10px;
     overflow: hidden; font-size: clamp(0.82rem, 1.1vw, 0.92rem);">
  <!-- Header row -->
  <div style="padding: 12px 20px; background: var(--primary);"></div>
  <div style="padding: 12px 20px; background: var(--primary); color: #F0F5F5; font-weight: 700; text-align: center;">Option A</div>
  <div style="padding: 12px 20px; background: var(--highlight); color: #F5F5F5; font-weight: 700; text-align: center;">Option B</div>
  <!-- Data rows -->
  <div style="padding: 10px 20px; border-bottom: 1px solid var(--light-gray); font-weight: 600; color: var(--primary);">Row label</div>
  <div style="padding: 10px 20px; border-bottom: 1px solid var(--light-gray); color: var(--dark-gray);">Value A</div>
  <div style="padding: 10px 20px; border-bottom: 1px solid var(--light-gray); color: var(--dark-gray);">Value B</div>
</div>
```

## Converging arrows

Two elements flowing into a single point (merge pattern):

```html
<path d="M195,44 L195,70 Q195,82 207,82 L300,82" fill="none" stroke="#5BA3A2" stroke-width="1.5"/>
<path d="M505,44 L505,70 Q505,82 493,82 L400,82" fill="none" stroke="#5BA3A2" stroke-width="1.5"/>
<circle cx="350" cy="82" r="4" fill="#5BA3A2"/>
```

## Artifact card (with top-bar accent)

Document-style element with a colored top bar:

```html
<rect x="30" y="200" width="190" height="56" rx="5" fill="rgba(255,255,255,0.06)"/>
<rect x="30" y="200" width="190" height="4" rx="2" fill="var(--primary-light)" fill-opacity="0.5"/>
<text x="125" y="226" text-anchor="middle" fill="#E8E8E8" font-size="11" font-weight="700">filename.ext</text>
<text x="125" y="244" text-anchor="middle" fill="#7BBBBA" font-size="9">description</text>
```
