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

## Adding private skills

Put private Agent Skills in `skills/`:
```
skills/my-tool/SKILL.md
```

The migration appends, once, these three explicit map rows for every private skill:

```text
skills/my-tool:$HOME/.claude/skills/my-tool
skills/my-tool:$HOME/.omp/agent/skills/my-tool
skills/my-tool:$HOME/.codex/skills/my-tool
```

They are private-source projections to admitted user roots. `dot links` applies public links,
private skills, allowlisted retired-link cleanup, and the append-only private-skill map migration;
it does not migrate legacy/private non-skill state. Full `dot sync` and `dot private sync` use the
same skill-map migration. Collision checks always run before any symlink mutation.

## Private OMP MCP configuration

An OMP-native `mcp.json` may be mapped privately to `$HOME/.omp/agent/mcp.json`. Treat it as a
private-overlay migration: retain definitions, credentials, and server names only in the private
repository. The public map and documentation must never copy them.
