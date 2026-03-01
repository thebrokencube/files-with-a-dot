# Gather Workflow

Read by `/folio gather [args]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

Gathering brings sources into the folio. It spans from a simple CLI scaffold to a full multi-agent research session.

## The Gather Spectrum

```
folio gather <url>                    # CLI: scaffold source entry in folio.yml
folio gather <url> --materialize      # CLI: scaffold + create reference file stub
folio gather <url> --read             # Requires Claude skill (prints message and exits)
/folio gather <topic>                 # Skill: parallel agent research, synthesize, materialize
```

## CLI: URL Scaffold

`folio gather <url>` adds a source entry to folio.yml:

```yaml
sources:
  - path: reference/<name>.md          # only with --materialize
    derived_from:
      - external: web
        url: "https://..."
        cached: "YYYY-MM-DD"
```

Options:
- `--name <name>`: specify reference file name (default: derived from URL)
- `--materialize`: also create a stub `reference/<name>.md` and add `path:` + `derived_from:` entry
- `--read`: print "requires /folio gather (Claude skill)" and exit — clear seam for skill layer

## Skill: Deep Gather (`/folio gather <topic>`)

When invoked as a skill (not a URL), gather becomes a research workflow:

1. Identify what to research from the topic
2. Search available sources (MCP tools, web, codebase)
3. Synthesize findings into a reference file
4. Add the source entry to folio.yml with `derived_from` provenance
5. Report what was gathered and where it was placed

## Knowledge Graduation

Every artifact has a maturity level. Gather moves it forward:

```
URL -> source entry -> reference file -> (compose takes over from here)
```

| From | To | How |
|------|----|-----|
| URL | Source entry | `folio gather <url>` |
| Source entry | Reference file | `folio gather <url> --materialize` |
| URL | Reference with content | `/folio gather <topic>` (skill mode) |

> Schema reference: see references/schema.md for the source entry format.
