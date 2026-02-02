#!/bin/bash
# brew-cleanup.sh - Find and optionally remove packages not in Brewfile
#
# Usage:
#   ./brew-cleanup.sh              # Show packages not in Brewfile
#   ./brew-cleanup.sh --remove     # Remove packages not in Brewfile (with confirmation)
#   ./brew-cleanup.sh --force      # Remove without confirmation

set -e

# ============================================================================
# Parse arguments
# ============================================================================

REMOVE=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --remove)
            REMOVE=true
            shift
            ;;
        --force)
            REMOVE=true
            FORCE=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./brew-cleanup.sh [OPTIONS]"
            echo ""
            echo "Find packages installed via Homebrew that are not in your Brewfiles."
            echo ""
            echo "Options:"
            echo "  --remove    Remove unlisted packages (with confirmation)"
            echo "  --force     Remove without confirmation"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# ============================================================================
# Setup
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$SCRIPT_DIR"
MACHINE_FILE="$DOTFILES_DIR/.machine"
PROFILE_FILE="$DOTFILES_DIR/.profile"
MACHINE_TYPE="aggressive"
MACHINE_PROFILE=""

[[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")
[[ -f "$PROFILE_FILE" ]] && MACHINE_PROFILE=$(cat "$PROFILE_FILE")

# Source private overlay support
source "$SCRIPT_DIR/lib/private.sh"

echo "============================================"
echo "  Homebrew Package Cleanup"
echo "============================================"
echo ""

# ============================================================================
# Build list of expected packages from Brewfiles
# ============================================================================

EXPECTED_FORMULAS=()
EXPECTED_CASKS=()

parse_brewfile() {
    local brewfile="$1"
    if [[ -f "$brewfile" ]]; then
        while IFS= read -r line; do
            # Skip comments and empty lines
            [[ "$line" =~ ^[[:space:]]*# ]] && continue
            [[ -z "$line" ]] && continue

            # Parse brew "package" or brew 'package'
            if [[ "$line" =~ ^[[:space:]]*brew[[:space:]]+[\"\'"]([^\"\']+)[\"\'"] ]]; then
                EXPECTED_FORMULAS+=("${BASH_REMATCH[1]}")
            # Parse cask "package" or cask 'package'
            elif [[ "$line" =~ ^[[:space:]]*cask[[:space:]]+[\"\'"]([^\"\']+)[\"\'"] ]]; then
                EXPECTED_CASKS+=("${BASH_REMATCH[1]}")
            fi
        done < "$brewfile"
    fi
}

echo "Reading Brewfiles..."

# 1. Brewfile.shared (always)
parse_brewfile "$DOTFILES_DIR/Brewfile.shared"

# 2. aggressive/Brewfile (if aggressive mode)
if [[ "$MACHINE_TYPE" == "aggressive" && -f "$DOTFILES_DIR/aggressive/Brewfile" ]]; then
    parse_brewfile "$DOTFILES_DIR/aggressive/Brewfile"
fi

# 3. Private shared Brewfile
if has_private_overlay; then
    local_private_brewfile="$(get_private_brewfile)"
    if [[ -f "$local_private_brewfile" ]]; then
        parse_brewfile "$local_private_brewfile"
    fi

    # 4. Profile-specific Brewfile
    if [[ -n "$MACHINE_PROFILE" ]]; then
        local_profile_brewfile="$(get_profile_brewfile "$MACHINE_PROFILE")"
        if [[ -f "$local_profile_brewfile" ]]; then
            parse_brewfile "$local_profile_brewfile"
        fi
    fi
fi

echo "  Expected formulas: ${#EXPECTED_FORMULAS[@]}"
echo "  Expected casks: ${#EXPECTED_CASKS[@]}"

# ============================================================================
# Get installed packages
# ============================================================================

echo ""
echo "Getting installed packages..."

INSTALLED_FORMULAS=($(brew list --formula -1 2>/dev/null))
INSTALLED_CASKS=($(brew list --cask -1 2>/dev/null))

echo "  Installed formulas: ${#INSTALLED_FORMULAS[@]}"
echo "  Installed casks: ${#INSTALLED_CASKS[@]}"

# ============================================================================
# Find unlisted packages
# ============================================================================

is_in_array() {
    local needle="$1"
    shift
    local item
    for item in "$@"; do
        [[ "$item" == "$needle" ]] && return 0
    done
    return 1
}

UNLISTED_FORMULAS=()
UNLISTED_CASKS=()

for formula in "${INSTALLED_FORMULAS[@]}"; do
    if ! is_in_array "$formula" "${EXPECTED_FORMULAS[@]}"; then
        UNLISTED_FORMULAS+=("$formula")
    fi
done

for cask in "${INSTALLED_CASKS[@]}"; do
    if ! is_in_array "$cask" "${EXPECTED_CASKS[@]}"; then
        UNLISTED_CASKS+=("$cask")
    fi
done

# ============================================================================
# Report findings
# ============================================================================

echo ""
echo "============================================"
echo "  Packages not in Brewfiles"
echo "============================================"

if [[ ${#UNLISTED_FORMULAS[@]} -eq 0 && ${#UNLISTED_CASKS[@]} -eq 0 ]]; then
    echo ""
    echo "All installed packages are in your Brewfiles!"
    exit 0
fi

if [[ ${#UNLISTED_FORMULAS[@]} -gt 0 ]]; then
    echo ""
    echo "Formulas (${#UNLISTED_FORMULAS[@]}):"
    for formula in "${UNLISTED_FORMULAS[@]}"; do
        echo "  - $formula"
    done
fi

if [[ ${#UNLISTED_CASKS[@]} -gt 0 ]]; then
    echo ""
    echo "Casks (${#UNLISTED_CASKS[@]}):"
    for cask in "${UNLISTED_CASKS[@]}"; do
        echo "  - $cask"
    done
fi

# ============================================================================
# Optionally remove
# ============================================================================

if [[ "$REMOVE" == false ]]; then
    echo ""
    echo "============================================"
    echo "  Run with --remove to uninstall these"
    echo "============================================"
    exit 0
fi

echo ""
echo "============================================"
echo "  Removing unlisted packages"
echo "============================================"

if [[ "$FORCE" == false ]]; then
    echo ""
    echo "This will uninstall the packages listed above."
    read -p "Continue? [y/N] " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        echo "Aborted."
        exit 0
    fi
fi

if [[ ${#UNLISTED_FORMULAS[@]} -gt 0 ]]; then
    echo ""
    echo "Removing formulas..."
    for formula in "${UNLISTED_FORMULAS[@]}"; do
        echo "  Removing $formula..."
        brew uninstall --ignore-dependencies "$formula" 2>/dev/null || echo "    (failed or has dependents)"
    done
fi

if [[ ${#UNLISTED_CASKS[@]} -gt 0 ]]; then
    echo ""
    echo "Removing casks..."
    for cask in "${UNLISTED_CASKS[@]}"; do
        echo "  Removing $cask..."
        brew uninstall --cask "$cask" 2>/dev/null || echo "    (failed)"
    done
fi

echo ""
echo "============================================"
echo "  Cleanup complete!"
echo "============================================"
echo ""
echo "Note: Some packages may be dependencies of others."
echo "Run 'brew autoremove' to remove unused dependencies."
