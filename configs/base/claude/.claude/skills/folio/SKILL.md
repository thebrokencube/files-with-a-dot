---
name: folio
description: Knowledge work lifecycle — research, plan, compile, audit. Replaces the plan skill's PLAN_MANIFEST.md with folio.yml and derived status.
user_invocable: true
---

# Folio

Lifecycle toolkit for knowledge work projects. Local source files compile into external targets (Jira descriptions, Google Docs, specs), with `folio.yml` declaring the structure and status derived at read time from file mtimes.

## Tiers

Projects naturally fall into tiers based on complexity. The tier determines what infrastructure you need.

| Tier | What it is | Infrastructure | Example |
|------|-----------|---------------|---------|
| 0 | Single scratch file | Nothing | One-off research note |
| 1 | Multi-file research | README.md entry point | Architecture investigation |
| 2 | Local compilation targets | folio.yml + compiled/ | Spec with local summary doc |
| 3 | External targets | folio.yml + tooling.yml | Investigation → Jira comment |

Tier 0-1 don't need folio.yml. Tier 2+ do. Tier 3 adds external system integration.

## Architecture

```
Layer              Provides                                  Location
-----              --------                                  --------
Skill              The generic pattern (this file)           ~/.claude/skills/folio/
Tooling defaults   System type -> access method              ~/.claude/skills/folio/tooling.yml
Environment        What external systems are available       CLAUDE.md (global/workspace/repo)
Repo routing       Which projects exist, where to start      <repo>/CLAUDE.md
Project config     Source mappings, targets, structure        <project>/folio.yml
```

### File Roles

- **Source files**: Working memory. Accumulate detail — open questions, checklists, decision rationale. Written for authors and agents. Single source of truth for their content.
- **Reference files**: Domain knowledge consumed during compilation but not compiled themselves. Architecture docs, prior art, etc.
- **Compiled targets**: Communication layer. Two flavors:
  - *Summary targets*: High-level overview, significantly condensed from sources.
  - *Tab-ready targets*: Cleaned-up source material — strip local metadata, normalize formatting, make paste-ready.
- **Infrastructure** (folio.yml, README.md entry point): Operational state.

**Compile = distill**: Compilation is distillation. Match target detail to its role and audience. Target instructions in folio.yml specify the transformation — what to condense, strip, and preserve.

## Finding Config

1. Look for `folio.yml` in the current directory or use the `--folio` flag
2. If no folio.yml exists, the project is Tier 0/1 (no folio infrastructure needed)
3. If folio.yml has only local outputs, it's Tier 2
4. If folio.yml has external outputs (`external:` fields), it's Tier 3

The skill reads `folio.yml` for all project-specific config. No other manifest files needed.

## Tooling Resolution

When a target has an external output (e.g., `external: jira`), the skill resolves it through `tooling.yml`:

1. Read `external:` field from the target's output in folio.yml (e.g., `jira`)
2. Look up that system in `tooling.yml` (lives in the skill directory)
3. Get the `pull` and `push` methods (e.g., `mcp:jira-confluence`)
4. Use those methods for pulling from / pushing to external systems

**Graceful degradation**: If a system isn't listed in tooling.yml, pull is skipped and push is manual. The skill reports this rather than failing.

## Link Model

Every correspondence between a local file and an external resource is a **link**. Links are bidirectional — push and pull are operations on links, not properties of the resource.

| Context | Local side | External side |
|---|---|---|
| Tree node | `file:` | tree system + node id |
| Batch item | `source:` | item output |
| Target output | `path:` sibling | `external:` + `id:` |
| Derived source | `path:` | `derived_from[].external` |

## Local-by-Default

All commands default to local-only, offline, fast operations. External interaction requires an explicit flag.

| Command | Default (no flag) | With flag |
|---|---|---|
| `folio validate` | Structural schema validation. Always local. | (no external flag) |
| `folio status` | Local mtimes only. External links show `unknown`. | (no external flag) |
| `folio compile` | Local compilation only (source → compiled/). No push, no pull. | `--push` pushes to externals. `--pull` fetches from externals. |
| `folio audit` | Local integrity: structural validation, cross-reference consistency, mtime checks. | `--external` fetches from external systems and compares. |

