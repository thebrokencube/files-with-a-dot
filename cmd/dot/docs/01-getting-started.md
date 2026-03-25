# Getting Started with dot

## What dot Does

`dot` manages the dotfiles environment -- symlinks, Homebrew packages, private configs, shell integration, and system health. It's the entry point for installing, updating, and troubleshooting the dotfiles repo on any macOS machine.

## Bootstrap from Zero

On a fresh machine with nothing installed:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/cmd/dot/scripts/bootstrap.sh)"
```

This self-contained script:
1. Installs Xcode Command Line Tools (if missing)
2. Clones the repo to `~/.dotfiles`
3. Checks for Homebrew (offers to install if missing)
4. Runs `dot sync` to apply everything

For non-interactive setup (CI or scripted installs):

```bash
./bootstrap.sh --machine conservative --git-name "Name" --git-email "email@example.com"
```

## First Sync

After bootstrap, or when pulling changes:

```bash
dot sync
```

Sync applies the full dotfiles state:
- Creates symlinks from `symlink_map.txt`
- Merges base + private configs from `managed_map.txt`
- Installs Homebrew packages from `Brewfile.shared` (+ mode-specific Brewfile)
- Integrates shell configs (sources `.shell_common` from your shell rc)
- Runs cleanup based on your machine mode

Use `dot sync --dry-run` to preview without making changes.

## Understanding Machine Modes

Your machine runs in one of two modes, stored in `~/.dotfiles/.machine`:

| Mode | Behavior |
|------|----------|
| **aggressive** | Full package management -- dot is source of truth for Homebrew. Cleanup removes packages not in Brewfiles. |
| **conservative** | Minimal changes -- shows cleanup opportunities but doesn't remove packages. Safe when other tools manage packages. |

First sync prompts you to choose. Change later by editing `.machine`.

## Setting Up Private Configs

Private configs (API keys, git identity, machine-specific settings) live in a separate repo:

```bash
dot private init    # scaffold ~/.dotfiles.private
dot private edit    # open in editor
dot private push    # commit and push to remote
dot private status  # show git state
```

Key private files:
- `env.local` -- API keys and secrets (sourced as `~/.env.local`)
- `gitconfig.local` -- git name/email (included by `~/.gitconfig`)
- `shell.local` -- private aliases and functions
- `symlink_map.txt` -- private symlinks (same format as the main one)
- `Brewfile` -- private Homebrew packages

Push to a private remote to sync across machines.

## Daily Commands

```bash
dot pull          # git pull + sync (the daily update)
dot health        # check system state, find issues
dot fix           # health check with auto-fix
dot status        # show mode, git state
dot clean         # show cleanup opportunities
dot validate      # check repo structure integrity
```

## Multi-Machine Workflow

To set up a second machine:

1. Run bootstrap (clones the main repo)
2. `dot private init` -- scaffold private overlay
3. Set the private remote: `cd ~/.dotfiles.private && git remote add origin <url>`
4. `dot private pull` -- pull private configs from remote
5. `dot sync` -- apply everything

After that, keeping machines in sync:
- Main dotfiles: `dot pull` pulls from the shared repo
- Private configs: `dot private pull` / `dot private push`

## Adding New Configs

1. Create the config directory: `configs/base/<app>/`
2. Add a symlink entry: `configs/base/<app>/config:$HOME/.config/<app>` in `symlink_map.txt`
3. If it needs a Homebrew package: add to `Brewfile.shared`
4. Apply: `dot links` (symlinks only, no brew)

## Troubleshooting

`dot health` is the first thing to run when something seems wrong. It checks symlink state, shell integration, Homebrew packages, and private overlay status.

Common issues:

| Issue | Solution |
|-------|----------|
| Symlink conflicts | `dot sync --dry-run` to see state |
| Shell changes not applied | `exec $SHELL -l` or `reload` |
| Icons broken in terminal | Select correct profile, check Nerd Font |
| Git operations fail | Check SSH: `ssh -T git@github.com` |
| Nvim plugins missing | Open nvim to trigger auto-install |

## What's Next

- [04-architecture.md](04-architecture.md) -- design decisions, FCIS pattern, script roles, lib/ responsibilities
- [README.md](../README.md) -- CLI reference and repository structure
