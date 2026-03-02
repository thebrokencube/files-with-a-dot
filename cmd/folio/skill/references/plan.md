# Plan Workflow

Read by `/folio plan [topic]`. Multi-agent diverge-converge planning for non-trivial tasks.

## When to Use

Use `/folio plan` instead of `EnterPlanMode` for any non-trivial task — multi-file changes, architectural decisions, unclear requirements, or multiple valid approaches.

**Skip planning entirely** for trivial changes: single-file fixes, obvious bugs, one-line tweaks. Just do them.

## Workflow

```
  gather         design            plan             execute
    ↓               ↓                ↓                 ↓
 sources ──→ design doc ──→ approved plan ──→ implement per step
(Phase 1)   (4b) LOCK      (Phase 5)        ┌─────────────────┐
    ↑                                        │ implement       │
    │                                        │ validate   GATE │
    │                                        │ review     GATE │
    │                                        │ commit          │
    │                                        └────────┬────────┘
    │                                                 │
    │                                          ┌──────┴───────┐
    │                                          │ compose →    │
    │                                          │   publish    │
    │                                          │ (if targets) │
    │                                          └──────┬───────┘
    │                                                 │
    └──── retro findings ←─── experience ←────────────┘
          (pending items)     (Phase 7)
```

The forward path: gather sources, freeze architecture in a design doc (Phase 4b — mandatory lock), derive the implementation plan from the frozen design (Phase 5), execute step by step (Phase 6).

Phase 6 is a loop per plan step: implement → validate → review (mandatory gate) → commit. No commit without review.

When the plan has external targets (Jira hierarchy, branch topology), execution feeds into compose/publish to sync outputs. This is optional — not every plan has external targets.

The feedback loop: implementation experience feeds the Phase 7 retrospective, whose actionable findings become pending items that inform future cycles. Design docs persist as durable reference material, not disposable intermediates.

## Invocation

- `/folio plan` — Infer topic from context, default lenses (pragmatic vs thorough)
- `/folio plan <topic>` — Explicit topic, default lenses

Custom lenses are specified naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). The agent parses the user's intent and crafts lens descriptions accordingly. Default lenses (pragmatic vs thorough) apply when no lens guidance is given.

## Phases

Phases 1-4 run in normal mode (full tool access), including the design lock at 4b. Phase 5 enters built-in plan mode for structured approval. Phase 6 implements with mandatory pre-commit review gates. Phase 7 runs the post-implementation retrospective.

### Phase 1: Understand

Gather context before spawning any agents. This happens in the main conversation.

1. Identify the task scope from the user's request (or `<topic>` argument)
2. Read relevant files — entry points, existing implementations, tests, CLAUDE.md
3. Search for related patterns in the codebase (Grep/Glob)
4. **Check for folio context**: Run `folio home list` to find active projects. If any project is relevant to the task, read its `folio.yml` and pull in relevant sources, cross-references, and tasks/pending items.
5. **Present folio findings to the user.** If a project matched: summarize project name, key
   sources pulled in, relevant pending tasks. Wait for confirmation before continuing. If no
   project matched: list the active projects that were considered, ask if any of them are
   relevant (the user may see a match the search missed), and ask if a new folio project
   should be created to track this work. Only proceed without folio context after explicit
   user confirmation.
5b. **Check source freshness — MUST run if folio project matched.** For every source with a
   `derived_from` entry, compare the `cached` date against today. If any source is >14 days
   stale, STOP and present the list to the user: "These sources may be outdated: [list with
   ages]. Refresh before planning?" Wait for the user's response. If yes, use the gather
   workflow (see `references/gather.md`) per stale source. If no, note staleness in the
   context summary so lenses can account for it. Do not auto-run gather. Do not proceed to
   step 6 until the user has acknowledged or dismissed the staleness report.
6. **Pin hard constraints**: Separate the user's stated decisions and explicit preferences
   (hard constraints) from open trade-offs. Hard constraints are non-negotiable — lenses
   must not re-evaluate them. Include pinned constraints as a distinct section at the top
   of the context summary.
