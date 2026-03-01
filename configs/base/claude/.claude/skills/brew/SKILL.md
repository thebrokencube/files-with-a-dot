---
name: brew
description: Homebrew package management - installing, updating, listing packages. Use when asked about brew, packages, or installing tools.
---

# Homebrew Management

## Commands
- `/brew` - Show status
- `/brew status` - List outdated packages
- `/brew update` - Update and upgrade all
- `/brew add <package>` - Add to Brewfile
- `/brew cleanup` - Remove old versions

## Brewfiles
| File | Purpose | When Installed |
|------|---------|----------------|
| `~/.dotfiles/Brewfile.shared` | Core dev tools | All machines |
| `~/.dotfiles/configs/aggressive/Brewfile` | Extra apps | Aggressive mode only |
| `~/.dotfiles.private/Brewfile` | Private packages | When private overlay exists |

## To Show Status (`/brew` or `/brew status`)
```bash
brew outdated
```

## To Update (`/brew update`)
```bash
brew update && brew upgrade
```

## To Add Package (`/brew add <package>`)
1. Determine if shared or home-only
2. Add to appropriate Brewfile:
   - Formula: `brew "package-name"`
   - Cask: `cask "app-name"`
3. Run: `brew bundle --file=<brewfile>`

## To Cleanup (`/brew cleanup`)
**Note:** Cleanup is mainly intended for home machines. On work machines, other tools may manage packages.

```bash
brew cleanup
```

## Current Core Tools (Brewfile.shared)
- git, neovim, mise
- zsh, starship
- ripgrep, fd, fzf, bat
- lazygit, gh, jq
- ghostty, font-inconsolata-nerd-font
