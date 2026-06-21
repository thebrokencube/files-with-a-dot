#!/usr/bin/env bash
# smoke-test.sh — clean-install acceptance test for the plugin marketplace.
# Exit 0 = all checks pass. The primary guard for the marketplace (generation, manifests, install).
# Network is used only in the install step (step 5); it gates/skips offline.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PLUGINS=(folio jf dendrik)
CATALOGS=(.claude-plugin/marketplace.json .cursor-plugin/marketplace.json .agents/plugins/marketplace.json)
fail() { echo "FAIL: $*" >&2; exit 1; }
ok()   { echo "  ok: $*"; }

command -v jq >/dev/null || fail "jq is required"
command -v shellcheck >/dev/null || fail "shellcheck is required"

echo "[1] catalogs present + deterministic (drift guard) + _generated header"
sum_catalogs() { for c in "${CATALOGS[@]}"; do md5 -q "$c" 2>/dev/null || md5sum "$c" | awk '{print $1}'; done; }
for c in "${CATALOGS[@]}"; do [[ -f "$c" ]] || fail "$c missing — run scripts/marketplace-generate"; done
before="$(sum_catalogs)"
./scripts/marketplace-generate >/dev/null
after="$(sum_catalogs)"
[[ "$before" == "$after" ]] || fail "catalog drift — regenerated output differs from committed; regenerate and commit"
for c in "${CATALOGS[@]}"; do
  jq -e '._generated' "$c" >/dev/null || fail "$c missing _generated header"
done
ok "3 catalogs present, deterministic, headered"

echo "[2] jq-parse plugins.json + catalogs + each plugin.json"
jq -e . plugins.json >/dev/null || fail "plugins.json not valid JSON"
for c in "${CATALOGS[@]}"; do jq -e . "$c" >/dev/null || fail "$c not valid JSON"; done
for t in "${PLUGINS[@]}"; do jq -e . "cmd/$t/.claude-plugin/plugin.json" >/dev/null || fail "cmd/$t plugin.json not valid JSON"; done
ok "all manifests parse"

echo "[3] version triple per tool: VERSION == plugin.json.version == catalog listing"
for t in "${PLUGINS[@]}"; do
  vfile="$(tr -d '[:space:]' < "cmd/$t/VERSION")"
  vjson="$(jq -r '.version' "cmd/$t/.claude-plugin/plugin.json")"
  [[ "$vfile" == "$vjson" ]] || fail "$t: VERSION ($vfile) != plugin.json.version ($vjson)"
  # catalog lists the plugin by name + source; assert it is present
  jq -e --arg n "$t" '.plugins[] | select(.name == $n)' .claude-plugin/marketplace.json >/dev/null \
    || fail "$t: not listed in Claude catalog"
  ok "$t: VERSION == plugin.json.version == $vfile, listed in catalog"
done

echo "[4] bin/setup conventions: byte-identical x3, shellcheck-clean, self-locating, no CLAUDE_PLUGIN_ROOT, no sibling source"
ref="cmd/folio/bin/setup"
refsum="$(md5 -q "$ref" 2>/dev/null || md5sum "$ref" | awk '{print $1}')"
for t in "${PLUGINS[@]}"; do
  s="cmd/$t/bin/setup"
  [[ -x "$s" ]] || fail "$s not executable"
  sum="$(md5 -q "$s" 2>/dev/null || md5sum "$s" | awk '{print $1}')"
  [[ "$sum" == "$refsum" ]] || fail "$s not byte-identical to $ref"
  shellcheck "$s" || fail "$s shellcheck failed"
  # shellcheck disable=SC2016  # literal $0 is intended — asserting the script self-locates
  grep -q 'dirname "\$0"' "$s" || fail "$s not self-locating (\$0)"
  grep -q 'CLAUDE_PLUGIN_ROOT' "$s" && fail "$s references CLAUDE_PLUGIN_ROOT"
  grep -qE '^\s*(source|\.)\s' "$s" && fail "$s sources a sibling file"
done
ok "bin/setup x3 byte-identical, shellcheck-clean, self-locating, harness-neutral"

echo "[5] HOME-isolated install of cmd/folio/bin/setup + idempotent re-run"
if ! curl -fsI "https://github.com" >/dev/null 2>&1; then
  echo "  SKIP: offline — install step needs network"
else
  TMPHOME="$(mktemp -d)"
  trap 'rm -rf "$TMPHOME"' EXIT
  HOME="$TMPHOME" ./cmd/folio/bin/setup
  dest="$TMPHOME/.local/bin/folio"
  [[ -x "$dest" ]] || fail "folio not installed at $dest"
  got="$("$dest" --version | awk '{print $2}')"
  want="$(tr -d '[:space:]' < cmd/folio/VERSION)"
  [[ "$got" == "$want" ]] || fail "installed folio version ($got) != VERSION ($want)"
  # idempotent re-run: must report no-op and not re-download
  out="$(HOME="$TMPHOME" ./cmd/folio/bin/setup)"
  echo "$out" | grep -q "already installed" || fail "re-run was not a no-op: $out"
  ok "installed folio $got and re-run is a no-op"
fi

echo "PASS: all smoke-test checks green"
