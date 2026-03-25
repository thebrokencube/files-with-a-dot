# Getting Started with Folio

## What is Folio?

Folio is a CLI and skill toolkit that manages knowledge work projects -- plans, references, and compiled outputs -- with a YAML manifest (`folio.yml`) at the center. Sources (markdown files, Jira tickets, Google Docs) compose into targets (tech specs, ticket descriptions, published documents) through a declared pipeline.

Folio is not a note-taking app or task tracker. It's a compilation system: local markdown sources feed into external targets, and status is derived from file modification times rather than a database.

## The Lifecycle Model

Knowledge work in folio follows a progression:

```
observation -> spike -> design -> plan[tracks] -> implementation -> retro
     ^                                                                |
     '---------------------- findings feed back ----------------------'
```

A concrete example: you notice that symlink handling is fragile (observation). You investigate how other dotfile managers solve it (spike). You decide on an approach with tradeoffs documented (design). You break implementation into tracks with dependencies (plan). You execute track by track with review gates (implementation). You capture what worked and what didn't (retro), and findings feed back as observations for future cycles.

**Lifecycle types** progress through these stages: observation, spike, design, plan, track, retro. They always stay project-scoped.

**Reference labels** (research, insight, guide, domain, review) feed in at any stage. They describe the nature of a file rather than its position in a lifecycle.

| From | To | Trigger |
|------|------|---------|
| observation | spike | "I should investigate this" |
| spike | design | Findings warrant a solution |
| design | plan | Design approved, ready to execute |
| plan | implementation | Tracks defined, dependencies clear |
| implementation | retro | Work complete or shelved |
| retro | observation | Findings surface new questions |

## Project Anatomy

A folio project is a directory with a `folio.yml` manifest:

```
~/.folio/active/my-project/
├── folio.yml              # The manifest -- everything starts here
├── reference/
│   ├── research/          # Landscape scans, tool surveys
│   ├── design/            # Architecture decisions (legacy location)
│   ├── spike/             # Time-boxed investigations
│   └── retro/             # What did we learn?
├── work/
│   ├── active/            # In-progress implementation plans
│   │   └── my-feature/
│   │       ├── README.md  # The plan brief
│   │       └── reference/
│   │           └── design/  # Colocated design docs
│   └── archive/           # Completed or shelved work
└── output/                # Composed artifacts ready to share
```

### folio.yml

The manifest declares sources, targets, cross-references, and observations. Here's a simplified real example:

```yaml
schema: 2
project: "App Benefits Structure"

sources:
  - path: reference/guide/benefits-conventions.md
  - path: reference/domain/benefits-ownership.md
  - external: jira
    id: "BEN-47474"

targets:
  tech-spec:
    how: "Extract overview and plan into a standalone tech spec."
    sources:
      - path: work/active/app-benefits-structure/.jf/README.md
    outputs:
      - path: output/tech-spec.md
      - external: google_docs
        id: "1TaUG..."
        field: body
```

Sources feed into targets. Targets can depend on other targets via `blocked_by`, forming a DAG (directed acyclic graph). `folio dag` visualizes this dependency chain.

## Two-Tier Residency: Projects and Vault

Projects live under `~/.folio/active/<name>/` with their own `folio.yml`.

The vault (`~/.folio/vault/`) is a shared knowledge layer with no `folio.yml` -- its directory structure is its index:

```
~/.folio/vault/
├── research/    # Tool surveys, ecosystem landscapes
├── domain/      # Business/technical domain knowledge
├── guide/       # Reusable procedures
└── insight/     # Patterns extracted from experience
```

Projects reference vault files with the `vault:` prefix in source paths (e.g., `vault:research/comparable-dvc.md`), which resolves to `~/.folio/vault/`.

**When to promote to vault**: a reference proves useful across multiple projects. A tool survey used by 3 different projects belongs in the vault. A project-specific investigation stays as a spike.

## Quick Start

Prerequisites: dotfiles installed (`dot health` passes), `folio` on PATH.

```bash
# Create a new project
folio init --name "My Project"

# Add your first observation
folio observe 'idea(scope): what I noticed'

# Scaffold a spike for investigation
folio new spike my-topic

# Check project state
folio status

# Push to ~/.folio home
folio home push
```

For planning and composition workflows, use `/folio plan` and `/folio compose` in Claude Code. See [02-workflows.md](02-workflows.md) for details.

## Key Concepts

| Term | Definition |
|------|-----------|
| observation | A thing that needs attention -- idea, bug, gap, debt, or task |
| spike | Time-boxed investigation with a clear question |
| design | Architecture decision document with tradeoffs |
| plan/brief | Work plan with tracks and dependencies |
| track | A unit of work within a plan |
| retro | Retrospective -- what worked, what didn't |
| target | A declared output destination (file, Jira, Google Doc) |
| source | Input material declared in folio.yml |
| vault | Shared cross-project knowledge layer at `~/.folio/vault/` |
| compose | Turn local sources into communication artifacts |
| publish | Push composed output to an external system |
| DAG | Directed acyclic graph of target dependencies |
