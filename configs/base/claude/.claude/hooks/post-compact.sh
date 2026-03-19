#!/bin/bash
# PostCompact hook: restore orientation after context compaction.
# Outputs key session state so the model can re-orient.

INPUT=$(cat /dev/stdin)

BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
DIRTY=$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')

jq -n \
  --arg branch "$BRANCH" \
  --arg dirty "$DIRTY" \
  '{
    hookSpecificOutput: {
      hookEventName: "PostCompact",
      suppressOutput: false,
      output: ("Context was compacted. Session state:\n- Branch: " + $branch + "\n- Dirty files: " + $dirty + "\n- Re-read CLAUDE.md and any active skill references for current task context.")
    }
  }'
