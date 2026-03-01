# dot CLI

The `dot` command manages dotfiles — symlinks, Homebrew packages, private configs, and system health. Available on PATH after first sync (`~/.local/bin/dot`).

## CLI Reference

| Command | Description |
|---------|-------------|
| `dot sync` | Apply dotfiles state |
| `dot pull` | Pull latest, then sync |
| `dot links` | Symlinks only (no brew, no pull) |
| `dot health` | Run health diagnostics |
| `dot fix` | Health check with auto-fix |
| `dot clean` | Show cleanup opportunities |
| `dot clean --force` | Execute cleanup without prompts |
| `dot private` | Manage private overlay (init, status, sync, push, edit) |
| `dot status` | Show current state (mode, git) |
| `dot edit` | Open dotfiles in editor |
| `dot validate` | Validate repo structure |
| `dot help` | Show all commands |

Options: `--dry-run` and `--skip-brew` pass through to underlying scripts.

## How It Works

### Symlink Map

`symlink_map.txt` defines where each config links:

```
configs/base/git/.gitconfig:$HOME/.gitconfig
configs/base/nvim/.config/nvim:$HOME/.config/nvim
configs/base/iterm2/dotfiles-profile.json:$HOME/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json
cmd/dot/dot:$HOME/.local/bin/dot
```

This allows linking to any location — useful for apps like iTerm2 that store configs in `~/Library`.

### Multi-Mode Support

- **Base configs** (`configs/base/`) are linked in all modes
- **Aggressive mode** (`configs/aggressive/Brewfile`) — additional packages, aggressive cleanup (repo is source of truth)
- **Conservative mode** — minimal changes, show cleanup opportunities only (other tools may manage packages)

### Private Overlay

`~/.dotfiles.private/` holds machine-specific configs in a separate git repo:

```
~/.dotfiles.private/
├── symlink_map.txt     # Private symlinks
├── Brewfile            # Private Homebrew packages
├── skills/             # Private Claude Code skills
├── gitconfig.local     # Git name/email → ~/.gitconfig.local
├── env.local           # API keys, secrets → ~/.env.local
└── shell.local         # Private aliases/functions → ~/.shell.local
```

Initialized on first sync. Manage with `dot private` (init, status, sync, push, edit). Push to a private remote to sync across machines.

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
│   │   ├── ARCHITECTURE.md     # design decisions and internals
│   │   └── scripts/            # one-offs + helpers
│   │       ├── bootstrap.sh    # standalone first-time setup
│   │       ├── uninstall.sh    # standalone removal
│   │       └── validate*.sh    # dot validate
│   ├── md-to-adf               # Markdown → Atlassian Document Format
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
└── README.md
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for design decisions and internals.
