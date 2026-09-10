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
printf 'skills/folio:$HOME/.claude/skills/folio\n' > "$PRIVATE_ROOT/symlink_map.txt"
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
copy_sync_repo() {
  local destination="$1"
  mkdir -p "$destination"
  (
    cd "$ROOT"
    tar --exclude=.git --exclude=.jj --exclude=.backup -cf - .
  ) | (
    cd "$destination"
    tar -xf -
  )
}

LINK_HOME="$TMP/links-home"
LINK_PRIVATE="$LINK_HOME/.dotfiles.private"
mkdir -p "$LINK_PRIVATE/skills/link-only-skill" "$LINK_PRIVATE/work/skills/legacy-link-only" "$LINK_HOME/.claude/rules"
touch "$LINK_PRIVATE/skills/link-only-skill/SKILL.md" "$LINK_PRIVATE/work/skills/legacy-link-only/SKILL.md" "$LINK_PRIVATE/unrelated"
printf 'unrelated:$HOME/.unrelated-link\n' > "$LINK_PRIVATE/symlink_map.txt"
ln -s "$ROOT/configs/base/claude/.claude/rules/code-comments.md" "$LINK_HOME/.claude/rules/code-comments.md"
ln -s "$LINK_HOME/foreign-target" "$LINK_HOME/.claude/rules/workflow.md"
link_out="$(HOME="$LINK_HOME" DOTFILES_NONINTERACTIVE=1 PATH="/usr/bin:/bin" "$ROOT/cmd/dot/sync.sh" --links-only --force 2>&1)"
for skill_root in "$LINK_HOME/.claude/skills" "$LINK_HOME/.omp/agent/skills" "$LINK_HOME/.codex/skills"; do
  [[ "$(realpath "$skill_root/link-only-skill")" == "$(realpath "$LINK_PRIVATE/skills/link-only-skill")" ]] || fail "private skill link missing at $skill_root"
done
for row in \
  'skills/link-only-skill:$HOME/.claude/skills/link-only-skill' \
  'skills/link-only-skill:$HOME/.omp/agent/skills/link-only-skill' \
  'skills/link-only-skill:$HOME/.codex/skills/link-only-skill'; do
  grep -Fqx "$row" "$LINK_PRIVATE/symlink_map.txt" || fail "links-only private map migration missing $row"
done
[[ -d "$LINK_PRIVATE/work/skills/legacy-link-only" ]] || fail "links-only migrated legacy private state"
[[ ! -e "$LINK_HOME/.unrelated-link" ]] || fail "links-only applied unrelated private map entry"
[[ ! -L "$LINK_HOME/.claude/rules/code-comments.md" ]] || fail "retired raw dotfiles link was not removed"
[[ -L "$LINK_HOME/.claude/rules/workflow.md" ]] || fail "foreign dangling link was removed"
! grep -Fq 'Homebrew' <<<"$link_out" || fail "links-only performed Homebrew work"
! grep -Fq 'Migrate private overlay legacy state' <<<"$link_out" || fail "links-only planned legacy private migration"
ok "links-only applies public and declared private skill links without full-sync work"

agent_roots_out="$(HOME="$LINK_HOME" DOTFILES_DIR="$ROOT" "$ROOT/cmd/dot/dot" health --check agent-roots)"
grep -Fq 'Declared candidate agent roots:' <<<"$agent_roots_out" || fail "agent-roots health check is not visible"
for root in '.claude' '.omp/agent' '.codex'; do
  grep -Fq "$root: 7 declared map link(s)" <<<"$agent_roots_out" || fail "agent-roots health check missed $root"
done
ok "agent-roots health check reports declared map topology"

MAPLESS_HOME="$TMP/mapless-home"
MAPLESS_PRIVATE="$MAPLESS_HOME/.dotfiles.private"
mkdir -p "$MAPLESS_PRIVATE/skills/private-source-token"
touch "$MAPLESS_PRIVATE/skills/private-source-token/SKILL.md"
mapless_out="$(HOME="$MAPLESS_HOME" DOTFILES_NONINTERACTIVE=1 "$ROOT/cmd/dot/sync.sh" --links-only --dry-run 2>&1)"
grep -Fq 'Add private skill map rows to declared agent roots' <<<"$mapless_out" || fail "private map migration was not planned"
! grep -Fq 'private-source-token' <<<"$mapless_out" || fail "private skill source leaked during planning"
[[ ! -e "$MAPLESS_PRIVATE/symlink_map.txt" ]] || fail "dry-run wrote the private map"
ok "private planning uses opaque actions and performs no dry-run writes"

