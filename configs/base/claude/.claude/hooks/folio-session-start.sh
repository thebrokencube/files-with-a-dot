#!/bin/bash
# SessionStart hook: create a jj workspace for ~/.folio per Claude session.
# Sets FOLIO_HOME via CLAUDE_ENV_FILE so it persists across all bash calls.

# Only runs if ~/.folio has jj
if [ ! -d "$HOME/.folio/.jj" ]; then exit 0; fi

# Generate session workspace ID
WS_ID="folio-ws-$(date +%s)-$$"
WS_DIR="/tmp/$WS_ID"

# Reap stale workspaces (>2 days old)
# BSD find (macOS): -mtime +2 means modified more than 2*24h ago
find /tmp -maxdepth 1 -name 'folio-ws-*' -type d -mtime +2 -exec sh -c '
  NAME=$(basename "$1")
  jj workspace forget "$NAME" -R "$HOME/.folio" 2>/dev/null
  rm -rf "$1"
' _ {} \;

# Create workspace
jj workspace add "$WS_DIR" -r main -R "$HOME/.folio" 2>/dev/null

# Set FOLIO_HOME for the session
if [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export FOLIO_HOME=\"$WS_DIR\"" >> "$CLAUDE_ENV_FILE"
fi

echo "folio workspace: $WS_DIR"
