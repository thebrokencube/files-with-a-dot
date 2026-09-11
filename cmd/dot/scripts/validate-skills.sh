#!/bin/bash
# validate-skills.sh - Verify skill frontmatter has required fields
#
# Each shared skill must project once to each declared candidate agent root.
# Skill sources are de-duplicated so every SKILL.md is validated once.
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

echo "Checking skill frontmatter and declared roots..."

AGENT_SKILL_ROOTS=(
    '$HOME/.claude/skills'
    '$HOME/.omp/agent/skills'
    '$HOME/.codex/skills'
)
SKILL_SOURCES=()

agent_skill_name() {
    local destination="$1" root name
    for root in "${AGENT_SKILL_ROOTS[@]}"; do
        if [[ "$destination" == "$root/"* ]]; then
            name="${destination#"$root/"}"
            [[ -n "$name" && "$name" != */* ]] || return 1
            printf '%s\n' "$name"
            return 0
        fi
    done
    return 1
}
source_seen() {
    local candidate="$1" source
    [[ ${#SKILL_SOURCES[@]} -gt 0 ]] || return 1
    for source in "${SKILL_SOURCES[@]}"; do
        [[ "$source" == "$candidate" ]] && return 0
    done
    return 1
}

while IFS=: read -r source dest; do
    [[ -z "$source" || "$source" =~ ^[[:space:]]*# ]] && continue

    skill_name="$(agent_skill_name "$dest" || true)"
    [[ -n "$skill_name" ]] || continue

    if [[ "$(basename "$source")" != "$skill_name" ]]; then
        echo -e "  ${RED}WRONG BASENAME:${NC} $source:$dest"
        ERRORS=$((ERRORS + 1))
    fi
    source_seen "$source" || SKILL_SOURCES+=("$source")
done < "$SYMLINK_MAP"

if [[ ${#SKILL_SOURCES[@]} -gt 0 ]]; then
for source in "${SKILL_SOURCES[@]}"; do
    skill_name="$(basename "$source")"
    skill_source="$DOTFILES_DIR/$source"
    skill_file="$skill_source/SKILL.md"

    if [[ ! -f "$skill_file" ]]; then
        echo -e "  ${RED}MISSING:${NC} $skill_name/SKILL.md (source: $source)"
        ERRORS=$((ERRORS + 1))
        continue
    fi

    first_line=$(head -1 "$skill_file")
    if [[ "$first_line" != "---" ]]; then
        echo -e "  ${RED}NO FRONTMATTER:${NC} $skill_name/SKILL.md"
        ERRORS=$((ERRORS + 1))
        continue
    fi

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

    for root in "${AGENT_SKILL_ROOTS[@]}"; do
        expected="$source:$root/$skill_name"
        count=$(grep -Fxc -- "$expected" "$SYMLINK_MAP" || true)
        if [[ "$count" -eq 0 ]]; then
            echo -e "  ${RED}MISSING PROJECTION:${NC} $expected"
            ERRORS=$((ERRORS + 1))
        elif [[ "$count" -gt 1 ]]; then
            echo -e "  ${RED}DUPLICATE PROJECTION:${NC} $expected"
            ERRORS=$((ERRORS + 1))
        fi
    done
done
fi

POLICY_FILE="$DOTFILES_DIR/agents/AGENTS.md"
if [[ ! -f "$POLICY_FILE" ]]; then
    echo -e "  ${RED}MISSING:${NC} agents/AGENTS.md"
    ERRORS=$((ERRORS + 1))
elif [[ $(wc -c < "$POLICY_FILE") -gt 2048 ]]; then
    echo -e "  ${RED}OVERSIZED:${NC} agents/AGENTS.md exceeds 2,048 bytes"
    ERRORS=$((ERRORS + 1))
fi

if [[ $ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}All skills valid${NC}"
else
    echo ""
    echo -e "${RED}$ERRORS skill issue(s)${NC}"
fi

exit "$ERRORS"