7. Compile a **context summary** (max 30 lines): pinned hard constraints first, then what
   exists, what needs to change, key trade-offs, and relevant folio context (if any)

This summary is passed to all downstream agents. Diversity comes from lenses, not information asymmetry.

**Materialization gate**: If Phase 1 research produces substantial findings (not just reading
existing files), materialize them as spike files via `folio new spike <topic>` before Phase 2
begins. Commit via `folio home push`. The spikes become sources that propose agents can reference.

### Phase 2: Propose

Launch 2 Plan agents in parallel, each with the same context summary but a different lens. Each returns a proposal (max 80 lines).

Default lenses:
- **Pragmatic**: Minimize changes, reuse existing code, prefer the simplest approach that works
- **Thorough**: Consider edge cases, maintainability, architectural fit, future extensibility

### Phase 3: Converge

Launch 1 agent (subagent_type: general-purpose) to merge the two proposals into a single plan (max 100 lines).

Convergence criteria:
- Every file to be changed is listed with what changes and why
- Implementation order is specified
- Trade-offs between the two proposals are noted where they diverged meaningfully
- If both proposals agreed on an approach, that's a strong signal — keep it
- The merged plan must be an executable spec: function signatures, struct definitions, and
  edge-case handling are pre-decided — not left to the implementer

After the converge agent returns, briefly summarize (3–5 lines) the key divergence decisions to the user — which proposal won on each point and why. Informational only, not blocking. Proceed to Phase 4 immediately after.

### Phase 4: Review

Launch 1 agent (subagent_type: general-purpose) to review the merged plan. The review covers three checks:

1. **Accuracy**: Does the plan reference correct file paths, function names, and APIs? Are assumptions about existing code valid?
2. **Feasibility**: Can each step actually be implemented as described? Are there missing dependencies or ordering issues?
3. **Scope**: Is everything in the plan necessary for the task? **Meta-review: should any of this work not exist?** Flag anything that's over-engineered, gold-plated, or solving a problem the user didn't ask about.
4. **Commit decomposition**: Does the plan decompose into clean commits? Each step should map to 1 logical commit. Flag steps that bundle unrelated concerns (refactor + feature, deps + implementation). Cross-reference against the scope target in execution conventions. See the commit skill's "Reviewable History" section.
5. **Version impact**: If the target repo uses tagged-repo versioning (per commit skill), does the plan account for version bumps? Flag if feat/fix/refactor mix within a single step implies ambiguous version impact.

Review output: max 40 lines. For each issue found, state: what's wrong, where, and a suggested fix.

### Phase 4b: Design Lock

**Mandatory — no skip.** Freezes architectural decisions before implementation begins. Every plan produces a design doc — the doc may be lightweight for simple changes, but it always exists. Do NOT enter plan mode (Phase 5) without a committed design doc.

**Steps:**
1. Scaffold a design doc via `folio new design <topic>`. This creates the file at
   `reference/design/YYYY-MM-DD-<topic>.md` with the design template and registers it in
   folio.yml. Fill in the template from converge output:
   - **Problem**: What and why
   - **Architecture**: Key decisions, type definitions, function signatures
   - **Divergence decisions**: Table from converge phase
   - **What's NOT included**: Explicit scope boundary
   - **Design Provenance**: Agent count, lens names, review findings
2. If a folio project exists: commit via `folio home push`
3. If no folio project: use `--no-register` and write to the plan file's directory instead
4. Present to user: "Design doc committed. Proceeding to implementation plan."
5. **Lock**: Phase 5 derives the implementation plan from the committed design doc, not raw converge output. If implementation later contradicts the design doc, stop and consult the user.

### Phase 5: Present (enter plan mode)

After Phases 1-4b complete, hand off to built-in plan mode for structured approval:

