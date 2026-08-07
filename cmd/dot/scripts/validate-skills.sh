#!/bin/bash
# validate-skills.sh - Verify skill frontmatter has required fields
#
# Each skill SKILL.md must have YAML frontmatter with 'name' and 'description'.
# Skill paths are derived from symlink_map.txt (lines targeting ~/.claude/skills/*).
#
# Usage:
#   ./scripts/validate-skills.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SYMLINK_MAP="$DOTFILES_DIR/symlink_map.txt"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

if [[ ! -f "$SYMLINK_MAP" ]]; then
    echo -e "${RED}Error:${NC} symlink_map.txt not found at $SYMLINK_MAP"
    exit 1
fi

echo "Checking skill frontmatter..."

# Parse symlink_map.txt for skill entries (destination matches $HOME/.claude/skills/*)
while IFS=: read -r source dest; do
    # Skip comments and blank lines
    [[ -z "$source" || "$source" =~ ^[[:space:]]*# ]] && continue

    # Expand $HOME in destination and check if it's a skill
    expanded_dest="${dest/\$HOME/$HOME}"
    [[ "$expanded_dest" == "$HOME/.claude/skills/"* ]] || continue

    # Resolve source relative to DOTFILES_DIR
    skill_source="$DOTFILES_DIR/$source"
    skill_name=$(basename "$expanded_dest")
    skill_file="$skill_source/SKILL.md"

    if [[ ! -f "$skill_file" ]]; then
        echo -e "  ${RED}MISSING:${NC} $skill_name/SKILL.md (source: $source)"
        ERRORS=$((ERRORS + 1))
        continue
    fi

    # Check for YAML frontmatter (starts with ---)
    first_line=$(head -1 "$skill_file")
    if [[ "$first_line" != "---" ]]; then
        echo -e "  ${RED}NO FRONTMATTER:${NC} $skill_name/SKILL.md"
        ERRORS=$((ERRORS + 1))
        continue
    fi

    # Extract frontmatter (between first and second ---)
    frontmatter=$(sed -n '2,/^---$/p' "$skill_file" | sed '$d')

    has_name=false
    has_description=false

    if echo "$frontmatter" | grep -q '^name:'; then
        has_name=true
    fi
    if echo "$frontmatter" | grep -q '^description:'; then
        has_description=true
    fi

    if [[ "$has_name" == false || "$has_description" == false ]]; then
        missing=""
        [[ "$has_name" == false ]] && missing="name"
        [[ "$has_description" == false ]] && missing="${missing:+$missing, }description"
        echo -e "  ${RED}MISSING FIELDS:${NC} $skill_name ($missing)"
        ERRORS=$((ERRORS + 1))
    fi
done < "$SYMLINK_MAP"

if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}All skills valid${NC}"
else
    echo ""
    echo -e "${RED}$ERRORS skill issue(s)${NC}"
fi

exit "$ERRORS"
