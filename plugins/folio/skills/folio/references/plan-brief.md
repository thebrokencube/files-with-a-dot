# Plan Workflow — Brief Phase (Phases 5-6)

Read by Agent 2 (Brief). Also read by `/folio dispatch`: the transient dispatch projection and
pre-dispatch gate live at the end of the Execution Setup section. A dispatch reader uses only the
committed work-plan README, not the design doc.

## Phase 5: Decompose

For Phase 5, read the committed design doc — no prior conversation context.

Analyze the design doc and break it into implementation tracks:

1. **Read the design doc.** For Phase 5, this is the only input — do not rely on conversation history.
2. **Identify tracks.** Each track is an independently executable stream of work. Tracks
   should be scoped so an execution agent can pick up any single track without needing
   context from other tracks.
   Default to one track for small features (3 or fewer commits). Decompose only when there are
   genuinely independent streams or different risk profiles.
   Each track title is phrased as its future Jira issue title (`type(scope): description`) with the
   frozen design title as parent — a **soft convention**: carry the design's `scope` forward; edit
   only on genuine scope change; a track may become an epic or a story by size. Cross-cutting
   commits (`refactor:`/`chore(deps):`) keep their own scope per `/commit`.
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
- **What this work is** (2-3 sentences): the problem, the bounded task sentence and goal, what repo(s) are involved
- **Non-negotiable decisions** (bulleted): constraints from the design doc's Non-Negotiable
  Constraints and Chosen Approach sections. Frame as: "Do X" / "Do NOT do Y" / "X stays
  because Y."
- **Scope boundary** (bulleted): stop signals from the design doc's Scope Boundary section.

Target: 10-20 lines. Dense, not conversational. Every line should change how the agent
behaves — if a line doesn't affect execution decisions, cut it.
For a transient dispatch projection, `outcome` is the goal clause of **What this work is**, the
bounded `task` is its bounded task sentence, accepted decisions are the **Non-negotiable decisions**
verbatim, and scope boundary is the **Scope boundary** bullets. The projector may not broaden that task.

#### Interface Spec section (required)

Cross-boundary contracts that execution agents must respect. Sourced from the design doc's
Interfaces section.

Include:
- **File change manifest**: path, action (create/modify/delete), what changes
- **Cross-boundary type definitions** (code blocks): from design doc's Component Contracts
- **Validation commands** (pass criteria): build, test, lint (exact commands with working directory)
For dispatch, this section also supplies the file manifest and registered source locator/role table.
Read each locator at projection time from canonical `folio.yml`; do not copy manifest identity or
freshness metadata into the brief. The committed README is the required semantic subject. The design
doc and sketch are optional parent-session comparison references, not fresh-consumer inputs.

| Registered reference | Locator from canonical `folio.yml` | Role | Fresh-consumer input |
|---|---|---|---|
| committed README | registered project-relative README path | required semantic subject | yes |
| design doc | registered project-relative design path | optional parent-session comparison | no |
| sketch | registered project-relative sketch path | optional parent-session comparison | no |

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
For a transient dispatch projection, this section supplies predicted slices, dependencies, and the
explicitly selected track. `next_artifact.job` is the first output of the first unfinished future
track after the explicit current-track request; the request names the current selected work. Resolve
that future track at capture, record the job in the view, and never make the receiver recompute it.
In the local dogfood fixture, the explicit request is `track=T2` (Track 2); the view records Track
3's `plan-brief.md` change as `next_artifact.job`.

A predicted slice is one track-level unit; each `after` value is an explicitly declared predecessor
slice label, never a manifest source path. Use `after` only for those labels in this view; it is not
folio.yml's validated `depends_on` source-path key. For multi-repository briefs, produce one view per
explicitly selected track/repository. An unselected target is unresolved and refuses dispatch; never
invent a target, task, next-artifact job, approval, or provider field. Preserve
`repository: <one repo mapping from Execution Setup, or unresolved>`. An unresolved repository is
emitted as the refusal record and is never dispatched.

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
For dispatch, this section supplies acceptance criteria, evidence references, test infrastructure, and
the not-tested boundary. Evidence references are the exact commands listed under **Acceptance
criteria** and their observed output or status; a transport receipt is not test proof.

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
For a dispatch projection, this section supplies the selected repository/track target, skill
invocations, commit/push workflow, escalation triggers, and Folio integration.