## Derived Status

There is no stored status in folio.yml. Status is always computed at read time.

| Output type | How status is derived | Possible states |
|---|---|---|
| Local file (`path:`) | Compare source mtimes vs output mtime | clean, stale, missing |
| External (`external:`) | Can't check without audit | unknown |

- **clean**: all source files exist, output exists, output mtime >= all source mtimes
- **stale**: output exists but at least one source is newer
- **missing**: output file doesn't exist yet
- **unknown**: external-only target, requires audit to check

After the skill compiles source to a local output, the output mtime is naturally newer than sources — status flips to "clean" with no explicit marking.

Run `folio status` to see current state.

## Commands

### `/folio status`

Run `folio status`. Shows derived state of all targets with DAG-aware transitive staleness.

Flags: `--folio PATH`, `--json`

### `/folio validate`

Run `folio validate`. Structural validation of folio.yml including DAG checks (output collisions, edge inference, cycle detection, transform/precompile rules).

Flags: `--folio PATH`, `--json`

### `/folio init`

Run `folio init --name "Project Name"`. Bootstrap a new folio.yml in the current directory.

### `/folio setup`

Run `folio setup`. Check folio binary is installed and report version.

Flags: `--check` (silent mode)

### `/folio compile [target]`

Compile source files into their targets.

**No target or `all`**: Compile everything stale or unknown.

1. Run `folio validate` to check structural integrity
2. Run `folio status` to identify what needs compilation
3. For each stale/missing/unknown target, in dependency order (DAG-aware):
   - Read source files listed in the target
   - Apply the transformation described in the target's `instructions`
   - For local outputs: write the compiled file (mtime updates naturally → clean)
   - For external outputs: push via the method from tooling.yml, or provide manual instructions
4. Report what was compiled and current status

**Specific target**: Compile one target by name (target IDs from folio.yml).

**Code references**: When source files contain links to code repositories, preserve them during compilation. Repository URL patterns from folio.yml's `repositories` section should produce clickable links in targets that support them.

#### Tree target compilation

Tree targets contain multiple internal nodes, each with its own linked file, instructions, and external output. Compile bottom-up (children before parents):

1. Run `folio status --json` to get per-node status (`TreeNodeStatus` in the JSON output)
2. Read the tree definition from `folio.yml` to get per-node `file` and `instructions` fields (status output has node IDs and status, but not compilation instructions)
3. Walk the tree bottom-up. For each stale/missing node:
   - Read the node's `file`
   - Apply the node's `instructions` as compilation instructions
   - Push to the external system via tooling.yml (using `tree.system` and `tree.field`)
4. After all nodes are compiled, touch the manifest output file (first `path:` in target outputs) to update its mtime

Each tree node compiles independently from its own source file — nodes do NOT consume child outputs. Bottom-up order matters for the external system (parent descriptions may reference children), not for data flow.

#### Jira Push Pipeline

Tree targets with `system: jira` and `compiled_ext: .json` use a two-phase pipeline: source markdown is compiled to ADF JSON intermediates, then pushed via acli.

**Pipeline overview:**

```
source .md → lint (md-to-adf --lint) → compile (md-to-adf --acli) → compiled .json → push (acli)
```

**Phase 1 — Lint:** Run the `lint` command from tooling.yml on the source file. If unsupported markdown features are found, report the issues to the user and stop. Do not produce broken ADF.

**Phase 2 — Precompile:** Run the `precompile` command from tooling.yml to convert source markdown to ADF JSON. The template placeholders resolve from tree node fields:

| Placeholder | Resolves from |
|-------------|---------------|
| `{id}` | Tree node `id` (Jira key) |
| `{source}` | Tree node `file` (source markdown path) |
| `{compiled}` | `{compiled_dir}/{id}{compiled_ext}` |

