#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
SCRIPT="$ROOT/cmd/folio/scripts/migrate-container.sh"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

UMBRELLA="$TMP/umbrella"
WRONG="$TMP/wrong"
BIN="$TMP/bin"
LOG="$TMP/folio.log"
mkdir -p "$UMBRELLA/.git" "$UMBRELLA/.jj" "$WRONG" "$BIN"

cat > "$BIN/folio" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
printf 'FOLIO_HOME=%s FOLIO_UMBRELLA=%s\n' "${FOLIO_HOME-unset}" "${FOLIO_UMBRELLA-unset}" >> "$FOLIO_LOG"
if [[ "${1-}" == "--version" ]]; then
  printf 'folio 0.0.12\n'
fi
SH

cat > "$BIN/jj" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
case "${1-}" in
  git)
    if [[ "${2-}" == "remote" && "${3-}" == "list" ]]; then
      printf 'origin https://example.invalid/folio.git\n'
    fi
    ;;
  log)
    if [[ " $* " == *" main@origin "* ]]; then
      printf 'same-main\n'
    elif [[ " $* " == *' mutable() '* ]]; then
      :
    else
      printf 'same-main\n'
    fi
    ;;
  workspace)
    if [[ "${2-}" == "list" ]]; then
      printf 'default: 00000000\n'
    fi
    ;;
esac
SH
chmod +x "$BIN/folio" "$BIN/jj"

output="$({
  FOLIO_HOME="$WRONG" FOLIO_UMBRELLA="$UMBRELLA" FOLIO_LOG="$LOG" \
    MIGRATE_STAMP=test PATH="$BIN:$PATH" bash "$SCRIPT" --check
} 2>&1)"

[[ "$output" == *"umbrella=$UMBRELLA"* ]] || {
  printf 'migration did not select FOLIO_UMBRELLA:\n%s\n' "$output" >&2
  exit 1
}
[[ "$output" != *"umbrella=$WRONG"* ]] || {
  printf 'migration selected conflicting FOLIO_HOME\n' >&2
  exit 1
}
[[ "$(cat "$LOG")" == *"FOLIO_HOME=unset FOLIO_UMBRELLA=$UMBRELLA"* ]] || {
  printf 'child folio call did not receive explicit umbrella state:\n%s\n' "$(cat "$LOG")" >&2
  exit 1
}

shell="$ROOT/configs/base/shell/.shell_common"
settings="$ROOT/configs/base/claude/settings.base.json"
expected_umbrella_export="export FOLIO_UMBRELLA=\"\$HOME/.folio\""
expected_home_export="export FOLIO_HOME=\"\$HOME/.folio\""
[[ "$(cat "$shell")" == *"$expected_umbrella_export"* ]] || exit 1
[[ "$(cat "$shell")" != *"$expected_home_export"* ]] || exit 1
for prefix in \
  'Bash(FOLIO_UMBRELLA=* FOLIO_HOME=/tmp/folio-ws-* folio *)' \
  'Bash(FOLIO_HOME=/tmp/folio-ws-* FOLIO_UMBRELLA=* folio *)' \
  'Bash(FOLIO_UMBRELLA=* FOLIO_HOME=/tmp/folio-ws-* ./folio *)' \
  'Bash(FOLIO_HOME=/tmp/folio-ws-* FOLIO_UMBRELLA=* ./folio *)'; do
  grep -Fq -- "\"$prefix\"" "$settings" || {
    printf 'missing Claude permission prefix: %s\n' "$prefix" >&2
    exit 1
  }
done

printf 'migrate-container contract: pass\n'
