# Jira Restructure Playbook

Workflow for creating new epics/projects, reparenting tickets, and reorganizing Jira hierarchy. Validated through SRM reorganization (2026-05-11) and gwc epic split (2026-05-11).

## When to Use

- Splitting a monolithic epic into focused epics
- Creating a Project Name to group related epics
- Reparenting tickets from one epic to another
- Closing an old epic and migrating active work to new ones

## Phases

### Phase A — Prepare local forests

1. Create `.jf/` directory (project-root or work-track — see conventions.md Forest Placement)
2. Write `forest.yml` with `defaults.project: <PROJECT>`
3. Write `README.md` with frontmatter:
   - `jira: TBD` — required, without this jf ignores the file
   - `type: Epic` (or `Project Name`, `Initiative`) — must be explicit, jf doesn't infer
   - `label:` — the Jira summary, following naming conventions
4. For existing tickets being moved: create stub files with `jira: <KEY>` (keep existing keys)
5. For tickets not in the old forest: create from Jira description content
6. Validate: `jf validate` + `jf status --json`
7. `folio home push`

**Watch out:** Files without `jira:` frontmatter are invisible to jf. Audit with `jf validate` — TBD count should match expected new tickets.

### Phase B — Create tickets via jf create-missing

1. `jf create-missing --dry-run` — **hard gate**, review before proceeding
2. `jf create-missing` — creates in pre-order (parents before children)
3. Spot-check in Jira: verify parent, type, description for 2-3 tickets
4. `folio home push` — commit frontmatter updates (TBD → real keys, file renames)

**Watch out:** Some projects require custom fields for certain issue types (e.g. RETIRE Epics need `customfield_16528` "Is this capitalizable"). Check `~/.jf.yml` project config before creating. If creation fails with a required field error, add the field to `~/.jf.yml` and retry.

### Phase C — Reparent existing tickets

jf has no reparent command. Use MCP `editJiraIssue`:

```
editJiraIssue: { issueKey: "PROJ-123", parent: { key: "PROJ-456" } }
```

One call per ticket. Verify in Jira after each batch.

**MCP explanation gate still applies** — the reason jf can't handle this is that it has no reparent command.

### Phase D — Sync descriptions

1. `jf sync --dry-run` — **hard gate**
2. `jf sync --resolve local --yes` — push descriptions
3. **Verify root nodes landed.** `jf sync` can SKIP root/epic nodes due to:
   - Label-only changes (sync doesn't detect these — use `jf push <KEY> --plain-text`)
   - ADF roundtrip divergence (~24% of nodes may be read-only)
   - First-sync baseline mismatch
4. For any SKIP'd high-value nodes: `jf push <KEY> --plain-text --yes`

### Phase E — Verify via jf clone

`jf tree` is local-only — it reads `.jf/` not Jira. The only reliable way to verify the remote hierarchy:

```
jf clone <ROOT_KEY> --dir /tmp/verify-<name>
```

Diff the cloned structure against local forest. Confirm:
- Reparented tickets are under correct parent
- New epics/projects have the right children
- No orphaned tickets

Clean up: `rm -rf /tmp/verify-<name>`

### Phase F — Close old tickets

If replacing an old epic:

1. Get transitions: MCP `getTransitionsForJiraIssue`
2. Transition to Done: MCP `transitionJiraIssue` with the Done transition ID
3. Some transitions require fields (e.g. "Eng Weeks Estimation") — expand transitions with `expand=transitions.fields` to discover

Optionally reparent the closed epic under the new Project Name for historical grouping.

### Phase G — Update folio.yml

1. Add new epic as `external: jira` source
2. Add forest target with `root:` pointing at `.jf/`
3. Update old epic source note to mark as closed/historical
4. `folio validate` — catches stale forest root paths
5. `folio home push`

## Checklist

- [ ] Local forests prepared and validated
- [ ] `create-missing --dry-run` reviewed
- [ ] Tickets created, frontmatter committed
- [ ] Existing tickets reparented via MCP
- [ ] Descriptions synced, root nodes verified
- [ ] Remote hierarchy verified via `jf clone`
- [ ] Old tickets closed/transitioned
- [ ] folio.yml updated with new sources and targets
