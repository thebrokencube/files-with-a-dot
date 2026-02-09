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
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/bootstrap.sh)"
```

This will:
1. Check for Xcode Command Line Tools (includes git on macOS)
2. Clone this repo to `~/.dotfiles`
3. Check for Homebrew (offers to install if missing)
4. Run sync to apply dotfiles state (if Homebrew available)
5. Prompt for machine type and git identity on first run

**Note:** If Homebrew installation can't complete (sudo issues, etc.), the script will clone the repo successfully and show you how to complete setup manually by running `~/.dotfiles/sync.sh` after installing Homebrew.

## Repository Structure

```
files-with-a-dot/
├── shared/                    # Configs for all machines
│   ├── nvim/.config/nvim/     # Neovim (kickstart.nvim)
│   ├── ghostty/.config/ghostty/  # Ghostty terminal
│   ├── starship/.config/      # Starship prompt
│   ├── bash/                  # Bash config
│   ├── zsh/                   # Zsh config (primary shell)
│   ├── git/                   # Git config
│   ├── shell/                 # Shared shell config (.shell_common)
│   ├── iterm2/                # iTerm2 profile (Nerd Font)
│   └── claude/.claude/        # Claude Code config + skills
│       ├── CLAUDE.md          # Global context (minimal)
│       └── skills/            # Slash commands
├── aggressive/                # Aggressive mode only
│   └── Brewfile               # Personal/extra apps
├── symlink_map.txt            # Defines where each config links
├── sync.sh                    # Main command: apply dotfiles state
├── health.sh                  # Pure diagnostics
├── cleanup.sh                 # Aggressive cleanup or show opportunities (conservative)
├── uninstall.sh               # Clean removal
└── Brewfile.shared            # Core dev tools
```

## How It Works

### Symlink Map

The `symlink_map.txt` file defines where each dotfile should be linked:

```
shared/git/.gitconfig:$HOME/.gitconfig
shared/nvim/.config/nvim:$HOME/.config/nvim
shared/iterm2/dotfiles-profile.json:$HOME/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json
shared/claude/.claude:$HOME/.claude
```

This allows linking to any location - useful for apps like iTerm2 that store configs in `~/Library`.

### Multi-Mode Support

- **Shared configs** (`shared/`) are linked in all modes
- **Aggressive mode** (`aggressive/Brewfile`) - additional packages, aggressive cleanup (repo is source of truth)
- **Conservative mode** - minimal changes, show cleanup opportunities only (other tools may manage packages)

### Local Configuration (`local/`)

Machine-specific configs in `~/.dotfiles/local/` (not committed to git):

| File | Purpose | Updated by sync? |
|------|---------|------------------|
| `shell.managed` | Repo-controlled shell config (aggressive mode only) | ✅ Yes - always |
| `shell.local` | Your private aliases/functions | ❌ Never touched |
| `env.local` | API keys, secrets | ❌ Never touched |
| `gitconfig.local` | Git name/email | ❌ Never touched |

**Aggressive mode**: `shell.managed` is copied from repo on every sync, includes mise activation enabled by default. Override in `shell.local` if needed.

**Conservative mode**: Only `shell.local` exists, fully manual configuration.

## Scripts

| Script | Purpose |
|--------|---------|
| `bootstrap.sh` | First-time setup on a fresh machine |
| `sync.sh` | Main command: apply dotfiles state (detects first-time vs. update) |
| `health.sh` | Pure diagnostics - check system state |
| `cleanup.sh` | Cleanup packages and files (aggressive vs conservative mode) |
| `uninstall.sh` | Remove symlinks and optionally local config |

### Common Commands

```bash
./sync.sh                # Sync dotfiles (auto-detects mode)
./sync.sh --pull         # Force pull before sync
./sync.sh --dry-run      # Preview changes without applying
./sync.sh --links-only   # Only re-create symlinks
./health.sh              # Check system health (read-only)
./cleanup.sh             # Show cleanup opportunities
./cleanup.sh --execute   # Actually clean (with confirmation)
```

### Aliases

After running sync, these aliases are available:

```bash
dfs      # Sync dotfiles
dfsp     # Sync with pull
dfl      # Links only
dfh      # Health check
dfc      # Cleanup (show)
dfce     # Cleanup (execute)
```

## First-Time Setup

The `sync.sh` script automatically detects first-time setup and prompts for:
- Mode (aggressive/conservative)
- Git identity (name/email)

After first sync:
1. Restart shell: `exec $SHELL -l`
2. Open nvim to trigger plugin installation
3. Run `./health.sh` to verify everything
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
| nvim-tree | File explorer | `<leader>e` |
| telescope | Fuzzy finder | `<leader>sf`, `<leader>sg` |
| lazygit | Git TUI | `<leader>gg` |
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
./sync.sh --pull
# or just: dfsp
```

This pulls the latest dotfiles, re-creates symlinks, and updates Homebrew packages.

## Adding New Configs

1. Create directory: `shared/<app>/`
2. Add to `symlink_map.txt`: `shared/<app>/config:$HOME/.config/<app>`
3. If needs brew package: Add to `Brewfile.shared`
4. Run: `./sync.sh --links-only` or `dfl`

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Git operations fail | Set up GitHub SSH: `ssh -T git@github.com` or check [GitHub docs](https://docs.github.com/en/authentication/connecting-to-github-with-ssh) |
| Icons broken in iTerm2 | Select "Dotfiles Default" profile, disable "Draw Powerline Glyphs" |
| Shell changes not applied | Run `exec $SHELL -l` or `reload` |
| Symlink conflicts | Run `./sync.sh --dry-run` to see state |
| Nvim plugins missing | Open nvim to trigger auto-install |

Run `./health.sh` for diagnostics.
