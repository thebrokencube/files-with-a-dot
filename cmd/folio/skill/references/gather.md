# Gather Workflow

Read by `/folio gather [args]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

Gathering brings sources into the folio. It spans from a simple CLI scaffold to a full multi-agent research session.

## The Gather Spectrum

```
folio gather <url>                              # CLI: scaffold source entry in folio.yml
folio gather <url> --materialize --type <type>  # CLI: scaffold + create typed reference file
folio gather <url> --read                       # Requires Claude skill (prints message and exits)
/folio gather <topic>                           # Skill: parallel agent research, synthesize, materialize
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

## Skill: Deep Gather (`/folio gather <topic>`)

When invoked as a skill (not a URL), gather becomes a research workflow:

1. Identify what to research from the topic
2. Search available sources (MCP tools, web, codebase). If no results found, report "no results for {topic}" — do not write an empty file.
3. Synthesize findings (hold in memory, do not write yet)
4. **Infer type** from the research content:
   - External system/tool summary → `survey`
   - Time-boxed investigation findings → `spike`
   - Multi-source distillation → `synthesis`
   - Domain knowledge capture → `domain`
   - Default for ambiguous content → `spike`
5. **Review gate (soft)**: Present proposed type, filename, content length, and 3 key facts. "Write to {path}?"
6. Materialize via `folio new <inferred-type> <topic>` — never raw file write. Then fill the template with synthesized content.
7. Report what was gathered, the inferred type, and where it was placed

## Knowledge Graduation

Every artifact has a maturity level. Gather moves it forward:

```
URL -> source entry -> reference file -> (compose takes over from here)
```

| From | To | How |
|------|----|-----|
| URL | Source entry | `folio gather <url>` |
| Source entry | Reference file | `folio gather <url> --materialize --type <type>` |
| URL | Reference with content | `/folio gather <topic>` (skill mode) |

> Schema reference: see references/schema.md for the source entry format.

## Error Handling

- **Empty search results**: Report "no results for {topic}" and stop. Do not write an empty reference file.
