#!/usr/bin/env bash
# migrate-container.sh — one-time migration of a single-home ~/.folio into the
# multi-store CONTAINER model (see cmd/folio/skill/references/container-migration.md).
#
# Demotes the live colocated git+jj home to a store nested under a plain umbrella
# directory. The store name derives from the origin repo name (--store overrides):
#
#   BEFORE: ~/.folio                    (colocated git+jj repo, single-home)
#   AFTER:  ~/.folio/                   (plain umbrella dir — NOT a repo)
#           └── <store>/                (fresh colocated clone of origin)
#           └── stores.yml              (registry; later dotfile-managed)
#
# STRATEGY: clone-beside, never move a .jj (jj workspaces store ABSOLUTE
# back-pointers to the repo root — moving the root corrupts every workspace).
# All work must be pushed first; a fresh clone from origin is then complete. A
# cp -a backup and a two-mv swap make every step reversible.
#
# SAFETY: defaults to --check (dry-run, no mutations). Pass --execute to mutate.
# Aborts on any dirty/unpushed state before the irreversible swap.
#
# PREREQUISITE (binary-first ordering): folio >= 0.0.4 MUST already be on PATH
# (the v2 binary that understands stores.yml). An older binary would treat the
# umbrella as the home and silently stop creating workspaces.

set -euo pipefail

UMBRELLA="${FOLIO_HOME:-$HOME/.folio}"
MODE="check"
STORE="" # store (= nested dir) name; defaults to the origin repo name (Phase 0)
STAMP="${MIGRATE_STAMP:-}" # caller may pin a timestamp; else derived below

# Store (= nested dir) name and bootstrap `default:` derive from the origin repo
# name (Phase 0). Override with --store <name> if the dir name must differ.
prev=""
for arg in "$@"; do
  if [[ "$prev" == "--store" ]]; then STORE="$arg"; prev=""; continue; fi
  case "$arg" in
    --execute) MODE="execute" ;;
    --check)   MODE="check" ;;
    --store)   prev="--store" ;;
    *) echo "unknown arg: $arg (use --check | --execute | --store <name>)" >&2; exit 2 ;;
  esac
done
[[ "$prev" == "--store" ]] && { echo "--store needs a value" >&2; exit 2; }

# Date.now is fine in a shell script (unlike the workflow engine); derive a
# timestamp for backup/old dir names unless the caller pinned one.
[[ -z "$STAMP" ]] && STAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP="$UMBRELLA.backup-$STAMP"
OLD="$UMBRELLA.old-$STAMP"
STAGING="$UMBRELLA.new-$STAMP"

say()  { printf '  %s\n' "$*"; }
step() { printf '\n[%s] %s\n' "$1" "$2"; }
die()  { printf 'ABORT: %s\n' "$*" >&2; exit 1; }
run()  { if [[ "$MODE" == execute ]]; then "$@"; else say "(dry-run) $*"; fi; }

printf 'folio container migration — mode=%s umbrella=%s\n' "$MODE" "$UMBRELLA"

# ── Phase 0: preconditions ────────────────────────────────────────────────────
step 0 "Preconditions"
command -v folio >/dev/null || die "folio not on PATH"
VER="$(folio --version | awk '{print $NF}')"
say "folio version: $VER"
case "$VER" in
  0.0.[0-3]) die "folio $VER is too old — release & 'dot sync' v0.0.4+ first (binary-first ordering)" ;;
esac
[[ -d "$UMBRELLA/.git" && -d "$UMBRELLA/.jj" ]] || die "$UMBRELLA is not a colocated git+jj repo (already migrated, or unexpected layout)"
[[ -f "$UMBRELLA/stores.yml" ]] && die "$UMBRELLA/stores.yml already exists — looks already migrated"
[[ -e "$STAGING" || -e "$OLD" ]] && die "stale $STAGING or $OLD exists — clean up a prior aborted run first"

cd "$UMBRELLA"
REMOTE="$(jj git remote list | awk '$1=="origin"{print $2}')"
[[ -n "$REMOTE" ]] || die "no 'origin' remote on $UMBRELLA"
say "origin: $REMOTE"
# Derive the store (= nested dir) name from the origin repo name unless overridden.
[[ -z "$STORE" ]] && STORE="$(basename "$REMOTE" .git)"
[[ -n "$STORE" ]] || die "could not derive store name from remote $REMOTE — pass --store <name>"
say "store: $STORE"

# ── Phase 1: push gate — a fresh clone only captures what origin has ──────────
step 1 "Push gate (abort on unpushed/dirty)"
# Every mutable, non-empty change must sit on a pushed bookmark. The cheap proxy:
# local main must equal main@origin, and no mutable non-empty change may be
# missing from origin. We surface the state and refuse if main is ahead.
jj git fetch >/dev/null 2>&1 || true
LOCAL_MAIN="$(jj log --no-graph -r main -T 'commit_id' 2>/dev/null || true)"
ORIGIN_MAIN="$(jj log --no-graph -r main@origin -T 'commit_id' 2>/dev/null || true)"
say "main        = ${LOCAL_MAIN:-<none>}"
say "main@origin = ${ORIGIN_MAIN:-<none>}"
if [[ "$LOCAL_MAIN" != "$ORIGIN_MAIN" ]]; then
  die "main != main@origin — run 'folio home push' (and push every workspace) before migrating"
