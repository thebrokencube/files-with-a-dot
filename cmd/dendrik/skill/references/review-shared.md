# Review — Shared Core

Universal dimensions, the cross-cutting Right Layer? check, scoring, and output rules. Loaded
on **every** review (single-file or multi-artifact), alongside exactly one `review-type-*.md`
leaf for the detected type. The leaf adds the type-local dimensions; this file is what they all
share.

## How to Score

Evaluate each applicable dimension (the ones below + the leaf's) as **pass / warn / fail**:

- **pass** — meets the bar; no action needed.
- **warn** — functional but with a real weakness worth fixing; or an opinion-level finding
  (see Right Layer?), never a hard failure.
- **fail** — a concrete defect that breaks the artifact's job (dead links, missing required
  structure, misroutes).

Surface the top findings ranked by impact. End with 1-2 concrete "do these first" items. If
everything passes, say so — do not manufacture findings.

## Universal Dimensions (every type)

### Link Integrity
- All file paths, arrow refs, and markdown links resolve to existing files.
- Cross-references are accurate and current. No dead references.
- **fail** on any broken link; **warn** on a link that resolves but points at the wrong target.

### Writing Quality
- Imperative/infinitive form (verb-first), not second person ("Run X", not "you should run X").
- Tables and lists for structured information; prose only when flow matters.
- Concise — no filler, no restated context.
- **pass:** clean imperative style, structured data in tables. **warn:** occasional
  second-person/passive but clear. **fail:** predominantly second-person, prose walls where
  tables belong.

## Right Layer? (cross-cutting — every type)

The one question applied to every agentic document: **is this content in the layer that matches
its load model and audience?** Each agentic doc sits somewhere on a recipe → adapter → reference
spectrum, and content in the wrong layer costs tokens, drifts, or misses its reader.

**Default-target rule:** agent-behavioral / project-convention content (build & test commands,
code-org rules, conventions, constraints) belongs in **AGENTS.md** — the portable source of
truth. Tool-specific files (**CLAUDE.md**, `.cursorrules`, `.github/copilot-instructions.md`)
are thin **adapters** that should mostly *point at* AGENTS.md (`@AGENTS.md` / symlink), not
contain it. Reference docs (ARCHITECTURE.md, deep guides) are **referenced**, not inlined into
always-loaded files.

This is dendrik's opinion, acknowledged as community practice — not Anthropic-published
behavior (Claude Code still treats CLAUDE.md as its native memory file). So findings here **name
the opinion** ("dendrik recommends…"); they do not assert a spec violation, and the
AGENTS.md-push is **warn at most**, never fail. The per-type leaves name the type-local
expression of this question; this section is the authoritative source so it survives new types.

**Flags:** project conventions hoarded in a project CLAUDE.md instead of AGENTS.md; architecture
prose inlined in AGENTS.md or CLAUDE.md instead of referenced; reference knowledge baked into an
always-loaded file; an adapter that duplicates the recipe instead of pointing at it.

### Volatile specifics / denormalization (sub-check)

High-level / always-read docs (README, CLAUDE.md, SKILL.md, AGENTS.md-class) carry
**direction / architecture / approach** — durable content that doesn't drift. They must **not**
carry **volatile specifics**: exact counts, per-layer numbers, enumerated lists of IDs, or any
value hand-copied from somewhere it's already defined. Those belong in **one derived source**
(the code itself, or a single generated/asserted reference) — never denormalized across prose.
Direction doesn't drift; "30 vs 31" does, and the same number copied into seven surfaces becomes
seven things to keep in sync.

**Flags:** an exact count or enumerated list in an always-read doc that restates what a code
file or a single reference already owns; the same volatile value appearing in more than one
high-level doc (drift factory). The fix is **de-denormalization** — describe the *shape*
qualitatively in the high-level doc and point at the one derived source — not propagating the
"correct" value to every copy. **warn** (an opinion-level layering finding); name it as
volatile-specifics denormalization. The automated form of this check (what counts as a volatile
specific, per-type thresholds) is a deferred future effort, not part of `dendrik lint` today.

**Recipe-vs-adapter worked example:** A "review a PR"
workflow — isolate a worktree → run linters → dispatch a reviewer subagent → report → cleanup —
placed entirely in a `Claude-specific` reference file is **mislabeled**. Only the reviewer
*dispatch* and the *isolation mechanism* are harness-bound; the rest is a portable recipe and
belongs in the agnostic layer (AGENTS.md / shared docs), with a thin adapter naming just the two
bindings. Flag portable logic siloed in a harness-specific file, and the same workflow
duplicated across the agnostic and harness layers (drift risk).

## Output Format

Present each finding as:

```
**[Area] — [pass/warn/fail] [brief verdict]**
> Specific observation. Concrete suggestion.
```

## Presentation Rules

- Maximum 5 findings per artifact, ranked by impact (for multi-artifact reviews, see the caps
  in `review-orchestration.md`).
- Lead with the highest-impact issue.
- End with "do these 1-2 things first" — specific, actionable next steps.
- If the artifact is strong, say so. Do not manufacture findings.
- Use pass/warn/fail vocabulary consistently.
