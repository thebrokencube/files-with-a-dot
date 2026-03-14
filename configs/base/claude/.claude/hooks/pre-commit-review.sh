#!/bin/bash
# Pre-commit review gate: shows staged diff and requires user approval.
# Used as a Claude Code PreToolUse hook on Bash(git commit*).
#
# Outputs JSON with permissionDecision: "ask" so Claude Code pauses
# for human confirmation before the commit proceeds.

# Read tool input from stdin (Claude Code passes it as JSON)
cat /dev/stdin > /dev/null

# Get the staged diff summary
DIFF=$(git diff --cached --stat 2>/dev/null)
if [ -z "$DIFF" ]; then
  DIFF="(no staged changes)"
fi

# Use jq to safely construct JSON with proper escaping
jq -n --arg diff "$DIFF" '{
  hookSpecificOutput: {
    hookEventName: "PreToolUse",
    permissionDecision: "ask",
    permissionDecisionReason: ("Staged changes:\n" + $diff)
  }
}'
