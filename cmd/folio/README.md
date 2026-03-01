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
~/.folio/active/<project>/
├── folio.yml           # Project manifest (sources, targets, DAG)
├── plans/              # Implementation plans
├── reference/          # Research, analysis, retrospectives
└── compiled/           # Composed outputs ready to share
```

Status is derived from file modification times — no separate tracking needed.
