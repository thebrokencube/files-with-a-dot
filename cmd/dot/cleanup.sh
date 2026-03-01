#!/bin/bash
# cleanup.sh - System cleanup (aggressive vs conservative mode)
#
# Aggressive mode: This repo is the source of truth, so aggressively clean
#                  packages not in Brewfiles, old downloads, caches, etc.
#
# Conservative mode: Other tools may manage packages, so just show what could
#                    be cleaned but don't automatically remove anything.
#
# Usage:
#   ./cleanup.sh             # Show opportunities, prompt to execute
#   ./cleanup.sh --dry-run   # Show opportunities only
#   ./cleanup.sh --force     # Execute without prompts

set -e

# ── Setup ────────────────────────────────────────────────────────────────────

DOTFILES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=lib/colors.sh
source "$DOTFILES_DIR/lib/colors.sh"
# shellcheck source=lib/logging.sh
source "$DOTFILES_DIR/lib/logging.sh"
# shellcheck source=lib/config.sh
source "$DOTFILES_DIR/lib/config.sh"
# shellcheck source=lib/prompt.sh
source "$DOTFILES_DIR/lib/prompt.sh"
# shellcheck source=lib/private.sh
source "$DOTFILES_DIR/lib/private.sh"
# shellcheck source=lib/brew.sh
source "$DOTFILES_DIR/lib/brew.sh"

init_dotfiles_vars

IS_AGGRESSIVE=$([[ "$MACHINE_TYPE" == "aggressive" ]] && echo true || echo false)

# ── Parse flags ──────────────────────────────────────────────────────────────

FORCE=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --force) FORCE=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        --help|-h)
            echo "Usage: ./cleanup.sh [OPTIONS]"
            echo ""
            echo "System cleanup:"
            echo "  - Homebrew packages/casks not in Brewfiles"
            echo "  - Unused dependencies"
            echo "  - Old Homebrew downloads"
            echo ""
            echo "Aggressive mode: Aggressive cleanup with --zap (this repo is source of truth)"
            echo "Conservative mode: Show opportunities only (Homebrew cache cleanup is safe)"
            echo ""
            echo "Options:"
            echo "  --dry-run  Show opportunities only"
            echo "  --force    Run without any prompts"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

# ── Detect ───────────────────────────────────────────────────────────────────

build_merged_brewfile

echo "============================================"
echo "  System Cleanup $DOTFILES_VERSION ($MACHINE_TYPE mode)"
echo "============================================"
if [[ "$IS_AGGRESSIVE" == false ]]; then
    warn "Conservative mode: Only Homebrew cache cleanup is safe."
    echo "Packages below may be managed by other tools - shown for reference only."
fi
echo ""

echo "Current disk usage:"
df -h / | tail -1 | awk '{print "  Used: " $3 " / " $2 " (" $5 ")"}'
echo ""

# 1. Packages not in Brewfiles
section "1. Homebrew packages not in Brewfiles:"
echo ""
HAS_BUNDLE_CLEANUP=false
if detect_brew_cleanup; then
    echo "$BREW_CLEANUP_OUTPUT" | grep -v "^Would uninstall" | grep -v "^$" | sed 's/^/  /'
    HAS_BUNDLE_CLEANUP=true
else
    echo "  (none)"
fi

# 2. Unused dependencies
section "2. Unused Homebrew dependencies:"
HAS_UNUSED=false
if detect_brew_autoremove; then
    echo "$BREW_AUTOREMOVE_OUTPUT" | grep "Would uninstall" | sed 's/^/  /'
    HAS_UNUSED=true
else
    echo "  (none)"
fi

# 3. Old downloads
section "3. Old Homebrew downloads:"
HAS_CACHE=false
detect_brew_cache && HAS_CACHE=true
echo "  Cache size: ${BREW_CACHE_SIZE:-0B}"

# ── Decide ───────────────────────────────────────────────────────────────────

HAS_PACKAGE_WORK=false
[[ "$HAS_BUNDLE_CLEANUP" == true || "$HAS_UNUSED" == true ]] && HAS_PACKAGE_WORK=true

HAS_WORK=false
[[ "$HAS_PACKAGE_WORK" == true || "$HAS_CACHE" == true ]] && HAS_WORK=true

if [[ "$HAS_WORK" == false ]]; then
    echo ""
    echo "Nothing to clean!"
    exit 0
fi

# Dry-run: stop after showing
if [[ "$DRY_RUN" == true ]]; then
    exit 0
fi

echo ""
echo "============================================"
if [[ "$IS_AGGRESSIVE" == false ]]; then
    echo "Conservative mode: Only Homebrew cache will be cleaned."
    echo "(Packages shown above are managed by other tools)"
fi

if ! confirm ${FORCE:+-f} "Run cleanup?" "no"; then
    echo ""
    echo "Run later with: dot clean"
    exit 0
fi

# ── Execute ──────────────────────────────────────────────────────────────────

echo ""
echo "============================================"
echo "  Executing Cleanup"
echo "============================================"
echo ""

# 1. Bundle cleanup
if [[ "$HAS_BUNDLE_CLEANUP" == true ]]; then
    info "Removing packages not in Brewfiles..."
    execute_brew_bundle_cleanup "$IS_AGGRESSIVE"
    echo ""
fi

# 2. Unused dependencies
if [[ "$HAS_UNUSED" == true ]]; then
    info "Removing unused dependencies..."
    execute_brew_autoremove "$IS_AGGRESSIVE"
    echo ""
fi

# 3. Cache cleanup (safe for all)
info "Cleaning Homebrew cache..."
execute_brew_cache_cleanup
echo ""

echo "============================================"
echo -e "  ${GREEN}Cleanup Complete!${NC}"
echo "============================================"
echo ""
echo "Final disk usage:"
df -h / | tail -1 | awk '{print "  Used: " $3 " / " $2 " (" $5 ")"}'
