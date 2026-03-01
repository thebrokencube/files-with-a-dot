# Private Dotfiles Overlay

This is your private dotfiles overlay. It layers on top of your public dotfiles.

## Structure

```
~/.dotfiles.private/
├── symlink_map.txt     # Private symlinks
├── Brewfile            # Private Homebrew packages
├── skills/             # Private Claude Code skills
│   └── my-skill/
│       └── SKILL.md
└── ...                 # Any other private configs
```

## Usage

1. Add your private configs, symlinks, Brewfile entries, or skills
2. Run sync to apply: `dot sync`
3. Or apply just private symlinks: `dot private sync`

## Merge Order

Configs are applied in this order (later overrides earlier):
1. Public `shared/`
2. Public `aggressive/` (if aggressive mode)
3. Private overlay

## Adding Skills

Put Claude skills in `skills/`:
```
skills/my-tool/SKILL.md
```

They'll be symlinked to `~/.claude/skills/` on sync.
