#!/bin/bash
# validate.sh - Run all validation checks on the dotfiles repo
#
# Checks:
#   1. shellcheck on all .sh files + dot script
#   2. bash -n syntax check on all scripts
#   3. symlink_map.txt source verification
#   4. Skill frontmatter validation
#
# Usage:
#   ./scripts/validate.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_DIR="$(cd "$DOT_DIR/../.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

TOTAL_ERRORS=0
CHECKS_RUN=0
CHECKS_PASSED=0

run_check() {
    local name="$1"
    shift
    ((CHECKS_RUN++))

    echo -e "${CYAN}[$CHECKS_RUN]${NC} $name"
    if "$@"; then
        ((CHECKS_PASSED++))
        echo ""
    else
        local rc=$?
        ((TOTAL_ERRORS += rc))
        echo ""
    fi
}

echo -e "${BOLD}============================================${NC}"
echo -e "${BOLD}  Dotfiles Validation${NC}"
echo -e "${BOLD}============================================${NC}"
echo ""

# ============================================================================
# 1. shellcheck
# ============================================================================

shellcheck_scripts() {
    if ! command -v shellcheck &>/dev/null; then
        echo -e "  ${YELLOW}SKIP:${NC} shellcheck not installed (brew install shellcheck)"
        return 0
    fi

    local errors=0
    local files=()

    # Collect core command .sh files (cmd/dot/)
    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR" -maxdepth 1 -name '*.sh' -print0)

    # Collect scripts/ .sh files (cmd/dot/scripts/)
    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR/scripts" -name '*.sh' -print0 2>/dev/null)

    # Collect lib/ .sh files (cmd/dot/lib/)
    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR/lib" -name '*.sh' -print0 2>/dev/null)

    # Add the dot script
    files+=("$DOT_DIR/dot")

    for f in "${files[@]}"; do
        local name="${f#"$REPO_DIR/"}"
        # SC1090/SC1091: sourced files not followed (dynamic source paths)
        if shellcheck -S warning -e SC1091,SC1090 "$f" 2>&1; then
            :
        else
            echo -e "  ${RED}FAIL:${NC} $name"
            ((errors++))
        fi
    done

    if [[ $errors -eq 0 ]]; then
        echo -e "  ${GREEN}All files pass shellcheck${NC}"
    fi
    return "$errors"
}

run_check "shellcheck" shellcheck_scripts

# ============================================================================
# 2. bash -n syntax check
# ============================================================================

syntax_check() {
    local errors=0
    local files=()

    # Collect core command .sh files (cmd/dot/)
    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR" -maxdepth 1 -name '*.sh' -print0)

    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR/scripts" -name '*.sh' -print0 2>/dev/null)

    while IFS= read -r -d '' f; do
        files+=("$f")
    done < <(find "$DOT_DIR/lib" -name '*.sh' -print0 2>/dev/null)

    # Add dot script
    files+=("$DOT_DIR/dot")

    for f in "${files[@]}"; do
        local name="${f#"$REPO_DIR/"}"
        if ! bash -n "$f" 2>&1; then
            echo -e "  ${RED}FAIL:${NC} $name"
            ((errors++))
        fi
    done

    if [[ $errors -eq 0 ]]; then
        echo -e "  ${GREEN}All files pass syntax check${NC}"
    fi
    return "$errors"
}

run_check "bash -n syntax check" syntax_check

# ============================================================================
# 3. symlink_map.txt validation
# ============================================================================

run_check "symlink_map.txt sources" "$SCRIPT_DIR/validate-symlink-map.sh"

# ============================================================================
# 4. Skill frontmatter validation
# ============================================================================

run_check "skill frontmatter" "$SCRIPT_DIR/validate-skills.sh"

# ============================================================================
# Summary
# ============================================================================

echo "============================================"
if [[ $TOTAL_ERRORS -eq 0 ]]; then
    echo -e "  ${GREEN}All $CHECKS_RUN checks passed!${NC}"
else
    echo -e "  ${RED}$((CHECKS_RUN - CHECKS_PASSED))/$CHECKS_RUN checks failed${NC} ($TOTAL_ERRORS total issue(s))"
fi
echo "============================================"

exit "$((TOTAL_ERRORS > 0 ? 1 : 0))"
