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

**CLI**: `dot <command>` (symlinked to `~/.local/bin/dot`):
- `dot sync` - Apply dotfiles state (auto-detects first-time vs update)
- `dot pull` - Pull latest, then sync
- `dot links` - Symlinks only (no brew, no pull)
- `dot health` / `dot fix` - Diagnostics / auto-repair
- `dot clean` / `dot clean!` - Show / execute cleanup
- `dot setup` - Interactive first-time setup
- `dot status` - Show mode, profile, git state
- `dot edit` - Open dotfiles in editor

**Multi-mode**: `~/.dotfiles/.machine` contains `aggressive` or `conservative`
- `shared/` configs apply in all modes
- `aggressive/Brewfile` only in aggressive mode (personal machines, repo is source of truth)
- Conservative mode: minimal changes, just shows cleanup opportunities (work machines)

**Profile**: `~/.dotfiles/.profile` contains `work` or `personal`
- Selects which private overlay subdirectory to apply (`~/.dotfiles.private/work/` or `personal/`)
- Affects profile-specific Brewfiles, symlinks, and Claude skills

**Not managed by dotfiles**:
- `~/.claude/settings.json` - Managed externally (e.g. AWS auth wizards), not tracked in this repo

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
4. Run `dot links`

## First-Time Setup (`/dotfiles setup`)
Run `~/.dotfiles/health.sh --setup` for interactive walkthrough (use direct path before first sync; after sync + shell restart, `dot setup` works):
1. Shell reload
2. Neovim plugin installation
3. iTerm2 profile selection
4. Claude Code authentication
5. Git identity setup

## Troubleshooting
- **Git operations fail**: Check GitHub SSH with `ssh -T git@github.com` (must be set up before bootstrap)
- **Symlink conflicts**: `dot sync --dry-run` shows state
- **Broken links**: `dot fix` repairs them
- **Changes not applied**: Run `dot links`
- **First-time issues**: Run `dot setup` for guided help
