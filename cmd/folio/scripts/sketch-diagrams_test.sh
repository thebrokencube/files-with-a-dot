#!/usr/bin/env bash
# Smoke test for the bundled sketch-diagrams renderer against the synthetic fixture.
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
bundle="$here/../../../plugins/folio/skills/folio/scripts/sketch-diagrams.mjs"
tmp="$(mktemp -d)"
trap '/bin/rm -rf "$tmp"' EXIT
/bin/cp -Rf "$here/testdata/sketch-diagrams/." "$tmp/"

node "$bundle" "$tmp/index.html" 2>"$tmp/err1"
if grep -q '^error:' "$tmp/err1"; then cat "$tmp/err1"; echo "FAIL: errors on first render"; exit 1; fi
for id in pipeline model lanes; do
  grep -q "<svg data-elk=\"$id\" class=\"elk" "$tmp/index.html" || { echo "FAIL: $id not rendered"; exit 1; }
done

/bin/cp -f "$tmp/index.html" "$tmp/first.html"
node "$bundle" "$tmp/index.html" 2>"$tmp/err2"
cmp -s "$tmp/index.html" "$tmp/first.html" || { echo "FAIL: second render differs"; exit 1; }

node "$bundle" "$tmp/index.html" --check >/dev/null

printf '{"diagrams":{"pipeline":{"type":"flow","nodes":[{"id":"a","title":"A"}],"edges":[{"from":"a","to":"zz"}]}}}' >"$tmp/bad.json"
if node "$bundle" "$tmp/index.html" --spec "$tmp/bad.json" --check 2>/dev/null; then echo "FAIL: --check accepted a bad spec"; exit 1; fi
cmp -s "$tmp/index.html" "$tmp/first.html" || { echo "FAIL: --check wrote the page"; exit 1; }

echo "PASS: sketch-diagrams renders, is idempotent, and --check rejects bad specs"