**Phase 3 — Push:** Push the compiled JSON to Jira via acli: `acli jira workitem edit --from-json {compiled} --yes`

**`compiled_dir` and `compiled_ext` convention:** When a tree target declares these fields, compiled artifacts go to `{compiled_dir}/{id}{compiled_ext}`. This gives inspectable intermediate artifacts that can be reviewed before push. Example: `compiled_dir: compiled/jira/`, `compiled_ext: .json` → node BEN-47882 compiles to `compiled/jira/BEN-47882.json`.

**Known limitations** (what `md-to-adf --lint` catches):

- No tables (`| ... |` syntax)
- No fenced code blocks (`` ``` ``)
- No blockquotes (`>`)
- No nested lists (indented `- ` or `1. `)
- No level 3+ headings (`###`)

These are limitations of the md-to-adf script, not of ADF itself. Source files that use these features must be flattened before compilation or use an alternative pipeline.

**Concrete example** — compiling one tree node:

```bash
# 1. Lint
md-to-adf --lint unify/testing/BEN-47882.md
# Exit 0 → clean

# 2. Precompile (source → ADF JSON)
md-to-adf --acli BEN-47882 < unify/testing/BEN-47882.md > compiled/jira/BEN-47882.json

# 3. Push to Jira
acli jira workitem edit --from-json compiled/jira/BEN-47882.json --yes
```

#### Batch target compilation

Batch targets contain multiple items sharing a pattern. Each item has its own source and external output:

1. Run `folio status --json` to get per-item status (`BatchItemStatus` in the JSON output)
2. For each stale/missing item:
   - Read the item's `source` file
   - Apply the target-level `instructions` as compilation instructions
   - Push to the external system. The output is resolved from the item's `output` merged with batch-level `system` and `field` defaults.
3. After all items are compiled, touch the manifest output file

### `/folio audit [scope]`

Check project state. Like `git status` for the compilation system.

**`local`** (default): Check compilation status and cross-reference consistency without external calls.

**`external`**: Local checks plus validate against external systems using tooling.yml methods.

**Specific target**: Validate one target against its external system.

#### Audit steps

1. **Structural validation**: Run `folio validate`
2. **Derived status**: Run `folio status`, report all targets
3. **Cross-reference consistency**: For each entry in `cross_references`, read the source of truth and all locations in `also_appears_in`, flag differences
4. **External validation** (if `external` scope): Fetch external targets using tooling.yml methods, compare semantically against sources. Account for expected format differences noted in target instructions.

#### Output format

```
## Status
- [target-id]: clean / stale / missing / unknown (details)

## Cross-References
- [fact]: clean / warning / error (what differs where)

## External Validation (if requested)
- [system] [id]: matches / differs (what changed)
```

Audit only reports — it does not fix anything.

## folio.yml Schema Reference

```yaml
schema: 1                          # Schema version (required)
project: "Project Name"            # Human-readable name (required)

# Project-level sources — where information comes from (optional)
sources:
  # Primary local — you wrote this
  - path: README.md

  # External — you read this from a remote system
  - external: jira
    id: "ACME-123"
    notes: "Original support ticket"

  # Derived — you summarized/cached this from external sources
  - path: reference/openai-harness.md
    derived_from:
      - external: web
        url: "https://example.com/article"
        cached: "2026-01-15"
        notes: "Summarized article. Lossy — stripped examples."
      - external: github
        url: "https://github.com/Org/repo"
        cached: "2026-01-15"
        notes: "Code patterns verified against repo"

  # Code — you're citing a repo or file
  - external: github
    id: "Acme/webapp"
    path: "src/models/account.rb"

# Code repository URL patterns (optional)
repositories:
  repo_name: "https://github.com/Org/repo/blob/main/{path}"

# Compilation targets (keyed by ID)
targets:
  target-id:
    instructions: "What this target does, how to transform source → output"
    transform: distill             # Required: distill, extract, adapt, compose, scaffold
    blocked_by: [other-target-id]  # Optional dependency
    sources:
      - path: relative/file.md     # Relative to folio.yml directory
    outputs:
      - path: compiled/output.md   # Local file output
      # OR
      - external: jira             # External system (resolved via tooling.yml)
        id: "PROJ-456"             # Resource identifier
        field: description         # Where within the resource to write

    # Batch targets — multiple items sharing a pattern
    batch:
      system: google_docs          # Default external for all items
      field: body                  # Default field (or omit if items differ)
      items:
        - id: "tab-name"
          source: compiled/tab.md
          output:
            id: "google-doc-id"
            field: "Tab Name"      # Per-item field (e.g., Google Doc tab)

    # Tree targets — hierarchical structures (e.g., Jira initiative)
    tree:
      system: jira
      field: description
      compiled_dir: compiled/jira/  # Optional: directory for compiled intermediates
      compiled_ext: .json           # Optional: extension for compiled files ({compiled_dir}/{id}{compiled_ext})
      root:
        id: "PROJ-100"
        label: "Initiative"
        file: README.md
        instructions: "Compilation instructions for this node"
        children:
          - id: "PROJ-200"
            label: "Project"
            file: project/README.md
            instructions: "Per-node compilation instructions"
            children:
              - id: "PROJ-300"
                label: "Epic"
                file: project/epic.md

# Facts shared across files (optional)
cross_references:
  - fact: "Description of the fact"
    source_of_truth: "path/to/file.md § Section Name"
    also_appears_in:
      - "other/file.md § Section"

# Open work items (optional)
tasks: []

# Notes / blockers (optional)
pending:
  - "Description of what's pending"
```

## Output Addressing

External outputs use three fields to identify their destination:

| Field | Role | Examples |
|-------|------|---------|
| `external` | System type | `jira`, `google_docs`, `confluence`, `slack` |
| `id` | Resource identifier | Jira key (`PROJ-123`), Google Doc ID, Confluence page ID |
| `field` | Where within the resource | Jira: `description`, `comment`. Google Docs: tab name (`Glossary`). |

The `field` identifies the sub-resource to write to. For systems with a single writing surface (e.g., a Confluence page body), `field` is optional. For systems with sub-resources (Jira issue fields, Google Doc tabs), `field` disambiguates.

The CLI uses `ext:<system>:<id>:<field>` as the output map key. Two outputs with the same `system` and `id` but different `field` values are distinct (no collision).

## Source Kinds

Project-level `sources` declare where information comes from. Four kinds:

| Kind | Schema shape | Where it lives | Staleness risk | Audit |
|------|-------------|----------------|---------------|-------|
| Primary | `- path: file.md` | Local | You control it | mtime |
| External | `- external: jira` + `id:` | Remote system | They change it anytime | Fetch + compare (if MCP available) |
| Derived | `- path: file.md` + `derived_from:` | Local (cached from external) | Upstream changes silently | Report cache age |
| Code | `- external: github` + `id:` + optional `path:` | Repo | Deploys, refactors | git log / blame |

**Kind detection**: A source entry with `derived_from:` is derived. A source with only `path:` is primary. A source with `external:` (and no `path:`) is external or code (code uses `external: github`).

**Provenance**: Derived sources carry a `derived_from` list — one or more entries, each with `external:`, optional `url:`, `cached:` date, and `notes:`. Multi-source derivation (e.g., a reference doc synthesized from codebase analysis + a legal survey) lists all upstreams. The CLI reports cache age based on the oldest `cached` date across all entries. This is the audit trail — folio tracks where the cached content came from without pretending it can automatically re-derive it.

**Audit behavior**:
- Primary: mtime comparison — deterministic
- External (MCP-accessible): fetch and semantic compare
- Derived (cached): report age, flag if old. Cannot automatically re-summarize or diff against upstream.
- External (no MCP): manual check only

**External sources can move between `sources` and `outputs`** as project ownership shifts. Same Jira ticket, different role depending on whether you're writing to it or reading from it.
