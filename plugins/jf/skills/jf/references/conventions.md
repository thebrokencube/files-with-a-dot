# Jira Conventions

Standards for ticket structure, naming, and content. These are opinionated conventions — they override Claude's generic Jira defaults.

## Ticket Naming

Stories/Tasks/Bugs use conventional commit format. Higher-level tickets use plain names with a bracketed prefix for scanability in flat views.

### Format by Level

- **Stories/Tasks/Bugs**: `type(scope): description`
  - e.g., `feat(resolver): add fallback heuristic`
- **Epics (no parent)**: `[ACRONYM] Name`
  - e.g., `[DPD] Decklet Plugin Development`
  - Acronym derived from epic name initials
- **Epics (with parent)**: `[PARENT/Scope] Name`
  - e.g., `[AID/RADR] Backlog`, `[AID/Docs] Legacy Docs Cleanup`
  - PARENT is the parent Project Name's abbreviation
  - Scope is the domain within that project
- **Project Names**: Plain descriptive name
  - e.g., `AI Docs`, `State Retirement Mandates`
  - Establish a short abbreviation (AID, SRM) for child epics to use in their prefix
- **Initiatives**: Plain name
  - e.g., `Platform Reliability`

### Hierarchical Tags

The bracketed prefix encodes ancestry so tickets are scannable in flat Jira views (boards, JQL results):

- **Top-level epic (no parent)**: `[ACRONYM]` — acronym from the epic's own name
- **Epic under Project Name**: `[ABBREV/Scope]` — project abbreviation + domain scope
- **Deeper nesting**: extend the path — `[ABBREV/Scope/Sub]`

The abbreviation is established when the Project Name is created and used consistently by all child epics. Keep it short (2-4 chars): AID, SRM, DPD.

Tags are always leading (prefix position), never trailing or backtick-wrapped.

## PR Title Format

When creating PRs for work tracked in Jira, prefix the title with the ticket key in brackets:

```
[PROJ-123] type(scope): description
```

The ticket key goes at the **beginning**, not the end. The rest follows conventional commit format (same as ticket naming). Examples:

```
[RETIRE-3108] feat(plans): add plan category system
[BEN-1234] fix(enrollment): handle missing SSN gracefully
```

If the branch name contains a ticket key (e.g., `RETIRE-3108-plan-categories`), extract it from there. If no ticket is associated, omit the prefix — don't invent one.

## Description Structure

By hierarchy level:

- **Tasks/Stories**: `## Context`, `## Goal`, `## Scope` (h2 sections)
- **Epics**: `## Context`, `## Goal` (no Scope — scope lives in child tickets)
- **Initiatives**: Minimal — context and goal only, children carry detail

## Content Rules

- Inline links (not reference-style). No footer sections (references, changelog).
- No downward duplication — parent descriptions should not repeat child content. Each ticket stands alone for its level.
- Source files must be pure precompile input — no metadata headers (title, status, type). Metadata belongs in frontmatter (`jira:`, `type:`, `sync:`, etc.). `jf push` strips YAML frontmatter automatically before converting to ADF. Source files should start directly at content after frontmatter (e.g., `## Goal`).

## Issue Type Hierarchy

Highest to lowest:

- Level 3: Initiative
- Level 2: Project Name
- Level 1: Epic
- Level 0: Story / Task / Bug
- Level -1: Sub-task

Directory structure mirrors Jira hierarchy — project root = top-level issue, subdirectories = child issue types.

## Forest Placement

jf forests (`.jf/` directories) live at two tiers within folio, each with a different lifecycle:

### Project-root forest

```
ret/radrs/.jf/          ← lives as long as the folio does
ret/radrs/folio.yml
ret/radrs/work/...
```

Use for **ongoing backlog epics** — the folio's standing relationship with Jira. The forest target in folio.yml points at `.jf/` directly. This forest persists across work tracks and sessions.

### Work-track forest

```
folio-hygiene/work/active/2026-05-11-ai-docs/.jf/    ← archives with the work track
```

Use for **operational Jira work** scoped to one piece of work — creating an initiative, restructuring hierarchy, one-time migrations. The forest exists to execute a specific Jira operation and archives with the work track when done.

### Deciding which tier

| Signal | Placement |
|---|---|
| Epic will receive new tickets over time | Project root |
| Epic is a backlog/catch-all for a domain | Project root |
| Jira operation is one-time (create initiative, restructure) | Work track |
| Forest will be archived when the work is done | Work track |

A folio can have both — a project-root forest for its ongoing epic and work-track forests for operational tasks.

## Folio ↔ Jira Mapping

Default expectation for how folio layers map to Jira hierarchy:

| Folio layer | Jira level | Example |
|---|---|---|
| Project group (`ret/`) | Project Name (level 2) | RETIRE-9338 "AI Docs" |
| Project (`ret/radrs`) | Epic (level 1) | RETIRE-9336 "[AID/RADR] Backlog" |
| Work track | Task/Story/Spike (level 0) | RETIRE-9341 "spike(plans): template modes" |
| Observation | Candidate for ticket creation | `folio observe` → `jf create-missing` |

This isn't 1:1 — a folio can have multiple epics (e.g. ret/docs has both a backlog epic and a legacy cleanup epic). But the mapping provides a default: when you create a new folio project, it probably needs an epic; when a project group emerges, it probably needs a Project Name.

## Project Defaults

Project-specific field values live in `~/.jf.yml` under `projects:`. See references/configuration.md for the schema.

When creating tickets, build a creation JSON with project-specific fields from `~/.jf.yml`. The creation JSON must NOT contain a `description` field — `acli create` silently drops inline ADF descriptions.

Example creation JSON:

```json
{
  "projectKey": "PROJ",
  "type": "Task",
  "summary": "feat(auth): add OAuth2 login flow",
  "parentIssueId": "PROJ-100",
  "additionalAttributes": {
    "components": [{"name": "My Component"}],
    "customfield_NNNNN": {"value": "Some Required Value"}
  }
}
```

Populate `additionalAttributes` from `~/.jf.yml` project defaults. Write creation JSON to `/tmp/{slug}-create.json`. `parentIssueId` is optional for top-level issues.

### Adding a New Project

1. Run `acli jira workitem create --generate-json` with the target project key to discover required fields
2. Use MCP `getJiraIssueTypeMetaWithFields` for full field metadata including allowed values
3. Add the project to `~/.jf.yml` under `projects:` with components and custom fields
4. Include no `description` field in creation templates — push separately via `jf push`
