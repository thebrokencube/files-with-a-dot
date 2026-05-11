# Lint

Read by `/folio lint [project]`. Assumes you've already read SKILL.md for orientation.

Periodic knowledge integrity pass across folio projects. Two-layer scan (CLI deterministic + LLM semantic), producing a structured cleanup plan before acting. Findings tracked in a persistent `folio-hygiene` project following standard folio lifecycle.

## Preflight

1. Check for `folio-hygiene` project: `folio home list`
2. If not found, bootstrap:
   ```
   cd ~/.folio/active
   mkdir folio-hygiene
   cd ~/.folio/active/folio-hygiene
   folio init --name "Folio Hygiene"
   ```
3. Record the folio-hygiene folio.yml path for use throughout.

## Phase 1 — Scan

Starts by loading prior state from folio-hygiene: most recent lint spike, open observations, relevant vault entries (`vault:insight/` and `vault:guide/` on folio health topics). First-ever pass has no prior state — produces full inventory.

### CLI layer (all active projects)

Run per project from `folio home list`:

```
folio health --folio <path>
folio stale --json --folio <path>
folio validate --folio <path>
folio observe list --json --folio <path>
folio observe lint --folio <path>
```

`folio health` emits human-readable text only (no JSON mode). Parse it as advisory signal.

Present results as a compact summary — project name, health grade, finding count per category.

### LLM layer (user-selected, up to 3 projects)

Use the alignment protocol (`references/alignment.md`) to steer project selection:

- minimum: 1
- categories: project selection (which projects get LLM review)
- grounding: CLI scan results from above (health grades, finding counts, staleness)
- target: the set of projects to review (up to 3)
- hard_constraints: none

Rank projects by CLI signal (high observation counts, stale targets, validation failures) and present the top candidate as a claim-first recommendation. One project per claim. Stop proposing after 3 confirmed or when the user says "enough."

For each selected project, read content and cross-reference against current state:

- **Observation triage**: Is each observation still true? Test bugs by grepping for referenced features. Check if described patterns were absorbed into docs/skills.
- **Lifecycle progression**: Spikes without follow-on (design or "concluded" signal). Designs still active that look settled. Completed work tracks missing retros. Tracks ready to archive.
- **Residency**: Vault promotion candidates (references appearing in multiple projects). Orphaned files on disk not in folio.yml (use `folio stale` output).
- **Cross-project patterns**: Observations across projects that cluster around a theme. Recurring issues visible by comparing against prior lint spikes.

Findings already reported in a prior lint spike and still unresolved are noted as recurring, not re-reported as new.

Present findings progressively — check in with the user after each project's LLM review.

## Phase 2 — Plan

### Materialize findings

Create a work track in folio-hygiene:

```
folio new --folio <folio-hygiene-path> spike lint-pass
```

Write all findings into the spike, categorized: structural, observation, lifecycle, residency, cross-project. Each finding includes evidence and recommended action.

### Produce cleanup plan

Create a design in the same work track:

```
folio new --folio <folio-hygiene-path> design lint-pass
```

Group actions by effort level:

- **Mechanical** (no judgment): Resolve confirmed-stale observations, archive completed work tracks, fix naming issues. These batch-execute with user confirmation.
- **Judgment required**: Vault promotion decisions, lifecycle progression nudges (should this spike become a design?), cross-project pattern responses. Present one at a time.
- **Deeper work**: Findings that reveal a real project need. These become observations in the target project via `folio observe` — lint surfaces the need, does not create forward work.

For large finding sets, structure execution as burndown waves per `references/burndown.md`. Use judgment on when this is appropriate, not a fixed threshold.

**Hard gate**: Present the cleanup plan and require explicit user confirmation before proceeding to execution.

### Observation cap

Cap at 20 observations per pass. If findings exceed 20, present a soft confirmation gate before continuing.

### Session handoff

The spike + design pair is the handoff artifact. If the pass spans sessions, the next session reads them to resume.

## Phase 3 — Execute

For each item in the cleanup plan:

- **Mechanical items**: Batch-resolve with `folio observe resolve`. Use `folio archive` for completed work tracks. Present batch before executing.
- **Judgment items**: Present one at a time. User decides.
- **Deeper work**: Add observation to target project via `folio observe`. This enforces type disambiguation and alignment — no direct folio.yml edits.

**Cross-project mutation gate (hard)**: Lint never directly writes to another project's folio.yml without explicit user confirmation. Present the proposed command and wait.

**Self-cleaning invariant**: As items resolve during execution, the lint work track's state reflects it. When execution is complete, the work track should have no open items.

## Phase 4 — Retro

Standard folio lifecycle retro in the folio-hygiene work track:

```
folio new --folio <folio-hygiene-path> retro lint-pass
```

Prompts:
- What patterns emerged across projects?
- Did any findings recur from a prior pass? (If so, the underlying system needs a fix, not just cleanup.)
- What heuristics proved useful vs noisy?

Extract durable insights to vault via existing types (`vault:insight/`, `vault:guide/`). These feed Phase 1 of the next pass — the flywheel.

After retro, archive the lint work track. Resolve any completed observations in folio-hygiene itself.

## Scoped Mode

`/folio lint <project>` runs lint on a single project:

- Phase 1 CLI layer runs only on the named project
- Phase 1 LLM layer auto-selects the named project (no selection step)
- Cross-project pattern detection is skipped
- Findings still materialize in folio-hygiene — same lifecycle, same self-cleaning

Useful after completing a big work track, or when a project's observation count is high.

## Guardrails

| Guardrail | Mechanism |
|-----------|-----------|
| User-steered depth | User selects up to 3 projects for LLM review |
| Cross-project mutation | Hard confirmation gate — lint proposes commands, user authorizes |
| Cleanup vs forward work | Lint resolves/archives. New design work becomes an observation in the target project. |
| Observe alignment | Cross-project observations go through `folio observe` (enforces type disambiguation) |
| Self-cleaning | Resolved items marked done during execution; work track archived after retro |
| Self-linting | folio-hygiene is included in future lint scans like any project |
| Observation cap | 20 per pass with soft confirmation gate |
