# Jira Conventions

Standards for ticket structure, naming, and content. These are opinionated conventions — they override Claude's generic Jira defaults.

## Ticket Naming

All ticket names use conventional commit format. Higher-level tickets add a bracketed prefix for scanability in flat views.

### Format by Level

- **Stories/Tasks/Bugs**: `type(scope): description`
  - e.g., `feat(resolver): add fallback heuristic`
- **Epics (no parent)**: `[ACRONYM] type(scope): description`
  - e.g., `[DPD] feat(decklet): plugin development`
  - Acronym derived from epic name initials (Decklet Plugin Development → DPD)
- **Epics (with parent)**: `[PARENT/Bucket] type(scope): description`
  - e.g., `[AUTH/Infra] feat(session): add token rotation`
  - PARENT is the parent epic's acronym
- **Project Names (buckets)**: `[PREFIX] Name`
  - e.g., `[AUTH] Infrastructure`
- **Initiatives**: Plain name
  - e.g., `Platform Reliability`

### Hierarchical Tags

The bracketed prefix encodes ancestry so tickets are scannable in flat Jira views (boards, JQL results):

- **Top-level epic**: `[ACRONYM]` — acronym from the epic's own name
- **Epic under bucket**: `[PARENT/Bucket]` — parent's acronym + bucket name
- **Deeper nesting**: extend the path — `[PARENT/Bucket/Sub]`

Tags are always leading (prefix position), never trailing or backtick-wrapped.

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
