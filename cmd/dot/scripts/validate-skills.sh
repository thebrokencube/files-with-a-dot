#!/bin/bash
# validate-skills.sh - Verify skill frontmatter has required fields
#
# Each skills/*/SKILL.md must have YAML frontmatter with 'name' and 'description'.
#
# Usage:
#   ./scripts/validate-skills.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
SKILLS_DIR="$DOTFILES_DIR/configs/base/claude/.claude/skills"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

ERRORS=0

if [[ ! -d "$SKILLS_DIR" ]]; then
    echo -e "${RED}Error:${NC} Skills directory not found at $DOTFILES_DIR/configs/base/claude/.claude/skills"
    exit 1
fi

echo "Checking skill frontmatter..."

for skill_dir in "$SKILLS_DIR"/*/; do
    [[ -d "$skill_dir" ]] || continue
    skill_name=$(basename "$skill_dir")
    skill_file="$skill_dir/SKILL.md"

    if [[ ! -f "$skill_file" ]]; then
        echo -e "  ${RED}MISSING:${NC} $skill_name/SKILL.md"
        ((ERRORS++))
        continue
    fi

    # Check for YAML frontmatter (starts with ---)
    first_line=$(head -1 "$skill_file")
    if [[ "$first_line" != "---" ]]; then
        echo -e "  ${RED}NO FRONTMATTER:${NC} $skill_name/SKILL.md"
        ((ERRORS++))
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
        ((ERRORS++))
    fi
done

if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}All skills valid${NC}"
else
    echo ""
    echo -e "${RED}$ERRORS skill issue(s)${NC}"
fi

exit "$ERRORS"
