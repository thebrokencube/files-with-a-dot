# Code Comments

**One comment per change, and name it.** That is the budget. When you show the diff, say which line you
spent it on and what it protects. Spending nothing is the ordinary case and needs no announcement; spending
two needs a reason.

Zero was the old rule and it did not hold. Abstinence reads as withholding, so it gets rationalised around.
A budget engages the same instinct instead: one is scarce, so it goes somewhere that earns it.

## The urge is a discard, not a comment

Nearly every comment that feels necessary is really *I considered this and it mattered*. Say that in chat,
on one line, at the end of the turn:

```
considered and dropped: X, Y, Z
```

The conversation is disposable and sits in front of the reader. The file is permanent and read by strangers
who did not ask. Verbose in chat, terse in the repo — not the reverse.

## Nowhere is a real destination

A cut comment does **not** move to the PR body. The body says what changed, not what you thought about.
Relocating narration is the same act with a different filename, and the more expensive one, because more
people read it.

## The bar for the one you spend

One line, carrying the single fact not recoverable from the code:

- an external API's surprising behaviour
- a constraint a future edit would silently break
- why the obvious alternative is wrong

## Placement

On the line it protects, never as a preamble above the method. A header block has no single line to defend,
so it grows into an essay about the whole design. If the fact applies to no particular line, it is not a
comment.

## Never

- **Narrate.** If it restates what the next line does, delete it.
- **Dump context.** The decision's history, the options weighed, the measurements taken — nowhere.
- **Argue with a reviewer.** `Deliberately NOT …`, `Not inside X because …`, `Revisit if …`. These address
  a critic rather than a reader.
- **Explain a name.** A method called `redact` needs no comment saying it redacts.
- **Banner.** No `# ── internals ──` dividers.

**Tests spend nothing.** A test name states the case and an assertion states the expectation. There is no
budget in a test file.

## When called out

Do not answer with a metric — that is `concision.md` rule 7. Do not answer by moving the text elsewhere.
Delete it, and if it was a discard, put it on the discard line.

**Why:** volume is how effort gets signalled, so prohibiting the output does not touch the driver — it
resurfaces one layer down, in the PR body, a docstring, a longer reply. The discard line gives that signal
a cheap home; the budget makes the expensive one scarce. Related:
[writing-structure.md](writing-structure.md).