fi
# Any mutable change that is not an ancestor of main@origin is unpushed work.
UNPUSHED="$(jj log --no-graph -r 'mutable() & ~empty() & ~::main@origin' -T 'change_id ++ "\n"' 2>/dev/null | grep -c . || true)"
if [[ "${UNPUSHED:-0}" -gt 0 ]]; then
  echo "Unpushed non-empty changes:" >&2
  jj log -r 'mutable() & ~empty() & ~::main@origin' --no-pager >&2 || true
  die "$UNPUSHED unpushed change(s) — push them (the clone would lose them) or stash to the backup, then retry"
fi
say "all work is on origin"

# ── Phase 2: backup (full, including unpushed/uncommitted, for safety) ────────
step 2 "Backup → $BACKUP"
run cp -a "$UMBRELLA" "$BACKUP"

# ── Phase 3: forget all jj workspaces (incl. this session's) ──────────────────
step 3 "Forget jj workspaces"
# Workspaces hold absolute back-pointers into the OLD repo root; the fresh clone
# is a different root, so they must be forgotten. The /tmp dirs are reaped by the
# session-start hook. This session must re-create its workspace post-migration.
WS="$(jj workspace list | awk -F: '{print $1}')"
for w in $WS; do
  [[ "$w" == "default" ]] && continue
  say "forget workspace: $w"
  run jj workspace forget "$w"
done

# ── Phase 4: clone-beside into staging umbrella ───────────────────────────────
step 4 "Clone-beside → $STAGING/$STORE"
run mkdir -p "$STAGING"
run git clone "$REMOTE" "$STAGING/$STORE"
# Colocate jj so 'folio home workspace' works against the new store.
if [[ "$MODE" == execute ]]; then
  ( cd "$STAGING/$STORE" && jj git init --colocate >/dev/null )
else
  say "(dry-run) cd $STAGING/$STORE && jj git init --colocate"
fi

# ── Phase 5: verification gate (abort BEFORE the swap) ────────────────────────
step 5 "Verify clone (gate before swap)"
if [[ "$MODE" == execute ]]; then
  [[ -d "$STAGING/$STORE/.git" && -d "$STAGING/$STORE/.jj" ]] || die "clone is not colocated git+jj — aborting before swap"
  WANT="$(ls -1 "$UMBRELLA/active" 2>/dev/null | wc -l | tr -d ' ')"
  GOT="$(ls -1 "$STAGING/$STORE/active" 2>/dev/null | wc -l | tr -d ' ')"
  say "active project dirs: backup=$WANT clone=$GOT"
  [[ "$WANT" == "$GOT" ]] || die "active project count differs (backup=$WANT clone=$GOT) — origin may be behind; aborting before swap"
else
  say "(dry-run) would assert: clone colocated + active project count matches"
fi

# ── Phase 6: two-mv swap (reversible) ─────────────────────────────────────────
step 6 "Swap umbrella into place"
run mv "$UMBRELLA" "$OLD"
run mv "$STAGING" "$UMBRELLA"

# ── Phase 7: write bootstrap stores.yml ───────────────────────────────────────
# A minimal registry so folio works immediately. This is later SUPERSEDED by the
# dotfile-managed stores.yml (author the base+private halves, then 'dot sync' —
# see the runbook). 'dot sync' drift-check will flag any divergence.
step 7 "Bootstrap stores.yml"
if [[ "$MODE" == execute ]]; then
  cat > "$UMBRELLA/stores.yml" <<EOF
schema: 2
default: $STORE
stores:
  $STORE: { path: ~/.folio/$STORE, kind: folio, remote: $REMOTE }
EOF
  say "wrote $UMBRELLA/stores.yml (default: $STORE)"
else
  say "(dry-run) would write $UMBRELLA/stores.yml (schema 2, default $STORE)"
fi

# ── Phase 8: smoke test ───────────────────────────────────────────────────────
step 8 "Smoke test"
if [[ "$MODE" == execute ]]; then
  if ( cd "$UMBRELLA" && folio home list >/dev/null ); then
    say "folio home list OK (resolves default store)"
  else
    die "folio home list failed post-migration — roll back (see below)"
  fi
else
  say "(dry-run) would run: folio home list (from umbrella)"
fi

cat <<EOF

DONE (mode=$MODE).
  Umbrella : $UMBRELLA
  Store    : $UMBRELLA/$STORE
  Backup   : $BACKUP
  Old repo : $OLD   (the pre-migration colocated repo)

NEXT (manual, see runbook):
  1. Author the dotfile-managed stores.yml halves and 'dot sync' so the registry
     is reproducible (configs/base/folio/stores.base.yml + ~/.dotfiles.private/folio-stores.yml).
  2. (optional) Clone folio-vault and the external KBs (adr, radr, ...) as siblings.
  3. Re-create THIS session's workspace: folio home workspace create
  4. After a confidence period: rm -rf "$OLD" "$BACKUP"

ROLLBACK (before you delete \$OLD):
  rm -rf "$UMBRELLA" && mv "$OLD" "$UMBRELLA"
EOF
