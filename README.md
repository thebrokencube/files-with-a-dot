# files-with-a-dot

Personal dotfiles with symlink-based management, supporting multiple modes (aggressive/conservative) with shared configurations.

## Quick Start

**GitHub SSH must be configured** before running bootstrap. Test with:

```bash
ssh -T git@github.com
```

If this fails, [set up SSH keys first](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).

On a fresh macOS machine, run:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/thebrokencube/files-with-a-dot/main/cmd/dot/scripts/bootstrap.sh)"
```

This clones the repo to `~/.dotfiles`, installs dependencies via Homebrew, and runs sync to apply dotfiles state. First run prompts for:
- Mode — aggressive (repo-controlled packages + cleanup) or conservative (show opportunities only)
- Git identity (name/email)
- Private overlay initialization (machine-specific configs in a separate repo)

**Note:** If Homebrew can't install (sudo issues, etc.), the script clones the repo and shows how to complete setup manually.

After first sync:
1. Restart shell: `exec $SHELL -l` (puts `dot` on PATH)
2. Open nvim to trigger plugin installation
3. Run `dot health` to verify everything
4. (Optional) Select iTerm2 "Dotfiles Default" profile for icons

## Plugin marketplace

This repo is also a cross-harness plugin marketplace for its CLI tools (`folio`, `jf`, `dendrik`).
In Claude Code:

```
/plugin marketplace add thebrokencube/files-with-a-dot
/plugin install folio@files-with-a-dot   # or jf, dendrik
```

Then invoke the tool (e.g. `/folio`) — its skill installs the binary on first use via the
plugin's bundled, self-locating `bin/setup` (idempotent; safe to re-run). No repo path needed.
See [AGENTS.md](AGENTS.md) for the full model (`plugins.json` is canonical; per-harness catalogs
are generated).

## Day-to-Day

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

## What's Included

### Shell
- **Zsh** as primary shell with starship prompt
- **Bash** also configured
- `.shell_common` with shared aliases (`v`, `gs`, `reload`, etc.)

### Development Tools
- **mise** for language version management
- **jj** (Jujutsu) as primary VCS, colocated with git
- **ripgrep**, **fd**, **fzf**, **bat** for search

### Neovim

Based on [kickstart.nvim](https://github.com/nvim-lua/kickstart.nvim) with:

| Plugin | Purpose | Key Binding |
|--------|---------|-------------|
| nvim-tree | File explorer | `<leader>e` (toggle), `<leader>f` (find file) |
| telescope | Fuzzy finder | `<leader>sf` (search files), `<leader>sg` (grep) |
| treesitter | Syntax highlighting | automatic |
| mason + lspconfig | LSP support | automatic |

### Terminal
- **Ghostty** as primary terminal (Inconsolata Nerd Font)
- **iTerm2** as fallback (with Nerd Font profile for icons)

### Claude Code Skills

Available globally (in any project) via [Claude Code](https://docs.anthropic.com/en/docs/claude-code):

| Skill | Purpose |
|-------|---------|
| `/folio` | Knowledge work lifecycle - plan, compose, gather, lint, publish |
| `/jf` | Jira Forest - push/pull/sync markdown to Jira tickets |
| `/dotfiles` | Manage dotfiles - install, update, health, setup |
| `/nvim` | Neovim help - plugins, config, troubleshooting |
| `/commit` | Git commit conventions and versioning |
| `/dendrik` | Validate tool conventions and contract compliance |

## Tools

### dot

Dotfiles management CLI — sync, health checks, cleanup. See [cmd/dot/](cmd/dot/).

### folio

Knowledge work project manager — plans, references, and compiled outputs. Interact through Claude Code skills:

- `/folio gather` → `/folio plan` → `/folio compose` → `/folio publish`
- Design decisions freeze before implementation begins (lock gate in plan phase)

CLI commands and project details: [cmd/folio/](cmd/folio/)

### jf

Jira Forest CLI — maps a local filesystem of markdown files to Jira ticket descriptions. Supports ad-hoc single-file push/pull or full forest management with bidirectional sync, clone, and content lint/roundtrip checking.

- `jf clone KEY` → `jf sync` → `jf push` / `jf pull`
- Derived sync mode: direction inferred from content mutability (lint + roundtrip)

CLI commands and architecture: [cmd/jf/](cmd/jf/)

### dendrik

Convention linter for dotfiles CLI tools — validates Go, Skill, and Bridge layer contracts. See [cmd/dendrik/](cmd/dendrik/)
