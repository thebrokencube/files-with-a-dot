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

  # Vault reference (resolves from the owning store or selected content root)
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
| `folio status` | `folio status [--json] [--folio PATH]` | Show target freshness and lifecycle state |
| `folio validate` | `folio validate [--folio PATH]` | Validate folio.yml structure |
| `folio health` | `folio health [--folio PATH]` | Project health report |
| `folio stale` | `folio stale [--folio PATH]` | Find stale or missing local outputs; explain unknown targets without failing |
| `folio new` | `folio new <type> <topic> [--folio PATH]` | Scaffold a typed artifact |
| `folio observe` | `folio observe '<type>(scope): desc' [--folio PATH]` | Add an observation |
| `folio observe list` | `folio observe list [--json] [--folio PATH]` | List all observations |
| `folio observe resolve` | `folio observe resolve <#N\|substring> [--folio PATH]` | Resolve an observation |
| `folio observe types` | `folio observe types` | Show valid observation types |
| `folio observe lint` | `folio observe lint [--folio PATH]` | Validate observation format |
| `folio gather` | `folio gather <url> [--materialize --type T] [--folio PATH]` | Scaffold source entry from URL |
| `folio dag` | `folio dag [--branches] [--folio PATH]` | Show project DAG |
| `folio touch` | `folio touch [--final] [--folio PATH] TARGET` | Record input digests without rewriting outputs; `--final` marks a reviewed terminal target |
| `folio archive` | `folio archive [--dry-run] [--folio PATH]` | Move a work track to archive after dependent preflight |
| `folio home archive` | `folio home archive <path>` | Move an active project to archive with dependent preflight |
| `folio home activate` | `folio home activate <path>` | Restore an archived project after dependent preflight |
| `folio home push` | `folio home push [<store>]` | Commit and push the selected Folio store |
| `folio home pull` | `folio home pull [<store>]` | Pull the selected Folio store |
| `folio setup` | `folio setup [--check]` | Setup or diagnose environment |

**Flag ordering**: flags must come before positional arguments (enforced by the CLI parser).
## Archive and activation safety

Project and work-track archive operations resolve structured path references from sibling projects in the selected Folio store before moving anything. If a sibling reference resolves inside the candidate root, the command refuses before rename, manifest write, sync, or push. Folio does not rewrite dependent manifests automatically; update the dependent project manually, validate it, and retry.


After a permitted move, Folio validates the projected manifest. A validation failure rolls the directory back to its original location, preserves manifest bytes, and removes only parent directories created by the attempted move. A rollback warning means the operation failed and the affected paths require manual inspection.

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

Folio derives local output freshness from recorded input digests when a complete
composition snapshot exists.

- `clean`: every declared local input matches the recorded digest
- `stale`: local input content changed after composition
- `missing`: a required local output is absent
- `unknown`: no complete snapshot exists, input freshness is delegated to an external
  system, or the target is a forest managed by `jf`

External outputs retain `status: "unknown"` without a cause and do not affect a target's
local aggregation. Unknown is advisory: `folio stale` exits zero for unknown-only
results, while stale or missing local outputs return a nonzero exit.

`folio touch TARGET` records the complete input digest and composition timestamp without
rewriting declared outputs. The fields are written as a pair. `folio touch --final TARGET`
requires a direct external output and an existing local review copy, and rejects batch and
forest targets. A final target does not receive propagated upstream status, but its own
stale or missing local outputs remain in `folio stale`.

## Troubleshooting

`folio setup --check` diagnoses the environment (binary location, selected control/content roots, and VCS state).

Common issues:

- **Missing folio.yml**: run `folio init` or use `--folio PATH` to point to an existing one
- **Invalid schema**: check `folio validate` output for field-level errors
- **Push blocked**: `folio home push` runs lint on all active projects -- fix the reported project's issues first
