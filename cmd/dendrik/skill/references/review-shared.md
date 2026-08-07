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

**Dogfood tiebreaker:** when a finding hinges on a *verifiable* fact (a command/path/mode that
exists or not), the runnable source decides — resolve by running, not debating.

## Universal Dimensions (every type)

### Link Integrity
- All file paths, arrow refs, and markdown links resolve to existing files.
- Cross-references are accurate and current. No dead references.
- **fail** on any broken link; **warn** on a link that resolves but points at the wrong target.

### Writing Quality
- Imperative/infinitive form (verb-first), not second person ("Run X", not "you should run X").
- Tables and lists for structured information; prose only when flow matters.
- Concise — no filler, no restated context.
- Modal ladder — an instruction uses `must`, `can`, or `will`. A requirement written as `should` reads as optional. A `should` describing behavior is fine, so judge which one it is.
- Condition before command — "If the build fails, read the log", never the reverse.
- One noun per concept — `config` here and `settings` there is a defect, not a stylistic wobble.
- Internal consistency — an edit matches its host section's cadence (no multi-clause bullet in a terse list, no lone heading mid-prose, no outsized example).
- Edit regressions (diff only) — a list collapsed into prose (list→paragraph regression), or an edited section whose register/format drifts from untouched sibling sections *in the same file*. When no diff is available (plain single-file review), this axis does not fire — it is never a repo-wide "you have a paragraph" nag.
- **pass:** clean imperative style, structured data in tables. **warn:** occasional
  second-person/passive but clear, an edit that breaks its section's established cadence/concision, or a
  single modal-ladder or condition-order slip. **fail:** predominantly second-person, prose walls where
  tables belong, or a `should` carrying a requirement the reader must follow.

-> Read references/writing-replace-table.md when a finding's fix is a word swap.


### Instruction Economy

What an instruction costs the model — distinct from how it reads (Writing Quality) and where it sits
(Right Layer?). Grounded in Anthropic's
[Opus 5 prompting guide](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/prompting-claude-opus-5).

- **Instructs behavior the model already has** — "double-check", "re-verify", "use a subagent to verify".
  These cause *over*-verification; removing them costs no quality. A domain check the model would not
  invent ("loop the suite over every commit, not the tip") is not this and stays.
- **Deny-list with no target stated** — negatives nudge unwanted tokens down, a positive statement of the
  wanted shape pulls harder. Three to five hard constraints beside a target is the working shape.
- **Mandated fan-out** — small or serial work routed through subagents.
- **Budget pressure** — an always-loaded file where a new line makes the existing lines less likely to be
  followed. Per line: would removing this cause a mistake?
- **fail** on an instruction to self-verify. **warn** on the other three.

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
always-loaded file; an adapter that duplicates the recipe instead of pointing at it; volatile
specifics — exact counts, per-layer numbers, enumerated ID lists — restated in an always-read doc
when code or a single reference already owns them (describe the shape, point at the one source).

**Recipe-vs-adapter worked example:** A "review a PR"
workflow — isolate a worktree → run linters → dispatch a reviewer subagent → report → cleanup —
placed entirely in a `Claude-specific` reference file is **mislabeled**. Only the reviewer
*dispatch* and the *isolation mechanism* are harness-bound; the rest is a portable recipe and
belongs in the agnostic layer (AGENTS.md / shared docs), with a thin adapter naming just the two
bindings. Flag portable logic siloed in a harness-specific file, and the same workflow
duplicated across the agnostic and harness layers (drift risk).

## Output Format

Reviews are **layered** so they read and share at three depths: a verdict line (the
Slack-forward sentence), a TL;DR (the one-glance shape), then risk-grouped findings the reader
opens on demand. Do not emit a flat list of equal-weight blocks.

```
**Verdict: <✅ Approve | 🟡 Approve with fixes | 🔴 Request changes> — <one-line why>.**

**▸ TL;DR** — <one-line quality read: what's strong>. Biggest issue: <the single top finding>.
`N fix-first · M nits · K strengths`

### 🔧 Fix first (N)
**🟡 [Area] — [brief verdict]**
> Specific observation. Concrete suggestion.

### ⚪ Nits (M)
**⚪ [Area] — [brief verdict]**
> Specific observation. Concrete suggestion.

### 👍 What's good (K)
- <strength> · <strength>
```

**Verdict maps off risk** (highest tier present wins):
- any 🔴 fail → **Request changes**
- only 🟡 warn / ⚪ nit → **Approve with fixes**
- all pass / clean → **Approve**

**Tiers** map from pass/warn/fail: **Fix first** = fails + high-impact warns; **Nits** =
low-impact warns and optional cleanups. Each finding keeps its 🔴/🟡/⚪ tag so risk stays visible
within a group.

**Render targets:**
- **Chat / terminal** (default) — use `###` group headers, as above.
- **GitHub PR comment** — swap each `###` group header for `<details><summary><b>…</b></summary>`
  so the groups collapse; keep the verdict + TL;DR outside any `<details>` so they're always visible.

## Presentation Rules

- Maximum 5 findings per artifact (excluding strengths), ranked by impact (for multi-artifact
  reviews, see the caps in `review-orchestration.md`).
- Lead with the verdict, then the TL;DR's single biggest issue.
- The "Fix first" group *is* the "do these 1-2 things first" — keep it to the highest-impact items.
- If the artifact is strong, say so — the "What's good" group is required, not optional filler.
- Use pass/warn/fail vocabulary (and the matching 🔴/🟡/⚪ tags) consistently.
