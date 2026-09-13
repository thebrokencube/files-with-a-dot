# Workflows

## The Knowledge Graduation Path

Most folio work follows a progression from raw input to published artifact:

```
URL/idea -> source entry -> reference file -> composed output -> published artifact
```

Each step maps to a folio command or skill:

| Step | Tool | Example |
|------|------|---------|
| Capture a URL | `folio gather <url>` | Scaffold a source entry from a Confluence page |
| Materialize content | `folio gather --materialize` | Download and create a local reference file |
| Full research | `/folio gather <topic>` | Agent-driven research with synthesis |
| Compose output | `/folio compose` | Turn sources into a tech spec |
| Publish externally | `/folio publish` | Push to Google Docs or Jira |

## Starting a New Project

```bash
folio init --name "My Project"
folio observe 'idea(cli): support batch operations'
folio observe 'gap(docs): no getting-started guide'
```

**When to create a new project vs extend an existing one**: if the work has its own deliverables and timeline, it's a new project. If it extends an existing effort, add it as a work track.

**When to split**: the files-with-a-dot project (this dotfiles repo's own folio project) grew to 600+ lines in folio.yml tracking 38 work tracks. At that scale, consider splitting related work into separate projects that cross-reference each other.

## Research: Gathering Sources

Two shapes of research:

- **Snapshot** (new topic): `folio gather <url>` scaffolds a source entry. Add `--materialize` to download content and create a local reference file. Use `--type research` to specify the reference type.
- **Re-seed** (update existing): the `folio gather <topic>` Agent Skill workflow does full agent-driven research — surveys the landscape, synthesizes findings, and updates existing references.

**Vault vs project-scoped**: landscape scans and tool surveys that apply across projects go in the selected Folio store's `<work-root>/vault/research/`. Project-specific investigations stay as spikes within the project.

## Planning: From Observation to Design

**When to plan**: multi-file changes, unclear requirements, architectural decisions, work that spans multiple systems.

**When NOT to plan**: single-file fixes, obvious bugs, trivial additions. Just do it.

The planning pipeline:

1. **Gather** context from sources
2. **Design** doc -- freeze architecture (mandatory lock gate)
3. **Work plan** -- break into tracks with dependencies
4. **Execute** track by track with review gates

The `folio plan` Agent Skill runs this full pipeline. The diverge-converge model generates multiple independent proposals from different perspectives (e.g., pragmatic vs thorough), then synthesizes them into one approach.

**Lightweight mode**: 5 or fewer files with clear scope get a combined design+brief instead of the full pipeline.

## Composing Outputs

Compose turns local sources into communication artifacts for a target audience. The `how` field in folio.yml is the composition instruction.

Three target shapes:

| Shape | Description | Example |
|-------|-------------|---------|
| **Simple** | One source, one output file | A spike summary becomes a tech spec |
| **Batch** | Multiple items sharing one `how` | 4 reference files become 4 Google Doc tabs |
| **Forest** | jf-managed Jira hierarchy | An epic's child tickets get descriptions from local markdown |

DAG ordering matters: compose upstream targets first. `folio dag` shows the dependency chain.

Use the `folio compose` Agent Skill to run composition. The skill reads the `how` field, gathers declared sources, and produces the output.

## Publishing to External Systems

Publish sends composed output to Jira, Google Docs, Slack, or Confluence.

Push methods vary by target:
- **Jira**: always through `jf` (the Jira forest CLI -- never direct API calls)
- **Google Docs**: via the host's configured gdrive tools
- **Manual**: copy output to clipboard with `pbcopy < output-file`

There is a mandatory review gate before every external push. After publishing, verify
the external result. If the target has a local output, run `folio touch <target>` to
record the input digest without changing the output. Use `--final` only for a
non-batch, non-forest target with a direct external output and an existing local review
copy; final targets stop dependency propagation but still surface local stale or missing state.

## Observations as an Open-Items Queue

Observations are not a task tracker -- they're things that need attention. They serve as the intake funnel for future work.

```bash
folio observe 'idea(cli): support batch operations'
folio observe 'bug(home): push blocked by unrelated lint failure'
folio observe 'gap(docs): no architecture doc'
folio observe list              # see all open observations
folio observe resolve '#3'      # mark as addressed
folio observe resolve 'batch'   # resolve by substring match
```

Valid types: `idea`, `task`, `bug`, `gap`, `debt`.

Observations live in the `observations:` section of folio.yml. `folio observe lint` validates their format.

Observation and source-registration commands validate their projected manifest changes.
If a new observation or gathered source would introduce a blocking finding, Folio rolls
back the manifest and any materialized artifact created by that command.


## Home Sync: Managing Folio Content Stores

`FOLIO_UMBRELLA` names the control root containing `stores.yml`; `FOLIO_HOME`
names the selected content work root. All git operations on a Folio content
store go through `folio home` subcommands — never use raw git or jj:

```bash
folio home list    # dashboard for the selected store
folio home push    # commit and push the selected store
folio home pull    # pull the selected store
folio home push <store> -m "type(scope): description"  # sync one registered store
```

`folio home push` runs lint as a gate. If any project's folio.yml has validation
errors, the push is blocked. Projects move between `active/` and `archive/` via
`folio archive`; active projects appear in `folio home list`, archived ones do not.

## Working with jf Forests

When a folio target has a `forest:` block, composition delegates to `jf` for Jira hierarchy management. The composed markdown files map to Jira ticket descriptions through jf's forest model.

This integration means you can plan in folio (design doc, work tracks) and publish to Jira (epic and child ticket descriptions) through a single pipeline. See the [jf documentation](../../jf/docs/01-getting-started.md) for forest details.
