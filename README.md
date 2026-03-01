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

This will:
1. Check for Xcode Command Line Tools (includes git on macOS)
2. Clone this repo to `~/.dotfiles`
3. Check for Homebrew (offers to install if missing)
4. Run sync to apply dotfiles state (if Homebrew available)
5. Prompt for machine type and git identity on first run

**Note:** If Homebrew installation can't complete (sudo issues, etc.), the script will clone the repo successfully and show you how to complete setup manually by running `~/.dotfiles/cmd/dot/sync.sh` after installing Homebrew.

## Repository Structure

```
files-with-a-dot/
├── cmd/
│   ├── dot/                    # dot CLI + engine
│   │   ├── dot                 # entrypoint (symlinked to ~/.local/bin/dot)
│   │   ├── sync.sh             # dot sync/pull/links
│   │   ├── cleanup.sh          # dot clean
│   │   ├── health.sh           # dot health/fix
│   │   ├── lib/                # sourced shell functions (11 files)
│   │   └── scripts/            # one-offs + helpers
│   │       ├── bootstrap.sh    # standalone first-time setup
│   │       ├── uninstall.sh    # standalone removal
│   │       └── validate*.sh    # dot validate
│   ├── md-to-adf               # Markdown -> ADF converter
│   └── folio/                  # Go CLI
├── configs/
│   ├── base/                   # always-applied configs
│   │   ├── nvim/.config/nvim/  # Neovim (kickstart.nvim)
│   │   ├── ghostty/            # Ghostty terminal
│   │   ├── starship/           # Starship prompt
│   │   ├── bash/               # Bash config
│   │   ├── zsh/                # Zsh config (primary shell)
│   │   ├── git/                # Git config
│   │   ├── shell/              # Shared shell config (.shell_common)
│   │   ├── iterm2/             # iTerm2 profile (Nerd Font)
│   │   └── claude/.claude/     # Claude Code config + skills
│   ├── aggressive/             # Aggressive mode overlay (Brewfile + shell.managed)
│   └── templates/              # Private overlay scaffold
├── symlink_map.txt             # Defines where each config links
├── managed_map.txt             # Base + overlay merge rules
├── Brewfile.shared             # Core dev tools
├── ARCHITECTURE.md
└── README.md
```

## How It Works

### Symlink Map

The `symlink_map.txt` file defines where each dotfile should be linked:

```
configs/base/git/.gitconfig:$HOME/.gitconfig
configs/base/nvim/.config/nvim:$HOME/.config/nvim
configs/base/iterm2/dotfiles-profile.json:$HOME/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json
cmd/dot/dot:$HOME/.local/bin/dot
```

This allows linking to any location - useful for apps like iTerm2 that store configs in `~/Library`.

### Multi-Mode Support

- **Base configs** (`configs/base/`) are linked in all modes
- **Aggressive mode** (`configs/aggressive/Brewfile`) - additional packages, aggressive cleanup (repo is source of truth)
- **Conservative mode** - minimal changes, show cleanup opportunities only (other tools may manage packages)

### Private Overlay (`~/.dotfiles.private/`)

Machine-specific configs live in a separate private overlay directory (not committed to this repo):

```
~/.dotfiles.private/
├── symlink_map.txt     # Private symlinks
├── Brewfile            # Private Homebrew packages
├── skills/             # Private Claude Code skills
├── gitconfig.local     # Git name/email → symlinked to ~/.gitconfig.local
├── env.local           # API keys, secrets → symlinked to ~/.env.local
└── shell.local         # Private aliases/functions → symlinked to ~/.shell.local
```

First-time sync offers to initialize this automatically. Manage with `dot private` (init, status, sync, push, edit).

The private overlay is its own git repo — push to a private remote to sync across machines.

**Aggressive mode** also sources `configs/aggressive/shell.managed` (repo-controlled, includes mise activation).

## Scripts

All scripts live under `cmd/dot/`:

| Script | Purpose |
|--------|---------|
| `cmd/dot/scripts/bootstrap.sh` | First-time setup on a fresh machine |
| `cmd/dot/sync.sh` | Main command: apply dotfiles state (detects first-time vs. update) |
| `cmd/dot/health.sh` | Pure diagnostics - check system state |
| `cmd/dot/cleanup.sh` | Cleanup packages and files (aggressive vs conservative mode) |
| `cmd/dot/scripts/uninstall.sh` | Remove symlinks and optionally local config |

### The `dot` CLI

After sync, the `dot` command is available on PATH (`~/.local/bin/dot`):

```bash
dot sync      # Apply dotfiles state
dot pull      # Pull latest, then sync
dot links     # Symlinks only (no brew, no pull)
dot health    # Run health diagnostics
dot fix       # Health check with auto-fix
dot clean     # Show cleanup opportunities
dot clean --force  # Execute cleanup without prompts
dot private   # Manage private overlay (init, status, sync, push, edit)
dot status    # Show current state (mode, git)
dot edit      # Open dotfiles in editor
dot help      # Show all commands
```

Options like `--dry-run` and `--skip-brew` pass through to the underlying scripts.

## First-Time Setup

The `sync.sh` script automatically detects first-time setup and prompts for:
- Mode (aggressive/conservative)
- Git identity (name/email)
- Private overlay initialization

After first sync:
1. Restart shell: `exec $SHELL -l` (this puts `dot` on PATH)
2. Open nvim to trigger plugin installation
3. Run `dot health` to verify everything
4. (Optional) Select iTerm2 "Dotfiles Default" profile for icons

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

This repo includes Claude Code skills that are available globally (in any project):

| Skill | Purpose |
|-------|---------|
| `/dotfiles` | Manage dotfiles - install, update, health, setup |
| `/nvim` | Neovim help - plugins, config, troubleshooting |
| `/brew` | Homebrew management - status, update, cleanup |
| `/system` | System troubleshooting - icons, fonts, shell |

### How Skills Work

- `~/.claude/CLAUDE.md` - Minimal context (always loaded)
- `~/.claude/skills/*/SKILL.md` - Full instructions (loaded on invoke)

This keeps context small while providing detailed help when needed.

## Updating

```bash
dot pull
```

This pulls the latest dotfiles, re-creates symlinks, and updates Homebrew packages.

## Adding New Configs

1. Create directory: `configs/base/<app>/`
2. Add to `symlink_map.txt`: `configs/base/<app>/config:$HOME/.config/<app>`
3. If needs brew package: Add to `Brewfile.shared`
4. Run: `dot links`

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Git operations fail | Set up GitHub SSH: `ssh -T git@github.com` or check [GitHub docs](https://docs.github.com/en/authentication/connecting-to-github-with-ssh) |
| Icons broken in iTerm2 | Select "Dotfiles Default" profile, disable "Draw Powerline Glyphs" |
| Shell changes not applied | Run `exec $SHELL -l` or `reload` |
| Symlink conflicts | Run `dot sync --dry-run` to see state |
| Nvim plugins missing | Open nvim to trigger auto-install |

Run `dot health` for diagnostics.
