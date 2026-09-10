# Private Dotfiles Overlay

This is your private dotfiles overlay. It layers on top of your public dotfiles.

## Structure

```
~/.dotfiles.private/
├── symlink_map.txt     # Private symlinks
├── Brewfile            # Private Homebrew packages
├── skills/             # Private agent skills
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
1. Public `configs/base/`
2. Public `configs/aggressive/` (if aggressive mode)
3. Private overlay

## Adding Skills

Put private agent skills in `skills/`:
```
skills/my-tool/SKILL.md
```

Add these explicit rows to `symlink_map.txt` for each skill:

```text
skills/my-tool:$HOME/.claude/skills/my-tool
skills/my-tool:$HOME/.omp/agent/skills/my-tool
skills/my-tool:$HOME/.codex/skills/my-tool
```

They declare candidate map roots; they do not claim host support.
