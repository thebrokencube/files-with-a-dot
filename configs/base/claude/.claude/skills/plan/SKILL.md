---
name: plan
description: Project planning toolkit — compile source files to external targets and audit project state. Use for projects that follow a compilation model with a PLAN_MANIFEST.md.
user_invocable: true
---

# Plan

Toolkit for projects that follow a **compilation model**: local source files compile into external targets (descriptions, docs, specs), with a `PLAN_MANIFEST.md` tracking what's clean and what's stale.

## Architecture

The compilation model separates **what the pattern is** from **what tools you have** from **what a specific project needs**. This separation means the skill works across different environments and projects without changes.

### Layers

```
Layer              Provides                                  Example location
─────              ────────                                  ────────────────
Skill              The generic pattern (this file)           ~/.claude/skills/plan/
Environment        What external systems are available       CLAUDE.md (global, workspace, or repo)
Repo routing       Which projects exist, where to start      <repo>/CLAUDE.md
Project config     Source mappings, targets, state            <project>/README.md + PLAN_MANIFEST.md
```

**Skill** (this file): The compile/audit patterns. Knows nothing about specific projects, tools, or external systems. Portable across any setup.

**Environment**: Configuration that tells the skill what external systems and tools are available — e.g., "Jira is available via MCP," "Google Drive is available via MCP." This is what differs between environments (a work setup might have Jira and Confluence; a personal setup might have only GitHub). Where this lives is up to you — a global CLAUDE.md, a workspace CLAUDE.md, or even inline in a repo's CLAUDE.md. The skill discovers it from whatever CLAUDE.md context is present.

**Repo routing**: If a repo contains multiple projects, its CLAUDE.md lists them with paths and entry points. For single-project repos, this can be minimal or combined with the project config.

**Project config**: The project's own files. An entry point (typically README.md) describes the compilation model and file roles. `PLAN_MANIFEST.md` tracks source mappings, external systems relevant to this project, compilation status, and workflows. All project-specific detail lives here.

### File Roles

Source files, compiled targets, and infrastructure serve different purposes. Knowing which role a file plays determines how to write it and what to expect during compilation.

- **Source files**: Working memory. Accumulate detail — open questions, checklists, decision rationale. Written for authors and agents. These are the single source of truth for their content.
- **Compiled targets**: Communication layer. Two flavors:
  - *Summary targets*: High-level overview, significantly condensed from sources. Audience is stakeholders who need the "what" without the "how."
  - *Tab-ready targets*: Cleaned-up versions of source material — strip local metadata (`- Last Updated:`, `- [ ]` checkboxes, `---` separators), normalize formatting, make paste-ready. Preserve substance but remove working-memory artifacts.
- **Infrastructure** (manifest, entry point navigation): Operational state. Note: an entry point (like README.md) may also be a compilation source — document this in the manifest so it's explicit.

**Directory convention** (recommended): Mirror the external issue hierarchy in the filesystem. The project root corresponds to the top-level issue. Subdirectories map to child issue types.

```
project/                     # Top-level issue (Initiative, Project Name, etc.)
├── {bucket}/                # Child issue groupings (if applicable)
│   ├── README.md            #   → Description for this level
│   └── {epic}/epic.md       #   → Description for child issue
├── reference/               # Source: domain knowledge (not issue-mapped)
├── compiled/                # Generated — edit source, recompile
├── PLAN_MANIFEST.md
└── README.md
```

**Compile = reduce**: Compilation is distillation. Match target detail to its role and audience. The manifest's target descriptions specify the transformation — what to condense, what to strip, what to preserve.

### Bootstrapping a New Project

1. **Environment** (one-time per environment): Ensure Claude Code has context about what external systems are available to you. This can live in any CLAUDE.md that's in scope — global, workspace, or repo level. If you've already set this up, skip it.

2. **Repo routing** (if multi-project repo): Add the project to the repo's CLAUDE.md with its path and entry point.