SYNC_REPO="$TMP/sync-repo"
copy_sync_repo "$SYNC_REPO"
printf 'conservative\n' > "$SYNC_REPO/.machine"
FAKE_BIN="$TMP/fake-bin"
mkdir -p "$FAKE_BIN"
cat > "$FAKE_BIN/brew" <<'SCRIPT'
#!/bin/sh
exit 0
SCRIPT
cat > "$FAKE_BIN/claude" <<'SCRIPT'
#!/bin/sh
exit 0
SCRIPT
cat > "$FAKE_BIN/curl" <<'SCRIPT'
#!/bin/sh
out=
while [ "$#" -gt 0 ]; do
  if [ "$1" = -o ]; then
    out="$2"
    shift 2
  else
    shift
  fi
done
[ -z "$out" ] || : > "$out"
exit 0
SCRIPT
chmod +x "$FAKE_BIN/brew" "$FAKE_BIN/claude" "$FAKE_BIN/curl"

REJECT_RUNNER="$TMP/reject-runner"
cat > "$REJECT_RUNNER" <<'SCRIPT'
#!/bin/bash
args=(--skip-brew)
[[ -z "${REJECT_MODE:-}" ]] || args=("$REJECT_MODE" "${args[@]}")
exec "$SYNC_REPO/cmd/dot/sync.sh" "${args[@]}"
SCRIPT
chmod +x "$REJECT_RUNNER"
run_interactive_sync() {
  local prompt="$1" answer="$2" spawn
  if [[ "$(uname -s)" == Darwin ]]; then
    spawn="script -q /dev/null $REJECT_RUNNER"
  else
    spawn="script -q -c $REJECT_RUNNER /dev/null"
  fi
  REJECT_SPAWN="$spawn" EXPECT_PROMPT="$prompt" EXPECT_ANSWER="$answer" expect <<'EXPECT'
set timeout 20
eval spawn $env(REJECT_SPAWN)
expect $env(EXPECT_PROMPT)
send -- "$env(EXPECT_ANSWER)\r"
expect eof
catch wait result
exit [lindex $result 3]
EXPECT
}

run_rejected_sync() {
  run_interactive_sync "$1" n
}

FULL_REJECT_HOME="$TMP/full-reject-home"
FULL_REJECT_PRIVATE="$FULL_REJECT_HOME/.dotfiles.private"
mkdir -p "$FULL_REJECT_PRIVATE/skills/full-reject-private-source" "$FULL_REJECT_PRIVATE/work/skills/legacy-full-reject"
touch "$FULL_REJECT_PRIVATE/skills/full-reject-private-source/SKILL.md" "$FULL_REJECT_PRIVATE/work/skills/legacy-full-reject/SKILL.md"
ln -s "$SYNC_REPO" "$FULL_REJECT_HOME/.dotfiles"
mkdir -p "$FULL_REJECT_HOME/.claude/rules"
ln -s "$SYNC_REPO/configs/base/claude/.claude/rules/code-comments.md" "$FULL_REJECT_HOME/.claude/rules/code-comments.md"
set +e
full_reject_out="$(HOME="$FULL_REJECT_HOME" SYNC_REPO="$SYNC_REPO" REJECT_MODE= PATH="$FAKE_BIN:/usr/bin:/bin" run_rejected_sync 'Proceed with sync?')"
full_reject_status=$?
set -e
[[ "$full_reject_status" -eq 2 ]] || fail "rejected full sync exited $full_reject_status"
[[ ! -e "$FULL_REJECT_PRIVATE/symlink_map.txt" ]] || fail "rejected full sync wrote the private map"
[[ -d "$FULL_REJECT_PRIVATE/work" ]] || fail "rejected full sync migrated legacy private state"
[[ ! -e "$FULL_REJECT_HOME/.claude/rules/personal-core.md" ]] || fail "rejected full sync wrote public links"
[[ -L "$FULL_REJECT_HOME/.claude/rules/code-comments.md" ]] || fail "rejected full sync removed a retired link"
! grep -Fq 'full-reject-private-source' <<<"$full_reject_out" || fail "full sync planning leaked a private source"

