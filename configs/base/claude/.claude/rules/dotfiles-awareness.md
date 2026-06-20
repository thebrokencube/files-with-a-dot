# Cross-Repo Awareness

Skill files (`~/.claude/skills/`) and rule files (`~/.claude/rules/`) are symlinked
from `~/.dotfiles`. When modifying a skill or rule while working in another repo,
commit the change in the dotfiles repo separately, following dotfiles commit
conventions (see dotfiles skill).

When a task touches files in multiple repos, commit each repo separately with
appropriate conventions.

**Dotfiles workflow:**
- Run `dot sync --dry-run` before committing dotfiles changes to preview
- Run `dot validate` (pre-existing shellcheck warnings in unrelated files are expected)
- Follow dotfiles skill conventions (conventional commits)

**Rebuilding CLI tools** (`folio`, `jf`, `dendrik`): Each has a Makefile in
`cmd/<tool>/`. Use `make check` (fmt + vet + test) then `make build` for a local
binary — it is **gitignored; never commit it**. Distribution is via **GitHub
Releases**: bump `cmd/<tool>/VERSION`, dispatch the release workflow
(`gh workflow run release.yml -f tool=<tool>`), and `dot sync` downloads the
release binary into `~/.local/bin/` (falling back to a local build if needed).
See `pkg/dendrik/conventions/release.md`.
