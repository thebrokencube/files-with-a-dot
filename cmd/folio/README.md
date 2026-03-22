# folio

Folio manages knowledge work projects — plans, references, and compiled outputs. The CLI handles project management (status, validation), while Claude Code skills handle the creative work (planning, writing).

## Skills (Claude Code)

Use these in Claude Code for planning and composition:

| Skill | What it does |
|-------|-------------|
| `/folio plan` | Plan non-trivial tasks by exploring options, then converging on an approach. Use for multi-file changes, architectural decisions, or unclear requirements. |
| `/folio compose` | Turn project sources (research, plans, references) into polished outputs for a target audience. |
| `/folio gather` | Add sources to a project — scaffold from URLs or do deep research on a topic. |
| `/folio publish` | Push compiled outputs to external targets (Confluence, Google Docs). |

## How It Works

```
/folio gather → design doc (lock) → impl plan → execute
        per step: implement → review (gate) → commit
        if targets: → /folio compose → /folio publish
        retro findings feed back into future work
```

`/folio plan` runs the full pipeline: gather context, freeze architecture in a design doc (mandatory lock gate), derive the implementation plan, then execute step by step with mandatory review before each commit. When the plan has external targets (Jira, branch topology), execution feeds into compose/publish. Retrospective findings loop back as observations for future cycles.

## CLI Commands

The `folio` binary handles project management:

```bash
folio status          # Show project status (sources, targets, staleness)
folio stale           # Find stale projects needing attention
folio home list       # List home-synced projects
folio home push       # Push project to home
folio validate        # Validate folio.yml structure
folio init            # Initialize a new folio project
```

## Project Structure

Each folio project has a `folio.yml` declaring its structure:

```
~/.folio/
├── active/<project>/
│   ├── folio.yml           # Project manifest (sources, targets, DAG)
│   ├── reference/          # Research, analysis, retrospectives
│   ├── work/               # Implementation plans and tracks
│   └── output/             # Composed outputs ready to share
├── archive/                # Completed/shelved projects
└── vault/                  # Cross-cutting knowledge (no folio.yml)
    ├── research/           # Tool surveys, ecosystem landscapes
    ├── domain/             # Business/technical domain knowledge
    ├── guide/              # Reusable procedures
    └── insight/            # Patterns extracted from experience
```

Projects source from the vault via `vault:` prefix paths (e.g., `vault:research/comparable-dvc.md`). Lifecycle types (spike, design, plan, retro) stay project-scoped; references promote to vault when proven cross-cutting.

Status is derived from file modification times — no separate tracking needed.
