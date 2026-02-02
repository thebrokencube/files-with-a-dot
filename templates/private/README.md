# Private Dotfiles Overlay

This is your private dotfiles overlay. It layers on top of your public dotfiles.

## Structure

```
~/.dotfiles.private/
├── shared/                 # Applies to all profiles
│   └── shell.local         # Private shell aliases/functions
├── work/                   # Work profile only
│   ├── symlink_map.txt     # Work-specific symlinks
│   ├── Brewfile            # Work-specific packages
│   └── skills/             # Work Claude skills
├── personal/               # Personal profile only
│   ├── symlink_map.txt     # Personal-specific symlinks
│   ├── Brewfile            # Personal-specific packages
│   └── skills/             # Personal Claude skills
├── symlink_map.txt         # Shared private symlinks
└── Brewfile                # Shared private packages
```

## Usage

1. Set your profile during sync:
   ```bash
   ~/.dotfiles/sync.sh --profile work
   ```

2. Add profile-specific configs to `work/` or `personal/`

3. Run sync to apply changes:
   ```bash
   dfu  # or ~/.dotfiles/sync.sh
   ```

## Merge Order

Configs are applied in this order (later overrides earlier):
1. Public `shared/`
2. Public `aggressive/` (if aggressive mode)
3. Private `shared/`
4. Private `{profile}/` (work or personal)

## Adding Skills

Put Claude skills in `{profile}/skills/`:
```
work/skills/my-work-tool/SKILL.md
```

They'll be symlinked to `~/.claude/skills/`.
