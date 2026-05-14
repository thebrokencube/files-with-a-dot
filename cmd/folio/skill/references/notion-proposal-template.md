# Notion Proposal Template

Default template for tech specs and proposals published to Notion. Applied during
compose when a target has `external: notion` and the `how` field references this
template (or doesn't specify an alternative).

## Template Structure

The template defines both **body sections** and a **feedback table**. Compose
produces the full document — body + feedback — as a single output file.

```
# {Title}

**Authors:** {names}
**Stakeholders:** {team, channel, or individuals}
**Last Updated:** {date}

## TL;DR

{One-line description of the change or decision.}

## Context

{What is the issue motivating this? Prior art, related proposals.}

## Decision

{What are we doing? Who is doing it? Key details.}

## Impact

{What becomes easier or harder because of this change?}

## Timeline

{Phases, milestones, dates. Table or list.}

## Alternatives

{Serious alternatives considered but not chosen. Omit section if none.}

## References

{Links to epics, docs, code, prior art.}

---

## Feedback

Add your name, stance, and feedback below.

<table fit-page-width="true" header-row="true">
  <tr>
    <td>Reviewer</td>
    <td>Stance</td>
    <td>Feedback</td>
  </tr>
  <tr>
    <td>*add your name*</td>
    <td>👍 / 😬</td>
    <td> </td>
  </tr>
</table>
```

### Section Guide

| Section | Required | Notes |
|---------|----------|-------|
| Authors | Yes | Above TL;DR. Who wrote it. |
| Stakeholders | Yes | Above TL;DR. Who should read/review — team, channel, or individuals. |
| Last Updated | Yes | Above TL;DR. Date of last substantive change. |
| TL;DR | Yes | One sentence. If you can't say it in one line, the decision isn't clear yet. |
| Context | Yes | Problem statement + motivation. Link prior proposals/RADRs if they exist. |
| Decision | Yes | What we're doing, concretely. Can include sub-sections for work streams. |
| Impact | Yes | What changes for stakeholders. Keep it honest — include downsides. |
| Timeline | Yes | Table or list with phases and dates. Link to Jira tickets. |
| Alternatives | No | Only if serious alternatives were considered. Delete section if none. |
| References | Yes | Links to epics, source docs, code locations, related specs. |
| Feedback | Yes | Always last. Divider before it. |

### Stance Values

| Emoji | Meaning |
|-------|---------|
| 👍 | Aligned — no concerns, approve as-is |
| 😬 | Concerned — has feedback that should be addressed |

Reviewers edit the table directly in Notion to add their row.

## When to Apply

Apply this template when composing any target with `external: notion` that represents
a proposal, tech spec, or design document — anything that benefits from structured
reviewer feedback.

Do NOT apply to:
- Launch plans (operational, not decisional)
- Reference docs or glossaries
- Status updates or reports

The `how` field controls this. To opt in, include "Apply the Notion proposal template"
in the target's `how`. To opt out, omit it or say "No feedback table."

## Edit and Publish Flow

**Notion targets are write-only.** All edits happen in the local output file. Publish
replaces the full page content. Never use `update_content` for incremental Notion edits —
compose the final version locally, then publish once with `replace_content`.

The iterative loop stays local:

```
align on changes → edit output file → review → repeat → publish once
```

Do not edit Notion directly during the review loop. The local output file is the source
of truth; Notion is a render target.

## How to Apply During Compose

When composing a target that references this template:

1. Read the template section guide above
2. Map source content to the template sections (TL;DR, Context, Decision, etc.)
3. Drop the Alternatives section if nothing to say
4. Append the `---` divider + Feedback section
5. Write the combined output to the target's `path:` output

The Notion-flavored table markup above is what gets written to the output file.
When published via `mcp:notion`, Notion renders it as an editable table using
`replace_content` (full page replacement, not incremental updates).

## Extending (future)

This is the default template. Future work may add:
- Named templates selectable via `how` field (e.g., "Apply template: launch-plan")
- Template variants per audience (leadership vs engineering)
- Custom sections beyond feedback

For now, one template. Keep it simple.
