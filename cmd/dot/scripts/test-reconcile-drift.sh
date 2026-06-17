#!/bin/bash
# test-reconcile-drift.sh - Isolated test for reconcile_plugin_drift (lib/private.sh).
# Seeds temp DOTFILES_DIR / PRIVATE_DIR / dest with disk-ahead drift and asserts that
# only the allowlisted plugin manifests are backfilled from disk into the overlay source.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"   # cmd/dot

# shellcheck source=../lib/colors.sh
source "$DOT_DIR/lib/colors.sh"
# shellcheck source=../lib/logging.sh
source "$DOT_DIR/lib/logging.sh"
# shellcheck source=../lib/private.sh
source "$DOT_DIR/lib/private.sh"

FAIL=0
assert_eq() {
    if [[ "$1" == "$2" ]]; then
        echo "  PASS: $3"
    else
        echo "  FAIL: $3 (expected '$2', got '$1')"
        FAIL=1
    fi
}

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

export DOTFILES_DIR="$TMP/dot"
export PRIVATE_DIR="$TMP/private"
DEST_DIR="$TMP/dest"
mkdir -p "$DOTFILES_DIR/base" "$PRIVATE_DIR" "$DEST_DIR"

# managed_map: one allowlisted plugin file + one non-allowlisted file
cat > "$DOTFILES_DIR/managed_map.txt" <<EOF
base/installed.base.json+claude-plugins-installed.json:$DEST_DIR/installed_plugins.json
base/settings.base.json+claude-settings.json:$DEST_DIR/settings.json
EOF

# base scaffolds (empty, as restructured)
echo '{"version":2,"plugins":{}}' > "$DOTFILES_DIR/base/installed.base.json"
echo '{}'                          > "$DOTFILES_DIR/base/settings.base.json"

# overlay sources at version X
echo '{"version":2,"plugins":{"foo@mp":[{"version":"1.0.0","lastUpdated":"2026-01-01T00:00:00Z"}]}}' \
    > "$PRIVATE_DIR/claude-plugins-installed.json"
echo '{"model":"opus"}' > "$PRIVATE_DIR/claude-settings.json"

# on-disk (dest): plugin file disk-ahead (version Y); settings also drifted (non-allowlisted)
echo '{"version":2,"plugins":{"foo@mp":[{"version":"2.0.0","lastUpdated":"2026-06-01T00:00:00Z"}]}}' \
    > "$DEST_DIR/installed_plugins.json"
echo '{"model":"xhigh"}' > "$DEST_DIR/settings.json"

reconcile_plugin_drift "$DOTFILES_DIR/managed_map.txt"

got_ver=$(jq -r '.plugins["foo@mp"][0].version' "$PRIVATE_DIR/claude-plugins-installed.json")
assert_eq "$got_ver" "2.0.0" "allowlisted plugin overlay backfilled to disk version"

got_model=$(jq -r '.model' "$PRIVATE_DIR/claude-settings.json")
assert_eq "$got_model" "opus" "non-allowlisted overlay left untouched"

if diff <(jq -S . "$PRIVATE_DIR/claude-plugins-installed.json") \
        <(jq -S . "$DEST_DIR/installed_plugins.json") >/dev/null; then
    echo "  PASS: plugin overlay matches disk after reconcile"
else
    echo "  FAIL: plugin overlay still differs from disk"
    FAIL=1
fi

echo ""
if [[ $FAIL -eq 0 ]]; then
    echo "All tests passed"
else
    echo "TESTS FAILED"
    exit 1
fi
