# System Context

My development environment, available globally across all projects.

## Setup
- **Dotfiles**: `~/.dotfiles` → github.com/thebrokencube/files-with-a-dot
- **Shell**: zsh + starship prompt
- **Editor**: Neovim (kickstart.nvim)
- **Terminal**: Ghostty (primary), iTerm2 (fallback)
- **Font**: Inconsolata Nerd Font

## Key Paths
- `~/.dotfiles/symlink_map.txt` - What's linked where
- `~/.dotfiles/shared/` - All config files
- `~/.env.local` - API keys (ANTHROPIC_API_KEY, etc.)
- `~/.gitconfig.local` - Git identity

## Quick Commands
- `v` / `nvim` - Editor
- `gs` - Git status
- `reload` - Restart shell
- `dot health` - Diagnose issues
- `dot status` - Show dotfiles state

## Git Preferences
- Do NOT add Co-Authored-By or other trailers to commits
- Use conventional commits: `type(scope): description` (see commit skill for details)
- When working with stacked branches, follow propagation and fixup rules (see stacked-pr skill for details)

## Cross-Repo Awareness
- **Skill files** (`~/.claude/skills/`) are symlinked from `~/.dotfiles`. When modifying a skill while working in another repo, commit the skill change in the dotfiles repo separately, following dotfiles commit conventions (versioned tags — see dotfiles-dev skill).
- When a task touches files in multiple repos, always commit each repo separately with appropriate conventions.
