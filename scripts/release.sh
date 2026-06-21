#!/usr/bin/env bash
# Release a dendrik tool. Thin shim over `dendrik build` (the build primitive) and `gh`
# (GitHub orchestration) — keeps logic out of release.yml so it is testable and reusable.
# Called by .github/workflows/release.yml; also runnable locally (needs gh auth).
#
# Usage: scripts/release.sh <tool>
#   Version is read from cmd/<tool>/VERSION (the single source of truth). Bump that file
#   first. Published releases are immutable: re-running an existing version fails.
set -euo pipefail

tool="${1:?usage: scripts/release.sh <tool>}"
repo_root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$repo_root"

version_file="cmd/$tool/VERSION"
[ -f "$version_file" ] || { echo "::error::no $version_file" >&2; exit 1; }
v="$(tr -d '[:space:]' < "$version_file")"

# --- validate: semver · immutable · monotonic ---
if ! printf '%s' "$v" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$'; then
  echo "::error::VERSION '$v' is not valid semver (X.Y.Z)" >&2; exit 1
fi
tag="$tool/v$v"
if gh release view "$tag" >/dev/null 2>&1; then
  echo "::error::release $tag already exists — bump $version_file" >&2; exit 1
fi
latest="$(gh release list --limit 200 --json tagName -q '.[].tagName' \
  | sed -n "s#^$tool/v##p" | sort -V | tail -1 || true)"
if [ -n "${latest:-}" ]; then
  top="$(printf '%s\n%s\n' "$latest" "$v" | sort -V | tail -1)"
  if [ "$v" = "$latest" ] || [ "$top" != "$v" ]; then
    echo "::error::version $v must be greater than the latest released $latest" >&2; exit 1
  fi
fi

# --- build via the dendrik primitive (bootstrap dendrik; it can't build itself via itself) ---
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
go build -C cmd/dendrik -o "$tmp/dendrik" .
"$tmp/dendrik" build "cmd/$tool" --matrix --out dist

# --- create the release (gh creates the tag server-side at this commit) ---
# No auto-generated notes: GitHub's --generate-notes diffs against the previous tag and
# is not tool-scoped, so it's noise under our per-tool tagging. Body is intentionally empty.
gh release create "$tag" dist/* --title "$tag" --notes "" --target "${GITHUB_SHA:-HEAD}"
echo "released $tag"
