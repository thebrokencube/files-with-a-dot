#!/usr/bin/env bash
# smoke-test.sh — deterministic acceptance test for the proven Claude distribution.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CATALOG=.claude-plugin/marketplace.json
fail() { echo "FAIL: $*" >&2; exit 1; }
ok() { echo "  ok: $*"; }

command -v jq >/dev/null || fail "jq is required"
command -v shellcheck >/dev/null || fail "shellcheck is required"
PLUGINS=()
while IFS= read -r plugin; do
  PLUGINS+=("$plugin")
done < <(jq -r '.plugins[].name' plugins.json)

before="$(md5 -q "$CATALOG" 2>/dev/null || md5sum "$CATALOG" | cut -d' ' -f1)"
scripts/marketplace-generate >/dev/null
after="$(md5 -q "$CATALOG" 2>/dev/null || md5sum "$CATALOG" | cut -d' ' -f1)"
[[ "$before" == "$after" ]] || fail "generated Claude catalog drifted"
[[ ! -e .cursor-plugin && ! -e .agents/plugins ]] || fail "unsupported catalog artifact exists"
ok "only deterministic Claude catalog exists"

jq -e . plugins.json >/dev/null
jq -e --argjson count "${#PLUGINS[@]}" '.description and (.plugins | length == $count)' "$CATALOG" >/dev/null
for tool in "${PLUGINS[@]}"; do
  bundle="plugins/$tool"
  manifest="$bundle/.claude-plugin/plugin.json"
  [[ -f "$manifest" && -f "$bundle/skills/$tool/SKILL.md" && -x "$bundle/bin/setup" ]] || fail "$tool bundle incomplete"
  cmp -s "cmd/$tool/VERSION" "$bundle/VERSION" || fail "$tool VERSION mirror drift"
  plugin_ver="$(jq -r '.version' "$manifest")"
  catalog_ver="$(jq -r --arg n "$tool" '.plugins[] | select(.name == $n) | .version' "$CATALOG")"
  [[ "$plugin_ver" == "$catalog_ver" ]] || fail "$tool native/catalog plugin version mismatch"
  while IFS= read -r path; do
    [[ ! -L "$bundle/$path" ]] || fail "$tool bundle contains symlink: $path"
    case "$path" in
      .claude-plugin/plugin.json|bin/setup|VERSION|skills/"$tool"/*) ;;
      *) fail "$tool bundle contains undeclared file: $path" ;;
    esac
  done < <(cd "$bundle" && find . \( -type f -o -type l \) | sed 's#^./##' | sort)
  shellcheck "$bundle/bin/setup"
  grep -q 'PLUGIN_RELEASE_BASE_URL' "$bundle/bin/setup" || fail "$tool setup lacks fixture seam"
done
cmp -s plugins/folio/bin/setup plugins/jf/bin/setup
cmp -s plugins/folio/bin/setup plugins/dendrik/bin/setup
ok "bundle closure, versions, and setup contracts pass"

if command -v claude >/dev/null; then
  claude plugin validate --strict "$CATALOG" >/dev/null
  for tool in "${PLUGINS[@]}"; do claude plugin validate --strict "plugins/$tool" >/dev/null; done
  ok "Claude native validation passes"
else
  echo "  SKIP: claude not installed"
fi

TMP="$(mktemp -d)"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$(uname -m)" in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) fail "unsupported test architecture" ;; esac
ver="$(tr -d '[:space:]' < cmd/folio/VERSION)"
asset="$TMP/http/folio/v$ver/folio-$os-$arch"
trap 'kill "${SERVER_PID:-}" 2>/dev/null || true; rm -rf "$TMP"' EXIT
mkdir -p "$(dirname "$asset")"
cat > "$asset" <<SCRIPT
#!/bin/sh
printf 'folio $ver\n'
SCRIPT
chmod +x "$asset"
port=18765
python3 -m http.server "$port" --directory "$TMP/http" >"$TMP/server.log" 2>&1 &
SERVER_PID=$!
sleep 1
out="$(HOME="$TMP/home" PLUGIN_RELEASE_BASE_URL="http://127.0.0.1:$port" plugins/folio/bin/setup)"
[[ -x "$TMP/home/.local/bin/folio" ]] || fail "fixture install missing"
[[ "$("$TMP/home/.local/bin/folio" --version)" == "folio $ver" ]] || fail "fixture version mismatch"
out="$(HOME="$TMP/home" PLUGIN_RELEASE_BASE_URL="http://127.0.0.1:$port" plugins/folio/bin/setup)"
grep -q 'already installed' <<<"$out" || fail "setup rerun not idempotent"
ok "network-free setup install and idempotency pass"


COLLISION_HOME="$TMP/collision-home"
PUBLIC_MAP="$TMP/public-map.txt"
PRIVATE_ROOT="$COLLISION_HOME/.dotfiles.private"
mkdir -p "$PRIVATE_ROOT/skills/folio"
printf 'plugins/folio/skills/folio:$HOME/.claude/skills/folio\n' > "$PUBLIC_MAP"
collision_out="$(HOME="$COLLISION_HOME" bash -c '
  set -e
  DOTFILES_DIR="$1"
  source "$1/cmd/dot/lib/paths.sh"
  source "$1/cmd/dot/lib/private.sh"
  check_private_destination_collisions "$2" "$3"
' _ "$ROOT" "$PUBLIC_MAP" "$PRIVATE_ROOT" 2>&1 || true)"
grep -q 'private overlay conflicts with public destination:' <<<"$collision_out" || fail "private collision not detected"
grep -q 'skills/folio' <<<"$collision_out" || fail "public destination missing from collision error"
! grep -q "$PRIVATE_ROOT" <<<"$collision_out" || fail "private source leaked in collision error"
ok "private collision fails without private source disclosure"

OWNERSHIP_HOME="$TMP/ownership-home"
mkdir -p "$OWNERSHIP_HOME/.dotfiles" "$OWNERSHIP_HOME/.dotfiles-evil"
ln -s "$OWNERSHIP_HOME/.dotfiles-evil/skill" "$OWNERSHIP_HOME/foreign-link"
if HOME="$OWNERSHIP_HOME" DOTFILES_DIR="$OWNERSHIP_HOME/.dotfiles" bash -c '
  source "$1/cmd/dot/lib/paths.sh"
  is_ours "$2"
' _ "$ROOT" "$OWNERSHIP_HOME/foreign-link"; then
  fail "sibling dotfiles path classified as managed"
fi
ok "managed path ownership requires an exact root boundary"
echo "PASS: Claude distribution smoke green"
