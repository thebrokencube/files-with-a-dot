---
name: dotfiles-dev
description: Guidelines for developing and maintaining the dotfiles repo. Use when making changes to dotfiles, adding new configs, updating scripts, or modifying skills.
user-invocable: false
---

# Dotfiles Development Guidelines

This skill provides guidelines for Claude when helping maintain the dotfiles repo.

## Repository Structure

```
~/.dotfiles/
├── shared/              # Configs for ALL machines
│   ├── <app>/           # One directory per app
│   └── claude/.claude/  # Claude Code config + skills
├── aggressive/          # Aggressive mode only
│   └── Brewfile         # Personal/extra apps
├── symlink_map.txt      # Defines all symlinks
├── sync.sh              # Main: apply dotfiles state
├── health.sh            # Pure diagnostics
├── cleanup.sh           # Aggressive cleanup or opportunities (conservative)
├── uninstall.sh         # Clean removal
└── Brewfile.shared      # Core tools
```

## Adding a New Config

1. **Create directory**: `shared/<app>/` mirroring target structure
2. **Add to symlink_map.txt**: `shared/<app>/config:$HOME/.config/<app>/config`
3. **If needs brew package**: Add to `Brewfile.shared` or `aggressive/Brewfile`
4. **Test**: Run `./sync.sh --links-only`

Example for adding lazydocker:
```
mkdir -p shared/lazydocker/.config/lazydocker
# Add config file
echo "shared/lazydocker/.config/lazydocker:$HOME/.config/lazydocker" >> symlink_map.txt
```

## Modifying Scripts

**Patterns to follow:**
- Use `--dry-run` flag for previewing changes
- Use colored output: `ok()`, `warn()`, `err()`, `info()`
- Detect state before acting (idempotent)
- Support both interactive and non-interactive modes

**When adding new checks to health.sh:**
1. Add check function: `check_<name>()`
2. Add to `SPECIFIC_CHECK` case statement
3. If manual action needed, add to `check_setup_status()`
4. If walkthrough needed, add step to `run_interactive_setup()`

## Adding/Modifying Skills

**Skill location**: `shared/claude/.claude/skills/<name>/SKILL.md`

**Required frontmatter:**
```yaml
---
name: skill-name
description: When to use this skill (for Claude's context)
---
```

**Optional frontmatter:**
- `disable-model-invocation: true` - User-only command
- `user-invocable: false` - Claude-only (background knowledge)
- `allowed-tools: Read, Grep` - Restrict tools

**Keep skills focused**: One topic per skill, detailed docs in supporting files.

## Commit Conventions

When committing changes to this repo:
- Amend to current version tag (e.g., `v0.2`) during active development
- Use descriptive commit messages
- Test with `./sync.sh --dry-run` before committing

## Testing Changes

1. **Quick test**: `./health.sh`
2. **Full test**: `./uninstall.sh && ./sync.sh --dry-run`
3. **Interactive test**: `./health.sh --setup`

## Multi-Mode Considerations

- `shared/` = ALL modes
- `aggressive/` = Aggressive mode only (check `$MACHINE_TYPE`)
- Conservative mode: minimal changes, stays local
- Claude Code auth: Required for both modes

## Updating This Skill

When patterns change, update this skill so future Claude sessions follow the new patterns.
