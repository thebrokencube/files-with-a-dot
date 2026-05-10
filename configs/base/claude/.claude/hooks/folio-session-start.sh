#!/bin/bash
# SessionStart hook: reap stale jj workspaces for ~/.folio.
# Workspace creation is lazy — handled by the /folio skill on first invocation.

# Only runs if ~/.folio has jj
if [ ! -d "$HOME/.folio/.jj" ]; then exit 0; fi

# Reap stale workspaces (>2 days old)
# Uses folio CLI which checks for unpushed changes before removing.
# BSD find (macOS): -mtime +2 means modified more than 2*24h ago
find /tmp -maxdepth 1 -name 'folio-ws-*' -type d -mtime +2 -exec sh -c '
  folio home workspace cleanup "$1" 2>/dev/null
' _ {} \;