LINK_REJECT_HOME="$TMP/link-reject-home"
LINK_REJECT_PRIVATE="$LINK_REJECT_HOME/.dotfiles.private"
mkdir -p "$LINK_REJECT_PRIVATE/skills/link-reject-private-source" "$LINK_REJECT_PRIVATE/work/skills/legacy-link-reject"
touch "$LINK_REJECT_PRIVATE/skills/link-reject-private-source/SKILL.md" "$LINK_REJECT_PRIVATE/work/skills/legacy-link-reject/SKILL.md"
ln -s "$SYNC_REPO" "$LINK_REJECT_HOME/.dotfiles"
mkdir -p "$LINK_REJECT_HOME/.claude/rules"
ln -s "$SYNC_REPO/configs/base/claude/.claude/rules/code-comments.md" "$LINK_REJECT_HOME/.claude/rules/code-comments.md"
set +e
link_reject_out="$(HOME="$LINK_REJECT_HOME" SYNC_REPO="$SYNC_REPO" REJECT_MODE=--links-only PATH="$FAKE_BIN:/usr/bin:/bin" run_rejected_sync 'Proceed with sync?')"
link_reject_status=$?
set -e
[[ "$link_reject_status" -eq 2 ]] || fail "rejected links-only sync exited $link_reject_status"
[[ ! -e "$LINK_REJECT_PRIVATE/symlink_map.txt" ]] || fail "rejected links-only sync wrote the private map"
[[ -d "$LINK_REJECT_PRIVATE/work" ]] || fail "rejected links-only sync migrated legacy private state"
[[ ! -e "$LINK_REJECT_HOME/.claude/rules/personal-core.md" ]] || fail "rejected links-only sync wrote public links"
[[ -L "$LINK_REJECT_HOME/.claude/rules/code-comments.md" ]] || fail "rejected links-only sync removed a retired link"
! grep -Fq 'link-reject-private-source' <<<"$link_reject_out" || fail "links-only planning leaked a private source"

DIRECT_REJECT_HOME="$TMP/direct-reject-home"
DIRECT_REJECT_PRIVATE="$DIRECT_REJECT_HOME/.dotfiles.private"
mkdir -p "$DIRECT_REJECT_PRIVATE/skills/direct-reject-private-source"
touch "$DIRECT_REJECT_PRIVATE/skills/direct-reject-private-source/SKILL.md"
DIRECT_REJECT_RUNNER="$TMP/direct-reject-runner"
cat > "$DIRECT_REJECT_RUNNER" <<'SCRIPT'
#!/bin/bash
exec env DOTFILES_DIR="$SYNC_REPO" "$SYNC_REPO/cmd/dot/dot" private sync
SCRIPT
chmod +x "$DIRECT_REJECT_RUNNER"
set +e
direct_reject_out="$(HOME="$DIRECT_REJECT_HOME" SYNC_REPO="$SYNC_REPO" REJECT_RUNNER="$DIRECT_REJECT_RUNNER" PATH="$FAKE_BIN:/usr/bin:/bin" run_rejected_sync 'Apply private symlinks?')"
direct_reject_status=$?
set -e
[[ "$direct_reject_status" -eq 0 ]] || fail "rejected direct private sync exited $direct_reject_status"
[[ ! -e "$DIRECT_REJECT_PRIVATE/symlink_map.txt" ]] || fail "rejected direct private sync wrote the private map"
[[ ! -e "$DIRECT_REJECT_HOME/.claude/skills/direct-reject-private-source" ]] || fail "rejected direct private sync wrote private links"
! grep -Fq 'direct-reject-private-source' <<<"$direct_reject_out" || fail "direct private planning leaked a private source"
[[ ! -e "$SYNC_REPO/.backup" ]] || fail "rejected sync initialized backups"
ok "terminal rejection preserves full, links-only, and direct private state"

