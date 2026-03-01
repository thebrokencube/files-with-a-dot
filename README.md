# files-with-a-dot

Personal dotfiles with symlink-based management, supporting multiple modes (aggressive/conservative) with shared configurations.

## Quick Start

### Prerequisites

**GitHub SSH must be configured** before running bootstrap. Test with:

```bash
ssh -T git@github.com
```

If this fails, [set up SSH keys first](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).

### Bootstrap

On a fresh macOS machine, run:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/cmd/dot/scripts/bootstrap.sh)"
```

This clones the repo to `~/.dotfiles`, installs dependencies via Homebrew, and runs sync to apply dotfiles state. First run prompts for machine type and git identity.

**Note:** If Homebrew can't install (sudo issues, etc.), the script clones the repo and shows how to complete setup manually.

## First-Time Setup

First sync automatically detects a fresh machine and prompts for:
- Mode — aggressive (repo-controlled packages + cleanup) or conservative (show opportunities only)
- Git identity (name/email)
- Private overlay initialization (machine-specific configs in a separate repo)

After first sync:
1. Restart shell: `exec $SHELL -l` (puts `dot` on PATH)
2. Open nvim to trigger plugin installation
3. Run `dot health` to verify everything
4. (Optional) Select iTerm2 "Dotfiles Default" profile for icons

## Day-to-Day Usage

### dot

```bash
# Keep dotfiles current
dot pull

# Check system health — or auto-fix issues
dot health
dot fix

# Show cleanup opportunities
dot clean
```

Full command reference: [cmd/dot/](cmd/dot/)

### folio

Folio manages knowledge work projects — plans, references, and compiled outputs. Interact through Claude Code skills:

- `/folio plan` — plan non-trivial tasks by exploring options, then converging on an approach
- `/folio compose` — turn research and references into polished outputs
- `/folio gather` — add sources to a project from URLs or deep research

CLI commands and project details: [cmd/folio/](cmd/folio/)

## What's Included

### Shell
- **Zsh** as primary shell with starship prompt
- **Bash** also configured
- `.shell_common` with shared aliases (`v`, `gs`, `reload`, etc.)

### Development Tools
- **mise** for language version management
- **neovim** with kickstart.nvim
- **lazygit** for git TUI
- **ripgrep**, **fd**, **fzf**, **bat** for search

### Terminal
- **Ghostty** as primary terminal (Inconsolata Nerd Font)
- **iTerm2** as fallback (with Nerd Font profile for icons)

## Neovim

Based on [kickstart.nvim](https://github.com/nvim-lua/kickstart.nvim) with:

| Plugin | Purpose | Key Binding |
|--------|---------|-------------|
| nvim-tree | File explorer | `<leader>e` (toggle), `<leader>f` (find file) |
| telescope | Fuzzy finder | `<leader>sf` (search files), `<leader>sg` (grep) |
| treesitter | Syntax highlighting | automatic |
| mason + lspconfig | LSP support | automatic |

## Claude Code Integration

This repo includes [Claude Code](https://docs.anthropic.com/en/docs/claude-code) skills available globally (in any project):

| Skill | Purpose |
|-------|---------|
| `/folio` | Knowledge work lifecycle - plan, compose, gather, publish |
| `/dotfiles` | Manage dotfiles - install, update, health, setup |
| `/nvim` | Neovim help - plugins, config, troubleshooting |
| `/commit` | Git commit conventions and versioning |
| `/stacked-pr` | Stacked branch workflows and propagation |

