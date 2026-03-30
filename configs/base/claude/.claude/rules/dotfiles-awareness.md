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