1. Call `EnterPlanMode`. The user will be prompted to consent to entering plan mode.
2. Once in plan mode, write the final plan to the **plan file path provided in the plan mode system message**. Include:
   - The plan (with any review fixes applied)
   - A summary of what the review flagged and how it was addressed
   - A brief note on where the two proposals differed and which was chosen
   - **The implementation order must end with an explicit step for Phase 7 (retrospective).** This is not implicit — it must appear as a numbered step so it cannot be skipped during execution.
   - **Execution conventions** (required section): Implementation idioms locked before
     execution begins. Required fields: commit format, scope target (max commits —
     typically ~5), validation commands (build, test, lint), module/package path, push
     workflow, and repo-specific patterns discovered in Phase 1.
3. Call `ExitPlanMode` with `allowedPrompts` populated from the plan — e.g., if the plan includes running tests, include `{"tool": "Bash", "prompt": "run tests"}`. This presents the plan to the user for approval.

The user sees the plan file cleanly. They may request changes, ask questions, or reject. Iterate until approved or abandoned.

**Note**: Full conversation context (Phases 1-4) is retained in plan mode — only tool access is restricted (read-only + no Agent). This is fine since the heavy lifting is done.

### Phase 6: Implement

Only after user approval (ExitPlanMode accepted). If you discover something unexpected during implementation that contradicts the plan, stop and consult the user rather than improvising.

For each plan step, execute this sequence in order. Do NOT skip or reorder steps.

1. **Implement** one logical unit (one feature, fix, or refactor). If a plan step spans
   multiple concerns, split it at commit time.
2. **Content extraction check** (if this step moved or extracted content across files): Diff
   old content against new locations. Verify nothing was dropped, duplicated, or silently
   truncated. Do not rely solely on review agents — run an explicit before/after comparison.
3. **Validate — hard gate.** Run the validation commands from execution conventions in order:
   build, then test, then lint. If ANY command fails, STOP and fix. Do NOT proceed to step 4
   until all validation commands pass.
4. **Review — hard gate.** Launch 2 review agents (subagent_type: general-purpose) — one
   checking accuracy, one checking scope. Converge findings and fix issues. Do NOT run
   `git commit` until both reviews complete and all findings are resolved. If fixes are
   mechanical (typos, imports, paths), proceed to step 5. If fixes change logic or add code
   paths, return to step 3.
5. **Commit.** One logical unit per commit.
6. **Repeat** from step 1 for the next plan step.

**Folio integration**: If a relevant folio project exists, record design decisions, progress, and rationale in the folio project as work progresses — not as a final cleanup step. This means updating folio.yml tasks/pending, adding reference files for significant decisions, and keeping cross-references current throughout implementation. All `~/.folio` commits must use `folio home push` (see SKILL.md § Git Operations).

### Phase 7: Retrospective

**Mandatory — no skip.** After all implementation commits are complete, review the planning
process in the main conversation (do NOT delegate to a subagent — the retrospective needs
full session context to be useful). Cover:

- What worked well? What added friction?
- Were the lenses useful? Did convergence surface good trade-offs?
- Was the folio context helpful (if used)?
- Was the design doc useful? Did it prevent architectural drift during implementation, or was it overhead for this task?
- What should change next time?
- If the plan changed folio source files, flag whether targets need recompilation.

Only capture actionable findings, not session notes. Findings that aren't worth planning
aren't actionable — note them in the retro summary and move on.

**Recording findings**: Retro findings worth preserving MUST be materialized via
`folio new retro <topic>` — not just added as pending items. The retro file captures the full
context; pending items capture only the actionable follow-ups. Both are written: the retro file
is the durable artifact, pending items are the work triggers. Commit via `folio home push`.

For lightweight retros (few findings, simple context), pending items alone are sufficient — use
judgment on whether the full context warrants a retro file.

## Re-run Rule

If plan feedback from the user requires rethinking (not just minor edits), re-run phases 2-4 (full diverge-converge). Do not patch the existing plan.

## Agent Prompts

### Propose Agent (Phase 2)

Use with `subagent_type: "Plan"`. Launch two instances in parallel with different lens values.

```
You are planning an implementation for: {task_description}

## Context
{context_summary}

## Your Lens: {lens_name}
{lens_description}

## Instructions
Propose an implementation plan. Your plan should:
- List every file to create or modify, with a description of the changes
- Specify implementation order and dependencies between steps
- Note any risks, assumptions, or open questions
- Stay within your lens — let it guide your trade-off decisions

Keep your proposal under 80 lines. Be concrete — file paths, function names, specific changes. No hand-waving.
```

