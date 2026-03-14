#!/bin/bash
# Block command chaining (&&, ||, ;) in Bash tool calls.
# Used as a Claude Code PreToolUse hook on Bash.
#
# Why: chained commands bypass both the permission allow-list
# (prefix-matched) and PreToolUse hooks on individual commands.

INPUT=$(cat /dev/stdin)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

# Check for &&, ||, or ; anywhere in the command
if echo "$COMMAND" | grep -qE '&&|\|\||;'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Don'\''t chain commands with &&, ||, or ; in a single Bash call.\n\nChained commands bypass the permission allow-list and PreToolUse hooks.\nUse separate parallel tool calls for independent commands,\nor separate sequential tool calls for dependent ones."
    }
  }'
fi
