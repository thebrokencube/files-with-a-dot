# Review — Multi-Artifact Orchestration

Loaded **only** for repo/PR review targets (a single-file review skips this — it just runs the
per-artifact unit). Covers input resolution, the cross-artifact composition pass, caps, and the
output skeleton.

> Reference the type leaves by name in prose (e.g. "its `review-type-skill.md` leaf"), NOT via
> arrow-ref syntax. `dendrik lint` checks arrow references inside reference files too — so an
> arrow reference may only point at a file that exists.

## Input Resolution

| Target | Resolve | Reviewed set |
|---|---|---|
| **file** | classify the one file | single artifact (no orchestration — handled by the Review lens directly) |
| **repo / dir** | `git ls-files`, filter to the glob set below | every detected doc |
| **PR** (`#123`, URL, `--pr`) | `gh pr diff <pr> --name-only` (fallback `git diff --name-only <base>...HEAD`), filter to the glob set | changed docs + bounded neighbors |

**Glob set:** `**/SKILL.md`, `**/CLAUDE.md`, `**/AGENTS.md`, `**/ARCHITECTURE.md`,
`**/README.md`, `**/.claude/commands/*.md`, and `references/*.md` under any directory containing
a `SKILL.md`. Using `git ls-files` honors `.gitignore` and skips `.git`/vendored content for
free. Classify each hit via the detection table in the Review lens.

**Bounded neighbor read-set (no read-amplification).** In PR mode, "neighbors" = the named
same-package-root counterparts of a changed file only — `AGENTS.md`, `CLAUDE.md`,
`.cursor/rules`, `ARCHITECTURE.md`, `README.md` at the changed file's directory or repo root.
**Never a recursive repo read.** Neighbors are read for context; findings are reported against
the changed file. The output caps below cap *findings*; this rule caps *reads*.

## Per-Artifact Pass

For each doc in the reviewed set, run the per-artifact unit from the Review lens: detect type →
load `review-shared.md` (once, reused) + the one matching `review-type-*.md` leaf → produce
pass/warn/fail findings.

## Cross-Artifact Composition Pass

The Right Layer? opinion applied across files — an **opportunistic catalog**, not a mandatory
sweep. Apply a check only when its evidence is present:

| Check | Fires when | Finding (severity) |
|---|---|---|
| **CLAUDE.md hoarding** | a *project* CLAUDE.md carries portable convention/command content | move it to AGENTS.md, leave a thin pointer (warn; project-scoped — honor the claude-md leaf's path branch) |
| **Thin/absent AGENTS.md** | fat project CLAUDE.md beside a thin or missing AGENTS.md | content is in the wrong file; AGENTS.md should be source of truth (warn) |
| **Adapter drift** | AGENTS.md + an adapter (CLAUDE.md/`.cursorrules`) duplicate the same instructions | adapter should point (`@AGENTS.md`), not duplicate (warn) |
| **Skill ↔ references** | a SKILL.md + its `references/`, or a skill that fronts an external guide | arrow-refs resolve, no eager-loaded reference content, every leaf reachable; a routing skill must add discovery or behavioral value, not merely point at a doc; a thin skill fronting an *external* guide must point at it, not restate it (warn/fail) |
| **ARCHITECTURE.md inlined/orphaned** | architecture prose inlined in AGENTS.md/CLAUDE.md, or an architecture-role doc (`ARCHITECTURE.md`, `docs/architecture.md`, `*-architecture.md`) nothing points to | move inline → ARCHITECTURE.md + reference; link an orphan (best-effort — the orphan inbound-ref search is whole-repo, not guaranteed) |
| **README ↔ AGENTS.md split** | agent detail in README, or README+AGENTS.md duplication, or a front-door doc that dead-ends | agent detail belongs in AGENTS.md; a front-door doc (README/AGENTS) routes to detail, it does not dead-end (best-effort) |
| **Cross-doc contradiction** | two docs assert conflicting *verifiable* facts — a command/path/mode that exists-or-not, a "creds are/aren't committed" claim | reconcile to one authoritative home; resolve by running per the dogfood tiebreaker. Fires on conflicting *facts* only — never mere duplication (that is Adapter drift / CLAUDE.md hoarding) (warn) |
| **Doc-claims grounding** | a doc states concrete checkable claims — exact commands, file paths, env/config names, "X is/isn't committed" | verify a bounded sample (cap ~3-5 claims) against the repo with cheap exact tools — grep the command token in dispatch source, `ls` the path, grep the secret string. Literal matches only — never paraphrase, never a recursive read (touch only the targets a claim names). **fail** on an unambiguous verified-false claim (phantom command, absent path); **warn** when not run or ambiguous. Beyond the sample, recommend the repo adopt a read-only verifier (warn/fail) |

The **ARCHITECTURE.md inlined/orphaned** and **README ↔ AGENTS.md split** checks are best-effort
(they can require a whole-repo scan to fire). Project-vs-user-global scope ambiguity is resolved
per the claude-md leaf (state and ask). When a finding wants to point at a target *shape* rather
than a single fix, name the recommended doc-set pattern (its `recommended-doc-set-pattern.md`).

## Aggregation & Caps (no silent truncation)

- Per-artifact: **top 3 findings each**.
- Composition: **global top 5**.
- If more docs or findings exist than shown, state the count skipped — never imply full coverage
  when truncated.

## Output Skeleton

```
# Review — <target>
## <path/to/artifact-1>  (type: <detected>)
> finding (pass/warn/fail) …            # top 3 per artifact
## <path/to/artifact-2>  (type: <detected>)
> …
## Cross-Artifact (Right Layer?)
> composition finding …                 # global top 5
## Do First
1-2 highest-impact actions across the whole review
## Skipped
N more docs / M more findings not shown   # only if capped; omit if nothing skipped
```