**Transient dispatch projection (guidance, not a storage schema).** Assemble the view from the five
sections above; do not add a schema key or serializer. Include `repository`, the explicitly selected
`track`, `outcome`, bounded `task`, accepted decisions, scope boundary, interface manifest and
registered locator/role table, predicted slices with `after` labels, `next_artifact.job`, `derived_from`,
test strategy, selected skills, workflow, escalation, Folio integration, and a `dispatch_key`.

For store-local evidence other than the brief locator, use the live jj change ID. The `dispatch_key`
always uses the brief's landed (`::main`) change ID. For external sources, use the canonical `id` or
`url`; include `derived_from.cached` only when that source is registered. Unavailable identity or
freshness is reported, not a refusal, when the source is readable. A registered stale fact or missing
or unreadable required context refuses dispatch; missing or unreadable optional context is reported
as unavailable.

The `brief locator` is the project-relative README path in its canonical `folio.yml` source entry,
not an absolute path or store identifier. Define `dispatch_key` as
`<registered README path>@<latest landed change ID>`. For example, the local dogfood key is
`work/active/2026-09-12-sketch-stage-rethink/README.md@wnzvmyqnvyzyqvzlyouznnmrvtzvprxn`.

`<isolated-workspace>` is the `FOLIO_HOME` root returned by `folio home workspace list` or created by
`folio home workspace create`; its root contains `active/`, and it is the cwd for the following
read-only lookup. First confirm the registered path exists at `main`:

```bash
jj file list -r main -- active/tooling/folio/work/active/2026-09-12-sketch-stage-rethink/README.md
```

Then resolve the latest landed change ID:

```bash
jj log --no-graph -r '::main & files("active/tooling/folio/work/active/2026-09-12-sketch-stage-rethink/README.md")' -n 1 -T 'change_id ++ "\n"'
```

For another brief, replace the `<project-path>` and `<date>-<topic>` segments while retaining the
`active/<project-path>/work/active/<date>-<topic>/` shape (`tooling/folio` and
`2026-09-12-sketch-stage-rethink` here); nothing else varies. Empty log output means not landed only
after the registered path check succeeds; a path mismatch is a stop condition, not evidence of
non-landing.
“Landed” means the brief was committed through `folio home push` onto `main`; unrelated store-head
movement does not invalidate the key.

**Pre-dispatch gate.** Before any provider invocation, walk through these gates: accepted Folio
planning gates; a landed, committed brief; an explicit selected target; resolvable required context;
no unresolved authority-affecting choice; and a named receiver. If `repository` is unresolved, emit the view as a
refusal record and never dispatch it. The receiver gets only the derived projection plus explicitly
named references and the selected target — never the conversation transcript or implicit parent
context. A receipt or transport success is not completion; terminal evidence must be explicit and
remains provider-owned. Capability negotiation, refusal vocabulary, lifecycle, workspace,
authentication, retries, callbacks, and terminal artifacts remain provider-owned.

Changed brief content or changed declared source freshness invalidates the view and requires the
planning gates to be revisited. Unrelated store-head movement does not. Missing freshness or identity
remains nonblocking when the source itself is readable.

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
10. [ ] Titles carry the frozen design `type(scope)` (track/commit titles); cross-cutting commits excepted
11. [ ] CLI ops verified — every CLI command/subcommand/flag the brief references actually exists (read source or `--help`; don't assume `--help` works)

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
     the committed work-plan path and skill invocations — no conversation history.
   - **New session**: Tell the user *"Work plan committed at `<path>`. Start the next
     session with `/folio <project>` — it'll surface the work plan and ready-state."* No
     paste-able prompt, no tempfile. The committed work plan IS the handoff.

   Format the choice: *"Work plan committed at `<path>`. **Continue to Execute phase, or
   hand off to a new session?**"*
