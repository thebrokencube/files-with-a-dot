---
name: system
description: System troubleshooting - icons not rendering, shell issues, font problems, terminal setup. Use when asked about icons, fonts, terminal issues, or general system problems.
---

# System Troubleshooting

## Commands
- `/system` - Show system overview
- `/system icons` - Test icon rendering
- `/system fonts` - Check font setup
- `/system shell` - Diagnose shell issues

## Icon Rendering Issues

**Test icons**:
```bash
echo -e "\uf015 \ue0a0 \uf121 \uf07b"
```
Should show: home, git-branch, code, folder icons.

**If icons are broken**:

1. **Check font**: Terminal must use Nerd Font
   - Ghostty: Set in `~/.config/ghostty/config`
   - iTerm2: Use "Dotfiles Default" profile

2. **iTerm2 specific**:
   - Profiles → select "Dotfiles Default"
   - Or: Settings → Profiles → Text → set font to "Inconsolata Nerd Font"
   - Disable "Draw Powerline Glyphs" (uses built-in instead of font)

3. **Font not installed**:
   ```bash
   brew install --cask font-inconsolata-nerd-font
   ```

## Shell Issues

**Changes not taking effect**:
```bash
exec $SHELL -l   # or use 'reload' alias
```

**Check shell config is sourced**:
```bash
grep "dotfiles" ~/.zshrc   # should show source line
echo $EDITOR               # should be 'nvim'
```

**Environment variables missing**:
- Check `~/.env.local` exists and has values
- Check `~/.shell_common` is sourced

## Quick Diagnostics
```bash
~/.dotfiles/health.sh        # Full check
~/.dotfiles/health.sh --fix  # Auto-repair
```

## Terminal Setup

**Ghostty** (primary):
- Config: `~/.config/ghostty/config`
- Font already configured

**iTerm2** (fallback):
- Dynamic Profile at: `~/Library/Application Support/iTerm2/DynamicProfiles/dotfiles-profile.json`
- Must select "Dotfiles Default" profile to use it
