# Reference

## folio.yml Schema

### Required Fields

```yaml
schema: 2          # Schema version (currently 2)
project: "Name"    # Human-readable project name
```

### Sources

Sources declare what feeds into the project:

```yaml
sources:
  # Local file
  - path: reference/research/landscape.md

  # Vault reference (resolves to ~/.folio/vault/)
  - path: vault:research/comparable-dvc.md

  # External system
  - external: jira
    id: "PROJ-123"

  - external: google_docs
    id: "1TaUG..."

  # Derived source (provenance tracking)
  - path: reference/insight/patterns.md
    derived_from: reference/research/landscape.md
```

### Targets

Targets declare what the project produces:

```yaml
targets:
  tech-spec:
    how: "Extract overview and plan into a standalone tech spec."
    sources:
      - path: work/active/feature/README.md
    outputs:
      - path: output/tech-spec.md
      - external: google_docs
        id: "1TaUG..."
        field: body
    blocked_by:
      - upstream-target    # DAG dependency
```

Target variants:
- **Simple**: single source, single output
- **Batch**: multiple items sharing one `how` instruction
- **Forest**: `forest:` block delegates to jf for Jira hierarchy management

### Repositories

```yaml
repositories:
  - name: my-repo
    url: "https://github.com/org/repo"
```

URL templates for code links referenced in sources.

### Cross-References

```yaml
cross_references:
  - fact: "The migration deadline is Q2 2026"
    source_of_truth: reference/domain/timeline.md
    also_appears_in:
      - work/active/migration/README.md
      - output/tech-spec.md§Timeline
```

Track where facts appear to prevent drift. The `§` separator targets a specific section within a file.

### Observations

```yaml
observations:
  - "idea(cli): support batch operations"
  - "bug(home): push blocked by unrelated lint failure"
  - "RESOLVED idea(docs): add getting-started guide"
```

Open-items queue managed through `folio observe` commands.

### Schema 2 Changes

Schema 2 replaced `pending` and `tasks` sections with `observations`. Design documents colocate in work directories (`work/active/<topic>/reference/design/`). Reference labels map to vault-eligible types.

## Artifact Types

| Type | Path | Purpose | Vault? |
|------|------|---------|--------|
| spike | `reference/spike/` | Time-boxed investigation | No |
| design | `work/active/<topic>/reference/design/` | Architecture decision | No |
| plan/brief | `work/active/<topic>/README.md` | Execution plan with tracks | No |
| retro | `reference/retro/` | Retrospective findings | No |
| research | `reference/research/` | Landscape scan, survey | Yes |
| domain | `reference/domain/` | Business/technical knowledge | Yes |
| guide | `reference/guide/` | Reusable procedures | Yes |
| insight | `reference/insight/` | Patterns from experience | Yes |
| review | `reference/review/` | Code/design review findings | Yes |

`folio new design <topic>` auto-creates a work directory with colocated design doc. Other types scaffold to `reference/<type>/`.

## CLI Command Reference

| Command | Usage | Description |
|---------|-------|-------------|
| `folio init` | `folio init --name "Name"` | Initialize new project with folio.yml |
| `folio status` | `folio status [--json] [--folio PATH]` | Show project status (sources, targets, staleness) |
| `folio validate` | `folio validate [--folio PATH]` | Validate folio.yml structure |
| `folio health` | `folio health [--folio PATH]` | Project health report |
| `folio stale` | `folio stale` | Find stale projects needing attention |
| `folio new` | `folio new <type> <topic> [--folio PATH]` | Scaffold a typed artifact |
| `folio observe` | `folio observe '<type>(scope): desc'` | Add an observation |
| `folio observe list` | `folio observe list [--json]` | List all observations |
| `folio observe resolve` | `folio observe resolve <#N\|substring>` | Resolve an observation |
| `folio observe types` | `folio observe types` | Show valid observation types |
| `folio observe lint` | `folio observe lint` | Validate observation format |
| `folio gather` | `folio gather <url> [--materialize --type T]` | Scaffold source entry from URL |
| `folio dag` | `folio dag [--branches]` | Show project dependency DAG |
| `folio touch` | `folio touch [--folio PATH]` | Clear staleness after manual publish |
| `folio archive` | `folio archive [--dry-run]` | Move project to archive |
| `folio pbcopy` | `folio pbcopy <path>` | Copy file to clipboard |
| `folio home list` | `folio home list` | List home-synced projects |
| `folio home push` | `folio home push` | Commit and push to ~/.folio remote |
| `folio home pull` | `folio home pull` | Pull from ~/.folio remote |
| `folio setup` | `folio setup [--check]` | Setup or diagnose environment |

**Flag ordering**: flags must come before positional arguments (enforced by the CLI parser).

## Observation Format

Format: `type(scope): description`

### Valid Types

| Type | When to use |
|------|-------------|
| `idea` | Feature ideas, improvements |
| `task` | Concrete work items |
| `bug` | Known defects |
| `gap` | Missing capability or documentation |
| `debt` | Technical debt to address |

### Scope Conventions

Scope is freeform but common values include: `cli`, `skill`, `agent`, `docs`, `health`, `process`, `roadmap`.

### Lint Rules

`folio observe lint` checks:
- Format matches `type(scope): description`
- Type is one of the valid types
- Resolved observations have the `RESOLVED` prefix
- No duplicate observations

## Status Derivation

Folio determines status from file modification times -- no database.

- A target is **stale** when any source file's mtime is newer than the output file's mtime
- `folio touch` updates output mtimes to clear staleness after manual publish
- `folio status --json` returns structured output with per-target staleness

## Troubleshooting

`folio setup --check` diagnoses the environment (binary location, home directory, git state).

Common issues:

- **Missing folio.yml**: run `folio init` or use `--folio PATH` to point to an existing one
- **Invalid schema**: check `folio validate` output for field-level errors
- **Push blocked**: `folio home push` runs lint on all active projects -- fix the reported project's issues first
- **Archive dry-run mutates state**: known bug in `folio archive --dry-run` -- avoid until fixed