3. **Project entry point** (README.md or similar): Document the compilation model — what are the source files, what do they compile into, what's the dependency order. A diagram helps.

4. **`PLAN_MANIFEST.md`**: Create in the project root using the template below. Fill in your project's external systems, source mappings, and compilation targets.

5. **Editing rule**: Note in the repo CLAUDE.md (or project README) that when editing source files, agents should mark affected targets as stale in `PLAN_MANIFEST.md`.

### `PLAN_MANIFEST.md` Template

```markdown
# Build Manifest

Tracks compilation status between local source files and external targets. See [README.md](./README.md) for the compilation model overview.

**Last reviewed:** YYYY-MM-DD

---

## External Systems

| System | Used for | Access method |
|--------|----------|---------------|
| _e.g., Jira_ | _Epic descriptions_ | _MCP (jira-confluence server)_ |
| _e.g., Google Doc_ | _Shared spec_ | _Manual: paste-from-markdown_ |

---

## Repositories

_Optional. Document repository URL conventions when source files link to code._

| Repo | GitHub URL pattern |
|------|-------------------|
| _e.g., zenpayroll_ | `https://github.com/Org/repo/blob/main/{path}` |

_Use `blob/main/` for files, `tree/main/` for directories._

---

## Compilation Targets

### [Target name] ([source] → [target type])

_Description of what this compilation target does and any notes on format differences._

| Source | Target | Status |
|--------|--------|--------|
| _source-file.md_ | _target location or identifier_ | Clean |

**Last compiled:** _YYYY-MM-DD or "Not yet"_

_Repeat this section for each compilation target._

---

## Cross-Reference Schema

Facts that appear in multiple files. The audit checks that each fact in "Also appears in" matches the value in its source of truth.

| Fact | Source of truth | Also appears in |
|------|----------------|-----------------|
| _e.g., Epic statuses_ | _Per-epic file headers_ | _PROJECT.md, README.md_ |
| _e.g., Jira epic keys_ | _This manifest_ | _PROJECT.md, per-epic files_ |

_Small projects may have 0-2 rows. Larger projects should enumerate all facts that are duplicated across files. Only list facts where drift would cause real confusion — don't track every trivial mention._

---

## Pending

