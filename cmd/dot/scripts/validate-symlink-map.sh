#!/bin/bash
# validate-symlink-map.sh - Verify all source paths in symlink_map.txt exist
#
# Usage:
#   ./scripts/validate-symlink-map.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SYMLINK_MAP="$DOTFILES_DIR/symlink_map.txt"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

if [[ ! -f "$SYMLINK_MAP" ]]; then
    echo -e "${RED}Error:${NC} symlink_map.txt not found at $DOTFILES_DIR"
    exit 1
fi

echo "Checking symlink_map.txt sources..."

while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

    source=$(echo "$line" | cut -d':' -f1)
    source_path="$DOTFILES_DIR/$source"

    if [[ ! -e "$source_path" ]]; then
        echo -e "  ${RED}MISSING:${NC} $source"
        ((ERRORS++))
    fi
done < "$SYMLINK_MAP"

if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}All sources exist${NC}"
else
    echo ""
    echo -e "${RED}$ERRORS missing source(s)${NC}"
fi

exit "$ERRORS"
