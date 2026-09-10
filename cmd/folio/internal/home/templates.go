package home

// Templates for folio home init scaffolding.

const TemplateAgents = `# Folio Home

Use Folio's lifecycle and source-declaration conventions. Keep project knowledge versioned in this home.
`

const LegacyTemplateClaude = `# Folio Home
`

const TemplateClaude = `# Folio Home

@AGENTS.md
`

const LegacyTemplateReadme = `# Folio Home

See [CLAUDE.md](./CLAUDE.md) for details.
`

const TemplateReadme = `# Folio Home

See [AGENTS.md](./AGENTS.md) for guidance.
`

const TemplateGitignore = `.obsidian/
.DS_Store

# Store registry is machine-local: it holds absolute paths that differ per
# machine, so it must never sync across homes.
stores.yml
`
