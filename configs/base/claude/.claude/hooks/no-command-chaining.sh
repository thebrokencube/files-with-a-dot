#!/bin/bash
# Block command patterns that bypass the permission allow-list or PreToolUse hooks.
# Used as a Claude Code PreToolUse hook on Bash.
#
# Blocks:
# 1. Command chaining (&&, ||, ;) — each command needs its own tool call
# 2. Command substitution ($(...), `...`) — executes arbitrary commands inside an allowed prefix
# 3. git -C / --git-dir / --work-tree — changes which repo git operates on

INPUT=$(cat /dev/stdin)
COMMAND=$(echo "$INPUT" | jq -r '.tool_input.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

# Block command chaining
if echo "$COMMAND" | grep -qE '&&|\|\||;'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Don'\''t chain commands with &&, ||, or ; in a single Bash call.\n\nChained commands bypass the permission allow-list and PreToolUse hooks.\nUse separate parallel tool calls for independent commands,\nor separate sequential tool calls for dependent ones."
    }
  }'
  exit 0
fi

# Block command substitution ($(...) and backticks)
if echo "$COMMAND" | grep -qE '\$\(|`'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Don'\''t use command substitution ($(...) or backticks).\n\nCommand substitution executes arbitrary commands inside an allowed prefix,\nbypassing permission matching. Use separate tool calls instead."
    }
  }'
  exit 0
fi


# Block git -C (bypasses all Bash(git ...*) prefix matching in allow list and hooks)
if echo "$COMMAND" | grep -qE '^git\s+-C\s'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Don'\''t use git -C.\n\ngit -C <path> <cmd> bypasses Bash(git <cmd>*) prefix matching in the allow list and hooks.\nRun git commands from the repo working directory instead."
    }
  }'
  exit 0
fi

# Block git --git-dir / --work-tree (same class as git -C)
if echo "$COMMAND" | grep -qE '^git\s+--?(git-dir|work-tree)\b'; then
  jq -n '{
    hookSpecificOutput: {
      hookEventName: "PreToolUse",
      permissionDecision: "deny",
      permissionDecisionReason: "Don'\''t use git --git-dir or --work-tree.\n\nThese flags change which repo git operates on, bypassing permission matching.\nRun git commands from the repo working directory instead."
    }
  }'
  exit 0
fi
