#!/bin/bash
# PostCompact can notify the terminal but cannot add model context.

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
DIRTY=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')

printf 'Context was compacted. Session state:\n- Branch: %s\n- Dirty files: %s\n- Re-read CLAUDE.md, ~/.claude/rules/, and any active skill references for current task context.\n' "$BRANCH" "$DIRTY" >&2