_Operational status: what needs attention next. Keep concise — stale targets and action items, not changelog._
```

### Template notes

- **Scaling**: Small projects might have one compilation target and no cross-references. Large projects might have many of each. The sections are optional — omit Cross-Reference Schema entirely if there's nothing to track.
- **Status column**: Maintained by agents when editing source files (mark as **Stale**) and by the compile command (mark as **Clean**). The audit verifies this by content comparison — it doesn't trust the column.
- **External Systems**: List MCP server names and capabilities, not specific tool names. The skill discovers tools from what's available at runtime.
- **Target descriptions**: The description text for each compilation target drives both compilation and audit behavior. Document not just *what* compiles where, but *how* the format changes — what's condensed, what's excluded, what's reformatted, and what's local-only. Vague descriptions lead to incorrect compilation and false audit results.
- **Format alignment**: When possible, use the same format in source and target files (e.g., both using tables for structured data). This reduces compilation to "condense prose, preserve structure" rather than "reformat tables to bullets." Format translation between source and target should be the exception, not the default.

## Finding Project Config

1. Check the repo's CLAUDE.md for project listings and entry points
2. Read the project's entry point (typically README.md) for the compilation model and file roles
3. Find **`PLAN_MANIFEST.md`** in the project root

`PLAN_MANIFEST.md` contains:
- **External systems**: What this project compiles to and how to access each target
- **Compilation targets**: Which source files compile into which targets, with status
- **Cross-reference schema**: Which facts are shared across files, with their source of truth
- **Compilation workflows**: How to compile each target type (MCP, manual, etc.)

## Commands

### `compile [target]`

Compile source files into their external targets.

**No target or `all`**: Compile everything stale.

1. Read `PLAN_MANIFEST.md`, identify stale targets
2. Compile in dependency order (intermediate files before their downstream targets)
3. For each target, follow the workflow documented in `PLAN_MANIFEST.md`. Apply the file role (see File Roles above) to determine the transformation:
   - **Intermediate file** (assembled from multiple sources): Read sources, apply the target's role (summary or tab-ready), update the file
   - **External description** (pushed to an external system): Condense source content per its role, push via the access method listed in `PLAN_MANIFEST.md`
   - **Manual target** (requires user action): Provide instructions and wait for confirmation
4. For manual-only targets, instruct the user on what to do

**Specific target**: Compile one target by name (names come from `PLAN_MANIFEST.md`).

**Code references**: When source files contain links to code repositories (e.g., GitHub URLs), preserve them during compilation. When compiling to targets that support links (Jira, Google Docs via paste-from-markdown), file references should remain clickable. Code blocks (directory trees, DSL examples) are excluded — only link references in tables and inline text.

After compilation:
1. Update status to **Clean** in `PLAN_MANIFEST.md`
2. Update **Last compiled** dates in `PLAN_MANIFEST.md`
3. If source files have "Last Propagated" header fields, update those too
4. Ask user about committing: "Commit compilation updates? (now / batch later)"

### `audit [scope]`

Check the state of the project. Think of this as `git status` for the compilation system. The audit is **data-driven** — it reads `PLAN_MANIFEST.md` to determine what to check, then verifies by reading actual content.

**`local`** (default): Check compilation status and cross-reference consistency without external calls.

**`external`**: Local checks plus validate against external systems listed in `PLAN_MANIFEST.md`, using available tools.

**Specific target** (e.g., `audit jira BEN-47872`): Validate one target against its external system.

#### Step 1: Compilation status (content comparison)

For each row in each **Compilation Targets** table:

1. Read the **source** file content
2. Read the **target** content (local file, or fetch from external system if `external` scope)
3. Compare semantically — does the target reflect the source? Account for expected format differences noted in the target section's description (e.g., tables → bullet points, condensed for Jira)
4. Report:
   - If content matches and Status says Clean → **Clean**
   - If content matches but Status says Stale → **Warning** (status column is wrong, should be Clean)
   - If content differs and Status says Stale → **Stale** (expected, needs compilation)
   - If content differs and Status says Clean → **Error** (status column is wrong, should be Stale)

For local audit, only check targets that are local files. For external audit, also fetch and compare external targets.

#### Step 2: Cross-reference consistency

For each row in the **Cross-Reference Schema** table:

1. Read the **source of truth** and extract the fact's current value
2. For each location in **Also appears in**, read and extract the same fact
3. Compare — do all locations agree with the source of truth?
4. Report each fact as Clean, Warning, or Error with specific details on what differs

If the project has no Cross-Reference Schema section, skip this step.

#### Step 3: External validation (if `external` scope)

For each external system in `PLAN_MANIFEST.md`, use available MCP tools to fetch current state and compare against local sources. Flag meaningful content differences, not formatting noise.

#### Output format

Group by section. Use severity levels:

- **Clean** — consistent, no action needed
- **Warning** — possible issue or minor inconsistency
- **Error** — clear inconsistency that should be fixed

```
## Compilation Status
- [target section]: [source → target]: Clean / Stale / Error (details)

## Cross-Reference Consistency
- [fact]: Clean / Warning / Error (what differs where)

## External Validation (if requested)
- [system] [target]: Matches / Differs (what changed)
```

Audit only reports — it does not fix anything. The user decides what to address.

## Usage

```
/plan compile                    # Compile everything stale
/plan compile tech-spec          # Compile a specific target
/plan compile jira all           # Compile all stale items for a system
/plan compile jira BEN-47872     # Compile one specific item
/plan audit                      # Full local check (default)
/plan audit local                # Same as above
/plan audit external             # Local check + external validation
/plan audit jira BEN-47872       # Validate one item externally
```

Target and system names come from the project's `PLAN_MANIFEST.md`.
