---
name: dotfiles
description: Manage dotfiles - installation, updates, adding configs, troubleshooting symlinks. Use when asked about dotfiles, system config, or shell setup.
---

# Dotfiles Management

## Commands
- `/dotfiles` - Show status and help
- `/dotfiles sync` - Sync dotfiles (auto-detects first-time vs update)
- `/dotfiles health` - Run diagnostics
- `/dotfiles setup` - Interactive first-time setup guide
- `/dotfiles add <app>` - Add new config

## How the System Works

**Symlink-based**: `~/.dotfiles/symlink_map.txt` defines source → destination mappings.

```
shared/git/.gitconfig:$HOME/.gitconfig
shared/nvim/.config/nvim:$HOME/.config/nvim
shared/iterm2/dotfiles-profile.json:$HOME/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json
```

**Scripts**:
- `sync.sh` - Main command: apply dotfiles state (auto-detects first-time vs update)
- `health.sh` - Pure diagnostics, `--fix` to auto-repair
- `cleanup.sh` - Aggressive cleanup or show opportunities (conservative)
- `uninstall.sh` - Clean removal

**Multi-mode**: `~/.dotfiles/.machine` contains `aggressive` or `conservative`
- `shared/` configs apply in all modes
- `aggressive/Brewfile` only in aggressive mode (personal machines, repo is source of truth)
- Conservative mode: minimal changes, just shows cleanup opportunities (work machines)

**Profile**: `~/.dotfiles/.profile` contains `work` or `personal`
- Selects which private overlay subdirectory to apply (`~/.dotfiles.private/work/` or `personal/`)
- Affects profile-specific Brewfiles, symlinks, and Claude skills

**Local config tiering** (`~/.dotfiles/local/`):
- `shell.managed` - Aggressive mode only, repo-controlled, updated on every sync (has mise activation)
- `shell.local` - Never touched by sync, your private customizations
- `env.local` - API keys, secrets (never in git)
- `gitconfig.local` - Git identity (never in git)

## To Show Status (`/dotfiles`)
1. Read `~/.dotfiles/.machine` for machine type
2. Check key symlinks exist (from symlink_map.txt)
3. Show available commands

## To Add New Config (`/dotfiles add <app>`)
1. Create directory: `~/.dotfiles/shared/<app>/`
2. Add config files mirroring target structure
3. Add entry to `symlink_map.txt`: `shared/<app>/config:$HOME/.config/<app>/config`
4. Run `./sync.sh --links-only`

## First-Time Setup (`/dotfiles setup`)
Run `~/.dotfiles/health.sh --setup` for interactive walkthrough:
1. Shell reload
2. Neovim plugin installation
3. iTerm2 profile selection
4. Claude Code authentication
5. Git identity setup

## Troubleshooting
- **Git operations fail**: Check GitHub SSH with `ssh -T git@github.com` (must be set up before bootstrap)
- **Symlink conflicts**: `./sync.sh --dry-run` shows state
- **Broken links**: `./health.sh --fix` repairs them
- **Changes not applied**: Run `./sync.sh --links-only`
- **First-time issues**: Run `./health.sh --setup` for guided help
