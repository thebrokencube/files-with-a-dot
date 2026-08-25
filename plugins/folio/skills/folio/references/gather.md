# Gather Workflow

Read by `/folio gather [args]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

Gathering brings sources into the folio. Two shapes: **snapshot** (capture new knowledge) and **re-seed** (update existing research). Spans from a simple CLI scaffold to a full phased research session.

## The Gather Spectrum

```
folio gather <url>                              # CLI: scaffold source entry in folio.yml
folio gather <url> --materialize --type <type>  # CLI: scaffold + create typed reference file
folio gather <url> --read                       # Requires Claude skill (prints message and exits)
/folio gather <topic>                           # Skill: snapshot — research, synthesize, materialize
/folio gather <existing-file-path>              # Skill: re-seed — update existing research file
```

## CLI: URL Scaffold

`folio gather <url>` adds a source entry to folio.yml:

```yaml
sources:
  - path: reference/<type>/YYYY-MM-DD-<name>.md   # with --materialize --type
    derived_from:
      - external: web
        url: "https://..."
        cached: "YYYY-MM-DD"
```

Options:
- `--type <type>`: reference type (spike, survey, design, ...) — **required with `--materialize`**
- `--name <name>`: specify reference file name (default: derived from URL)
- `--materialize`: create a typed reference file and add `path:` + `derived_from:` entry
- `--read`: print "requires /folio gather (Claude skill)" and exit — clear seam for skill layer

## Skill: Gather Phases

When invoked as a skill, gather follows a phased workflow. Not every phase runs every time — Shape A (snapshot) often collapses Survey and Synthesize into one pass; Shape C (re-seed) skips Scope detection and starts from the existing file.

| Phase | What happens | Materialization |
|-------|-------------|----------------|
| Scope | Define question, detect shape, check coverage | None (conversational) |
| Survey | Breadth-first source search | None (hold in memory) |
| Synthesize | Cross-reference, resolve contradictions | None (hold in memory) |
| Materialize | Type inference, file write, review gate | Reference file committed |
| Connect | Observations, absorption ledger updates | Observations committed |

### Scope

1. Identify research question from the user's topic
2. Detect shape:
   - If the argument resolves to an existing file path → **Shape C (re-seed)**
   - Otherwise → **Shape A (snapshot)**
3. Check existing coverage: search `folio status` output and vault for files on this topic
4. If existing coverage found and user wants new research: confirm intent to avoid duplicates

### Survey

Search available sources (MCP tools, web, codebase). Use parallel searches when possible.

**Track sources as you go** — for each finding, note the URL, document title, PR/ticket number, Slack channel, or repo path where it came from.

Stop when: (a) 3+ independent sources confirm key facts, or (b) search returns diminishing results.

If no results found, report "no results for {topic}" — do not write an empty file.

### Synthesize

Cross-reference findings across sources. Resolve contradictions — note which source is authoritative and why.

For Shape A, Survey and Synthesize often collapse: read a repo or doc, synthesize as you go, write one file. This is fine when the topic is bounded to a single system.

### Materialize

1. **Infer type** from the research content:
   - External system/tool summary → `research`
   - Multi-source distillation → `research`
   - Domain knowledge capture → `domain`
   - Time-boxed investigation findings → `spike`
   - Default for ambiguous content → `research`
2. **Choose residency**: Default to project-scoped (`folio new <type> <topic>`). Use `folio new vault:<type> <topic>` only when content is obviously cross-cutting — landscape scan of external systems, tool comparisons, domain knowledge that spans projects.
3. **Review gate (soft)**: Present proposed type, filename, content length, and 3 key facts. "Write to {path}?"
4. **Write** via `folio new <type> <topic>`. Fill the template with synthesized content.
   - **Inline source attribution** throughout — cite URLs, PR numbers, ticket IDs near the claims they support. Don't batch into a bibliography; weave in where relevant.
   - **Note:** `folio new design` creates a work directory and colocates the design inside it. All other reference types scaffold to `reference/<type>/`.
5. Report what was gathered, the inferred type, and where it was placed

### Connect

Post-materialization feedback:

1. For files with an Absorption Ledger: update "Under Evaluation" entries based on new findings
2. Create observations for actionable findings: `folio observe 'idea(scope): description'`
3. If findings cross into recommendations ("we should do X"), note the boundary — suggest a follow-up spike rather than mixing knowledge and recommendations

## Shape A: Snapshot

Capture the current state of an external system, tool, or topic area.

- **Invocation**: `/folio gather <topic>`
- **Duration**: 10-20 minutes, single session
- **Output**: One reference file (typically `research` or `domain`)
- **Phase collapse**: Survey and Synthesize usually run as one pass — read the source material, synthesize as you go. Full phase separation only when multiple independent sources need cross-referencing.

## Shape C: Re-seed

Update existing vault or project research when upstream has changed.

- **Invocation**: `/folio gather <existing-file-path>`
- **Duration**: 10-30 minutes, single session
- **Output**: Updated existing file (edit, not rewrite)

### Re-seed Protocol

1. **Read existing file** — load current content, extract Tracking section metadata
2. **Fetch upstream** — use upstream signals from Tracking section:
   - GitHub releases URL → WebFetch latest releases
   - Slack channel → search recent messages
   - "manual check" → ask user what's changed
3. **Diff analysis** — compare current snapshot against upstream:
   - New features/patterns not in our file
   - Deprecated/removed features still in our file
   - Changed behavior on things we've absorbed
4. **Update file** — edit existing content, don't replace:
   - Add new sections for newly discovered patterns
   - Mark deprecated patterns with removal or strikethrough
   - Update `last_checked` date in Tracking section
   - Update Absorption Ledger if absorbed patterns changed upstream
5. **Review gate (soft)**: Present summary of changes. "Update {path}?"

Key principle: re-seed is an **edit**, not a rewrite. Preserve existing structure. The file should read as a coherent whole after re-seeding, not as "old content + appended new stuff."

## Gather vs Spike

When research naturally leads to "we should do X," you've crossed from gather territory into spike territory.

| Output answers... | Type | Where it lands |
|-------------------|------|---------------|
| "What exists?" | `research` or `domain` | Vault-eligible |
| "What should we do about it?" | `spike` | Project-scoped |
| Both | Split into two files | Research in vault, spike in project |

If you realize mid-gather that the content is becoming a recommendation, stop and split. The landscape goes in a research file; the recommendation goes in a spike. Don't mix them — it makes the research file less reusable.

## Knowledge Graduation

Every artifact has a maturity level. Gather moves it forward:

```
URL -> source entry -> reference file -> (compose takes over from here)
                                ^
                  re-seed updates ┘
```

| From | To | How |
|------|----|-----|
| URL | Source entry | `folio gather <url>` |
| Source entry | Reference file | `folio gather <url> --materialize --type <type>` |
| URL | Reference with content | `/folio gather <topic>` (Shape A) |
| Stale reference | Updated reference | `/folio gather <existing-file-path>` (Shape C) |

> Schema reference: see references/schema.md for the source entry format.

## Error Handling

- **Empty search results**: Report "no results for {topic}" and stop. Do not write an empty reference file.
- **Re-seed with no Tracking section**: Treat as Shape A — the file has no upstream signals to check. Ask the user what to look for.
- **Re-seed finds major upstream changes**: If the upstream system has changed so fundamentally that the existing file structure doesn't fit, suggest writing a new snapshot instead of force-fitting updates into the old structure. Present the choice to the user.
