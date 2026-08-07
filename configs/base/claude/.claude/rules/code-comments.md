# Code Comments

Default to zero. A comment earns its place only where a reader would otherwise do the wrong thing.

## When called out

Do not answer the critique with a metric — the general rule is `concision.md` rule 7. Placement and shape
are the axis here, not density or block length. A number showing "within baseline" is a defence rather than
an answer, and self-chosen comparables make it a worthless one. Re-read the comments, find the narration,
cut it.

## The bar

One line, carrying the one fact not recoverable from the code:

- an external API's surprising behaviour
- a constraint a future edit would silently break
- why the obvious alternative is wrong

Everything else goes in the PR body, or nowhere.

## Placement

The comment sits **on the line it protects**, not above the method as a preamble. A header block over a
method is where narration collects: it has no single line to defend, so it grows into an essay about the
whole design. If the fact applies to one call, one constant, or one branch, put it there. If it applies to
no particular line, it is context, and context goes in the PR body.

## Never

- **Narrate.** If it restates what the next line does, delete it.
- **Dump context.** The decision's history, the options weighed, the measurements taken — PR body.
- **Argue with a reviewer.** `Deliberately NOT …`, `Not inside X because …`, `Revisit if …`. These address
  a critic rather than a reader, and they duplicate the PR body that already makes the case.
- **Explain a name.** A method called `redact` needs no comment saying it redacts.
- **Banner.** No `# ── internals ──` dividers.

Tests and frontend default hardest to zero. A self-explanatory test name needs no comment, and an assertion
needs no inline gloss.

**Why:** agent-authored comments run long and argumentative, and read as design-doc prose parked in the
source. Diffs here are reviewed closely and must read as human-written.

**How to apply:** write the code with no comments at all. Then add back only lines that prevent a specific
wrong edit, put each on the line it protects, and trim each to one sentence before showing the diff.
Related: [writing-structure.md](writing-structure.md).
