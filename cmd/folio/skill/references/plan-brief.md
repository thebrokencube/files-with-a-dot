# Plan Workflow — Brief Phase (Phases 5-6)

Read by Agent 2 (Brief). Self-contained for brief sessions.

## Phase 5: Decompose

Reads the committed design doc — no prior conversation context.

Analyze the design doc and break it into implementation tracks:

1. **Read the design doc.** This is the only input — do not rely on conversation history.
2. **Identify tracks.** Each track is an independently executable stream of work. Tracks
   should be scoped so an execution agent can pick up any single track without needing
   context from other tracks.
   Default to one track for small features (3 or fewer commits). Decompose only when there are
   genuinely independent streams or different risk profiles.
3. **Sequence by risk.** Order tracks so the highest-risk work runs first — failures surface
   early rather than after low-risk work is committed. Risk factors: architectural novelty,
   cross-file dependencies, test coverage gaps, integration surface area.
4. **Determine track dependencies.** Mark which tracks are independent (parallelizable) vs.
   sequential (each depends on the prior track's output).
5. **Present the track structure to the user.** List tracks with: name, risk assessment,
   sequencing rationale, and estimated commit count. Wait for approval before proceeding
   to Phase 6.

## Phase 6: Brief

Before writing, read 2-3 recent archived briefs from the project's work archive
(`work/archive/`) to absorb formatting conventions. Match structure, density, and tone.

Populate each track with execution-level detail. The brief must be self-contained — the
execution agent reads only the brief, not the design doc or conversation history.

Every brief has six required sections, in order:

#### Context section (required)

Distill the design doc into the minimum context an execution agent needs to make judgment
calls when implementation deviates from the plan. This replaces "go read the design doc."

Include:
- **What this work is** (2-3 sentences): the problem, the goal, what repo(s) are involved
- **Key design decisions** (bulleted, non-negotiable): architectural choices the execution
  agent must not revisit. Sourced from the design doc's Architecture and Divergence Decisions
  sections. Frame as constraints: "Do X" / "Do NOT do Y" / "X stays because Y."
- **Scope boundary** (bulleted): what is NOT included, framed as stop signals. Sourced from
  the design doc's What's NOT Included section.

Target: 10-15 lines. Dense, not conversational. Every line should change how the agent
behaves — if a line doesn't affect execution decisions, cut it.

#### Agent Setup section (required)

Tell the execution agent how to prepare before touching any code:

1. **Skill loading**: Which skills to invoke and why — `/folio status` loads folio conventions
   (key rule: `~/.folio` commits use `folio home push`, never raw git), `/commit` loads commit
   format with repo-specific conventions, plus additional skills as needed.
2. **Repo mapping**: Which tracks operate in which repos. Execution agents work in one repo
   at a time — make the mapping unambiguous.
3. **Escalation triggers**: When should the agent stop and ask the user? Common triggers:
   validation failures after migration steps, file paths or signatures that don't match the
   brief, temptation to add code not in the track, high-complexity commits.

#### Tracks section (required)

For each track, specify:
- Exact file changes (create, modify, delete) with paths
- Function signatures, struct diffs, type definitions
- Test names and what they validate
- Commit message(s) and what each commit contains
- Validation commands (build, test, lint)

#### Build & Deploy section (required)

How to build, test, and deploy in the target repo. An execution agent should be able to
validate its work without guessing. Include:
- **Build commands**: exact commands to compile/build (with working directory)
- **Test commands**: how to run the full test suite and targeted tests
- **Deploy/sync steps**: what to run after code changes land (e.g., `dot sync` for
  dotfiles, `make build` for checked-in binaries, deploy commands for services)
- **Prerequisites**: tools to install, dependencies to set up, environment to configure

This section prevents the most common handoff failure: the execution agent makes correct
code changes but doesn't know how to validate or ship them.

#### Execution Conventions section (required)

Commit format, scope target (max commits — typically ~5), ordered commit sequence
(what goes in each commit, in what order), push workflow, and repo-specific patterns.

Include a **Folio integration** subsection: targets to add for branches, existing observation
items to resolve on completion, `folio home push` checkpoints at milestones. Do not instruct
agents to add "completion" observations — observations are open items, not a changelog.
Execution agents should maintain folio state as they go — not as a final cleanup step.

#### Handoff Prompts section (required)

Copy-pasteable prompts for starting execution sessions. One prompt per independent
session (often one per track, but coupled tracks share a session).

Each prompt must:
- Point to the committed brief as the primary input
- Name the specific track(s) to execute
- Include build/test/deploy commands inline (don't just say "see brief")
- Include a subagent review step before committing (see plan-execute.md Phase 7 step 4)
- Include folio checkpoint instructions (observations to resolve, retro trigger)

The handoff prompt is the acid test of brief quality: if the prompt needs to add context
beyond "read the brief and execute Track N," the brief is underspecified.

Scaffold the brief under `work/active/YYYY-MM-DD-<topic>/README.md`. If the brief needs
per-track detail files (large tracks), create `track-N.md` siblings.

### Brief Verification Gate (hard)

Before committing, ALL prerequisites must be true. No exceptions — briefs with stale
references cause mid-execution corrections.

1. [ ] File paths verified — launch verification agents to confirm all referenced paths exist
2. [ ] Function signatures verified — agents confirm signatures match actual code
3. [ ] Tag/version sequences verified — `git tag --sort=-v:refname | head -1` matches brief
4. [ ] User approved track structure (Phase 5 gate passed)

Fix any inaccuracies, then commit via `folio home push`. The committed work brief is the
contract for Agent 3.

## Session Exit (mandatory)

1. **Retro prompt**: "Anything worth retroing on before we move to execution?"
   Materialize findings via `folio new retro <topic>` and observation items. Commit via
   `folio home push`. For lightweight retros, observation items alone suffice.
2. **Handoff gate** — two options:
   - **Continue** (default): Spawn next agent via Agent tool with fresh context. Pass only
     the committed artifact path and setup instructions — no conversation history.
   - **New session**: Provide a paste-able prompt for the user to start fresh.
   Format: "Work brief committed at [path]. **Continue to Execute phase, or hand off to a
   new session?**"
3. **Clipboard delivery** (mandatory for new-session handoff): Write the handoff prompt to a
   temp file and `pbcopy < /tmp/handoff-prompt.txt`. The prompt exists in the doc for
   durability, but clipboard is how the user actually starts the next session.