**Default lens descriptions:**

- **Pragmatic**: "Minimize the number of files changed and lines of code written. Reuse existing patterns and utilities. Prefer the simplest correct solution. Avoid new abstractions unless they pay for themselves immediately."
- **Thorough**: "Consider edge cases, error handling, and how this change interacts with the rest of the codebase. Prioritize maintainability and architectural consistency. Flag potential issues even if fixing them adds scope."

### Converge Agent (Phase 3)

Use with `subagent_type: "general-purpose"`.

```
You are merging two implementation proposals into a single plan.

## Original Task
{task_description}

## Proposal A (Pragmatic)
{proposal_a}

## Proposal B (Thorough)
{proposal_b}

## Instructions
Merge the two proposals into a single implementation plan:
- For each aspect, pick the stronger approach — or synthesize from both
- Note where the proposals diverged and which was chosen, with rationale
- If both agreed on an approach, that's a strong signal — keep it
- Every file to change must be listed with what changes and why
- Specify implementation order
- Pre-decide function signatures, type definitions, and edge-case handling where feasible
- Include an "Execution conventions" section with commit format, validation commands, and scope target

Keep the merged plan under 100 lines. Be concrete.
```

### Review Agent (Phase 4)

Use with `subagent_type: "general-purpose"`. Needs file access to verify claims.

```
You are reviewing an implementation plan. Your job is to find problems before code is written.

## Original Task
{task_description}

## Plan to Review
{plan_content}

## Review Checklist
1. **Accuracy**: Verify file paths, function names, and API references exist and are correct. Read the actual files.
2. **Feasibility**: Can each step be implemented as described? Are there missing imports, wrong method signatures, or ordering issues?
3. **Scope**: Is everything necessary? Meta-review: should any part of this plan NOT exist? Flag over-engineering, unnecessary abstractions, or work the user didn't ask for.
4. **Commit decomposition**: Does each step map to 1 logical commit? Flag steps that bundle unrelated concerns. Cross-reference against the scope target in execution conventions.
5. **Version impact**: If the repo uses tagged versioning, does the plan account for version bumps? Flag ambiguous version impact from mixed change types in a single step.

For each issue found, state: what's wrong, where in the plan, and a concrete fix.
Keep your review under 40 lines. Only flag real issues — don't nitpick style or add suggestions beyond the task scope.
```

### Implementation Review Agent (Phase 6)

Use with `subagent_type: "general-purpose"`. Launch two instances — one with accuracy focus, one with scope focus. Needs file access.

```
You are reviewing code changes before a commit. Your job is to catch implementation bugs.

## Original Task
{task_description}

## Approved Plan
{plan_summary}

## Changes to Review
{change_description}

## Your Focus: {focus}
{focus_description}

For each issue found, state: what's wrong, where, and a concrete fix.
Keep your review under 40 lines. Only flag real issues.
```

**Focus descriptions:**

- **Accuracy**: "Verify the changes match the approved plan. Check for typos, stale references, incorrect file paths, broken imports, and wrong function signatures. Read the actual changed files."
- **Scope**: "Check that changes are necessary and sufficient. Flag anything that wasn't in the plan, unnecessary abstractions, or missing pieces. Meta-review: should any of these changes NOT exist? Verify the commit bundles one logical unit — flag if unrelated concerns are mixed."

## Notes

- **Interaction with folio**: Phase 1 checks for active folio projects. If one is relevant, its sources and cross-references inform the context summary.
- **Phase 4b (Design Lock)**: Always runs. Every plan produces a design doc before entering plan mode. The doc is the authoritative source for Phase 5.
- **Custom lenses**: Users can specify lenses naturally in the topic text (e.g., `/folio plan redesign auth, considering performance and readability`). Parse the user's intent and craft lens descriptions accordingly.
- **Retrospective findings**: Phase 7 captures findings as pending items in the folio project, not as recursive `/folio plan` invocations.
