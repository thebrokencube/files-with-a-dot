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

**The 5-section structure is mandatory.** Every work plan MUST have all five sections listed
below, in order. Archived briefs that predate this structure are NOT precedent — this spec
supersedes them.

**Section transition from previous 6-section format:**

| Old Section (previous 6) | New Home (current 5) | Notes |
|--------------------------|----------------------|-------|
| Context | → Direction Summary | Compressed, 10-20 lines |
| Agent Setup | → Execution Setup | Skill loading, repo mapping |
| Tracks | → Track Decomposition | + deferral markers, wave grouping |
| Build & Deploy | → Execution Setup | Validation commands subsection |
| Execution Conventions | → Execution Setup | Commit conventions, folio integration |
| Handoff Prompts | Eliminated | Templatable from conventions; no longer mandatory |

Interface Spec is new — sourced from the design doc's Interfaces section, not remapped from
any prior brief section.

Test Strategy is new — sourced from the design doc's Testing Strategy section. Hard gate:
execution cannot begin without this section present and user-approved.

**Progressive disclosure structure.** Briefs follow the two-layer layout defined in
`references/progressive-disclosure.md`: action layer (track instructions, top) and reference
layer (Key Reference sections, bottom). The action layer should be self-sufficient for
mechanical steps. If a track step can't be executed without a Key Reference, the track
instruction is underspecified — add more inline detail.

Every brief has five required sections, in order:

#### Direction Summary section (required)

Distill the design doc's Direction section into the minimum context an execution agent needs
to make judgment calls when implementation deviates from the plan. This replaces "go read
the design doc."

Include:
- **What this work is** (2-3 sentences): the problem, the goal, what repo(s) are involved
- **Non-negotiable decisions** (bulleted): constraints from the design doc's Non-Negotiable
  Constraints and Chosen Approach sections. Frame as: "Do X" / "Do NOT do Y" / "X stays
  because Y."
- **Scope boundary** (bulleted): stop signals from the design doc's Scope Boundary section.

Target: 10-20 lines. Dense, not conversational. Every line should change how the agent
behaves — if a line doesn't affect execution decisions, cut it.

#### Interface Spec section (required)

Cross-boundary contracts that execution agents must respect. Sourced from the design doc's
Interfaces section.

Include:
- **File change manifest**: path, action (create/modify/delete), what changes
- **Cross-boundary type definitions** (code blocks): from design doc's Component Contracts
- **Validation commands** (pass criteria): build, test, lint (exact commands with working directory)

#### Track Decomposition section (required)

**Multi-track plans MUST create track-N.md files from the start.** Do not defer to "if
needed" — track files prevent README.md from growing unwieldy and signal clear boundaries
to execution agents. Single-track plans keep everything in README.md.

For each track, specify:
- Exact file changes (create, modify, delete) with paths
- Function signatures, struct diffs, type definitions
- Test names and what they validate; test helper signatures and shared fixtures
- Commit message(s) and what each commit contains
- Deferral markers: **`[PRE-DECIDED]`** (execution agent follows literally) vs
  **`[RESOLVE IN SPIKE]`** (execution agent makes the call during Step 0 spike)
- Flag creative vs mechanical steps with **"judgment required"** markers — helps execution
  agents know when to follow literally vs when to adapt
- Wave grouping (if multiple tracks): group independent tracks into waves. Tracks within a
  wave run in parallel; waves are sequential. Format: `### Wave 1 (parallel)` /
  `### Wave 2 (after Wave 1)`. Single-track plans skip wave notation.

**Specification depth must be approximately uniform across tracks.** If Track 1 has function
signatures and test tables, Track 4 cannot be a one-liner. Thin tracks signal the brief
author didn't think them through — flesh them out or merge them into an adjacent track.

#### Test Strategy section (required — hard gate)

Concrete testing plan sourced from the design doc's Testing Strategy section. This section
is a **hard gate**: present the test strategy to the user for explicit approval before
proceeding. Execution cannot begin without sign-off.

Include:
- **Test types** (bulleted): which types apply (unit, integration, e2e, manual) and why
  each is needed. Not a generic list — tied to specific tracks or components.
- **Acceptance criteria**: concrete conditions that define "passing." Exact commands where
  possible (e.g., `rspec spec/models/foo_spec.rb`, `yarn test --filter bar`).
- **Test infrastructure**: fixtures, factories, test data, external services, or mocks
  required. "None beyond existing" is valid when true.
- **Not tested**: what is explicitly out of scope for testing, with rationale. Omitting this
  signals the author didn't think about coverage boundaries.

Target: 10-20 lines. Every line should map to something an execution agent verifies.
Vague entries like "add appropriate tests" fail the gate — specify what tests, for what
behavior, using what approach.

**Gate behavior**: After writing this section, present it to the user and wait for explicit
approval. If the user requests changes, revise and re-present. Do not proceed to Execution
Setup until approved.