DIRECT_HOME="$TMP/direct-home"
DIRECT_PRIVATE="$DIRECT_HOME/.dotfiles.private"
mkdir -p "$DIRECT_PRIVATE/skills/direct-skill"
touch "$DIRECT_PRIVATE/skills/direct-skill/SKILL.md"
HOME="$DIRECT_HOME" DOTFILES_DIR="$SYNC_REPO" DOTFILES_NONINTERACTIVE=1 "$SYNC_REPO/cmd/dot/dot" private sync >/dev/null
[[ ! -e "$DIRECT_PRIVATE/symlink_map.txt" ]] || fail "default direct private sync wrote the map"
HOME="$DIRECT_HOME" DOTFILES_DIR="$SYNC_REPO" DOTFILES_NONINTERACTIVE=1 "$SYNC_REPO/cmd/dot/dot" private sync --force >/dev/null
direct_map_before="$(cat "$DIRECT_PRIVATE/symlink_map.txt")"
HOME="$DIRECT_HOME" DOTFILES_DIR="$SYNC_REPO" DOTFILES_NONINTERACTIVE=1 "$SYNC_REPO/cmd/dot/dot" private sync --force >/dev/null
[[ "$direct_map_before" == "$(cat "$DIRECT_PRIVATE/symlink_map.txt")" ]] || fail "repeated direct private sync changed the map"
for skill_root in "$DIRECT_HOME/.claude/skills" "$DIRECT_HOME/.omp/agent/skills" "$DIRECT_HOME/.codex/skills"; do
  [[ -L "$skill_root/direct-skill" ]] || fail "direct private sync missed $skill_root"
done
if HOME="$DIRECT_HOME" DOTFILES_DIR="$SYNC_REPO" DOTFILES_NONINTERACTIVE=1 "$SYNC_REPO/cmd/dot/dot" private sync --confirmed >/dev/null 2>&1; then
  fail "direct private sync accepted internal confirmation"
fi
ok "direct private sync gates, applies, and remains idempotent"

FULL_HOME="$TMP/full-home"
FULL_PRIVATE="$FULL_HOME/.dotfiles.private"
mkdir -p "$FULL_PRIVATE/work/skills/full-sync-skill"
touch "$FULL_PRIVATE/work/skills/full-sync-skill/SKILL.md"
interactive_full_out="$(HOME="$FULL_HOME" SYNC_REPO="$SYNC_REPO" REJECT_MODE= PATH="$FAKE_BIN:/usr/bin:/bin" run_interactive_sync 'Proceed with sync?' y)"
[[ "$(grep -Fo 'Proceed with sync?' <<<"$interactive_full_out" | wc -l | tr -d ' ')" -eq 1 ]] || fail "accepted full sync did not use one top-level gate"
! grep -Fq 'Apply private symlinks?' <<<"$interactive_full_out" || fail "accepted full sync prompted for private application"
[[ -d "$FULL_PRIVATE/skills/full-sync-skill" ]] || fail "accepted full sync did not move legacy private skill"
[[ ! -d "$FULL_PRIVATE/work" ]] || fail "accepted full sync retained legacy private state"
[[ -n "$(compgen -G "$FULL_PRIVATE/.legacy-work-*")" ]] || fail "accepted full sync did not archive legacy state"
for row in \
  'skills/full-sync-skill:$HOME/.claude/skills/full-sync-skill' \
  'skills/full-sync-skill:$HOME/.omp/agent/skills/full-sync-skill' \
  'skills/full-sync-skill:$HOME/.codex/skills/full-sync-skill'; do
  grep -Fqx "$row" "$FULL_PRIVATE/symlink_map.txt" || fail "accepted full sync map missing $row"
done
for skill_root in "$FULL_HOME/.claude/skills" "$FULL_HOME/.omp/agent/skills" "$FULL_HOME/.codex/skills"; do
  [[ -L "$skill_root/full-sync-skill" ]] || fail "accepted full sync missed $skill_root"
done
ok "accepted full sync migrates then links private skills in one run"

FILTERED_COLLISION_HOME="$TMP/filtered-collision-home"
FILTERED_PRIVATE="$FILTERED_COLLISION_HOME/.dotfiles.private"
mkdir -p "$FILTERED_PRIVATE"
touch "$FILTERED_PRIVATE/unrelated"
printf 'unrelated:$HOME/.claude/skills/folio\n' > "$FILTERED_PRIVATE/symlink_map.txt"
HOME="$FILTERED_COLLISION_HOME" bash -c '
  set -e
  DOTFILES_DIR="$1"
  source "$1/cmd/dot/lib/paths.sh"
  source "$1/cmd/dot/lib/symlinks.sh"
  source "$1/cmd/dot/lib/private.sh"
  check_private_destination_collisions "$2" "$3" true
' _ "$ROOT" "$PUBLIC_MAP" "$FILTERED_PRIVATE" || fail "links-only collision check considered an unrelated private map entry"
ok "links-only collision checks only declared private skill links"

echo "PASS: Claude distribution smoke green"
