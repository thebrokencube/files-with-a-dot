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

### Bootstrapping a New Project

1. **Environment** (one-time per environment): Ensure Claude Code has context about what external systems are available to you. This can live in any CLAUDE.md that's in scope — global, workspace, or repo level. If you've already set this up, skip it.

2. **Repo routing** (if multi-project repo): Add the project to the repo's CLAUDE.md with its path and entry point.

3. **Project entry point** (README.md or similar): Document the compilation model — what are the source files, what do they compile into, what's the dependency order. A diagram helps.

4. **`PLAN_MANIFEST.md`**: Create in the project root using the template below. Fill in your project's external systems, source mappings, and compilation targets.

5. **Editing rule**: Note in the repo CLAUDE.md (or project README) that when editing source files, agents should mark affected targets as stale in `PLAN_MANIFEST.md`.

### `PLAN_MANIFEST.md` Template

```markdown
# Build Manifest

Tracks compilation status between local source files and external targets.

**Last reviewed:** YYYY-MM-DD

---

## External Systems

| System | Used for | Access method |
|--------|----------|---------------|
| _e.g., Jira_ | _Epic descriptions_ | _MCP: editJiraIssue_ |
| _e.g., Google Doc_ | _Shared spec_ | _Manual: paste-from-markdown_ |

---

## Compilation Targets

### [Target name] ([source] → [target type])

_Description of what this compilation target does and any notes on format differences._

| Source | Target | Status | Last compiled |
|--------|--------|--------|---------------|
| _source-file.md_ | _target location or system_ | Clean | _YYYY-MM-DD_ |

_Repeat this section for each compilation target or tier._

---

## Pending

- [ ] _Items that need to be compiled_

---

## After Compilation

1. Update "Last compiled" dates in the tables above
2. Update "Last Propagated" dates in source file headers (if applicable)
3. Archive or clear the pending section
4. Commit the updates
```

## Finding Project Config

1. Check the repo's CLAUDE.md for project listings and entry points
2. Read the project's entry point (typically README.md) for the compilation model and file roles
3. Find **`PLAN_MANIFEST.md`** in the project root

`PLAN_MANIFEST.md` contains:
- **External systems**: What this project compiles to and how to access each target
- **Source mappings**: Which source files compile into which targets
- **Compilation status**: Clean vs Stale per target
- **Last compiled dates**: When each target was last updated
- **Compilation workflows**: How to compile each target type (MCP, manual, etc.)

## Commands

### `compile [target]`

Compile source files into their external targets.

**No target or `all`**: Compile everything stale.

1. Read `PLAN_MANIFEST.md`, identify stale targets
2. Compile in dependency order (intermediate files before their downstream targets)
3. For each target, follow the workflow documented in `PLAN_MANIFEST.md`:
   - **Intermediate file** (assembled from multiple sources): Read sources, reformat to match the target's style, update the file
   - **External description** (pushed to an external system): Condense source content appropriately, push via the access method listed in `PLAN_MANIFEST.md`
   - **Manual target** (requires user action): Provide instructions and wait for confirmation
4. For manual-only targets, instruct the user on what to do

**Specific target**: Compile one target by name (names come from `PLAN_MANIFEST.md`).

After compilation:
1. Update status to **Clean** and "Last compiled" dates in `PLAN_MANIFEST.md`
2. If source files have "Last Propagated" header fields, update those too
3. Ask user about committing: "Commit compilation updates? (now / batch later)"

### `audit [scope]`

Check the state of the project. Think of this as `git status` for the compilation system.

**`local`** (default): Check compilation status and cross-file consistency without external calls.

**`external`**: Local checks plus validate against external systems listed in `PLAN_MANIFEST.md`, using available tools.

**Specific target** (e.g., `audit jira BEN-47872`): Validate one target against its external system.

#### What to check

1. **Compilation status** — Read `PLAN_MANIFEST.md`. Report which targets are Clean vs Stale. Check if source files appear modified since last compilation.

2. **Cross-file consistency** — Read all source files and cross-reference:
   - Statuses that appear in multiple files should agree
   - Dates should reflect the most recent edit session
   - Shared facts (lists, references, identifiers) should be consistent everywhere
   - File references should resolve

3. **External validation** (if `external` scope) — For each external system in `PLAN_MANIFEST.md`, use available tools to fetch current state and compare against local sources. Flag meaningful differences, not formatting noise.

#### Output format

Group by section. Use severity levels:

- ✅ **Clean** — consistent, no action needed
- ⚠️ **Warning** — possibly stale or minor inconsistency
- ❌ **Error** — clear inconsistency that should be fixed

```
## Compilation Status
- [target]: Clean / Stale (details)

## Cross-File Consistency
- ✅ / ⚠️ / ❌ finding

## External Validation (if requested)
- [target]: Matches ✅ / Differs ⚠️ (what changed)
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