#### Execution Setup section (required)

Consolidates agent setup, build/deploy, and execution conventions into a single section.

Include:
- **Repo mapping**: which tracks operate in which repos
- **Skill loading**: which skills to invoke and why — `/folio status` loads folio conventions
  (key rule: `~/.folio` commits use `folio home push`, never raw git), `/commit` loads commit
  format with repo-specific conventions, plus additional skills as needed
- **Commit conventions**: format, scope target (max commits — typically ~5), ordered commit
  sequence (what goes in each commit, in what order), push workflow, repo-specific patterns
- **Validation commands** (run sequence): build, test, lint, deploy/sync steps (e.g.,
  `dot sync` for dotfiles, `make build` for checked-in binaries, `dot validate` after
  any skill or symlink deletion)
- **Escalation triggers**: when should the agent stop and ask? Common triggers: validation
  failures, file paths or signatures that don't match the brief, temptation to add code not
  in the track, high-complexity commits
- **Folio integration**: existing observation items to resolve
  on completion, `folio home push` checkpoints at milestones. Do not instruct agents to add
  "completion" observations — observations are open items, not a changelog. Execution agents
  should maintain folio state as they go — not as a final cleanup step.

**Handoff Prompts** are no longer a mandatory section. They are templatable from the
Execution Setup conventions. Include them only when the execution session needs context
that isn't obvious from "read the brief and execute Track N."

Scaffold the work plan under `work/active/YYYY-MM-DD-<topic>/README.md`. If a work dir
already exists for the topic (from a prior `folio new design`), use it — the design doc
will be a sibling at `reference/design/` inside the same directory. If the plan needs
per-track detail files (large tracks), create `track-N.md` siblings.

### Commit Checklist (ALL must pass before `folio home push`)

The commit instruction includes the gate — they are inseparable. Do not commit without
mentally processing every item.

1. [ ] File paths verified — launch verification agents to confirm all referenced paths exist
2. [ ] Function signatures verified — agents confirm signatures match actual code
3. [ ] Tag/version sequences verified — `git tag --sort=-v:refname | head -1` matches brief
4. [ ] External constraints verified — CI config, team names, templates referenced in the
   brief exist and are current
5. [ ] No unresolved choices — grep the brief for "or", "either", "optionally" and resolve
   each to a concrete decision
6. [ ] Specification depth approximately uniform across tracks
7. [ ] User approved track structure (Phase 5 gate passed)
8. [ ] Test strategy approved by user (Test Strategy hard gate passed)
9. [ ] Content review passed (see below)

Fix any inaccuracies, then commit via `folio home push`. The committed work plan is the
contract for Agent 3.

### Content Review (part of commit checklist, item 9)

Dispatch 1 review agent (subagent_type: general-purpose, model: "sonnet") to review:
1. **Self-containment**: Can an execution agent proceed without reading the design doc?
2. **Convention compliance**: Does structure match recent archived briefs? All 5 sections present?
3. **Completeness**: Are all design doc decisions reflected as constraints?
4. **Adversarial check**: Challenge the brief's own conclusions — is the track decomposition
   the right one? Could fewer tracks achieve the same result? Are any tracks solving
   problems the design doc didn't ask for? (See `references/adversarial-review.md`.)

For multi-track plans, use multi-persona review (3-4 agents with different perspectives:
accuracy, scope, completeness, executability). Single-track plans use 1 agent.

If issues found, fix and re-review. Loop until clean (cap 5 iterations).

## Session Exit (mandatory)

1. **Retro prompt**: "Anything worth retroing on before we move to execution?"
   Materialize findings via `folio new retro <topic>` and observation items. Commit via
   `folio home push`. For lightweight retros, observation items alone suffice.
2. **Handoff validation (mandatory)**: Spawn a fresh subagent with ONLY the committed work
   plan path and design doc path (no conversation context). The subagent reads both and
   reports ambiguities — anything unclear, contradictory, or requiring context the plan
   doesn't provide. Fix ambiguities before proceeding. This catches gaps that the author
   can't see because they have conversation context the execution agent won't.
3. **Handoff gate** — two options:
   - **Continue** (default): Spawn next agent via Agent tool with fresh context. Pass only
     the committed artifact path, setup instructions, and skill invocations — no conversation
     history.
   - **New session**: Provide a paste-able prompt for the user to start fresh. Include skill
     invocations (e.g., `/folio status`, `/commit`) so the new session loads the right context.
   Format: "Work plan committed at [path]. **Continue to Execute phase, or hand off to a
   new session?**"
4. **Clipboard delivery** (mandatory for new-session handoff): Write the handoff prompt to a
   temp file and `pbcopy < /tmp/handoff-prompt.txt`. The prompt exists in the doc for
   durability, but clipboard is how the user actually starts the next session.
