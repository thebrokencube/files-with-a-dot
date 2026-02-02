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
#   ./cleanup.sh              # Show cleanup opportunities (interactive prompt)
#   ./cleanup.sh --confirmed  # Show + execute (user already confirmed)
#   ./cleanup.sh --execute    # Run cleanup (with confirmation)
#   ./cleanup.sh --force      # Run without any confirmation

set -e

EXECUTE=false
FORCE=false
CONFIRMED=false  # Set when called from sync.sh/bootstrap.sh (user already said yes)

while [[ $# -gt 0 ]]; do
    case $1 in
        --confirmed) CONFIRMED=true; EXECUTE=true; shift ;;
        --execute) EXECUTE=true; shift ;;
        --force) EXECUTE=true; FORCE=true; shift ;;
        --help|-h)
            echo "Usage: ./cleanup.sh [OPTIONS]"
            echo ""
            echo "System cleanup:"
            echo "  - Homebrew packages/casks not in Brewfiles"
            echo "  - Unused dependencies"
            echo "  - Old Homebrew downloads"
            echo "  - Downloads folder (files older than 30 days, aggressive mode only)"
            echo "  - System caches (aggressive mode only)"
            echo ""
            echo "Aggressive mode: Aggressive cleanup with --zap (this repo is source of truth)"
            echo "Conservative mode: Show opportunities only (Homebrew cache cleanup is safe)"
            echo ""
            echo "Options:"
            echo "  --confirmed  Show opportunities and execute (called from sync/bootstrap)"
            echo "  --execute    Run cleanup (with confirmation)"
            echo "  --force      Run without any confirmation"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MACHINE_FILE="$SCRIPT_DIR/.machine"
MACHINE_TYPE="aggressive"
[[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")

GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
NC='\033[0m'

IS_AGGRESSIVE=$([[ "$MACHINE_TYPE" == "aggressive" ]] && echo true || echo false)
VERSION=$(cd "$SCRIPT_DIR" && git describe --tags --always 2>/dev/null || echo "unknown")

echo "============================================"
echo "  System Cleanup $VERSION ($MACHINE_TYPE mode)"
echo "============================================"
if [[ "$IS_AGGRESSIVE" == false ]]; then
    echo -e "${YELLOW}Conservative mode: Only Homebrew cache cleanup is safe.${NC}"
    echo "Packages below may be managed by other tools - shown for reference only."
fi
echo ""

# ============================================================================
# Check disk space
# ============================================================================

echo "Current disk usage:"
df -h / | tail -1 | awk '{print "  Used: " $3 " / " $2 " (" $5 ")"}'
echo ""

# ============================================================================
# 1. Homebrew bundle cleanup
# ============================================================================

echo -e "${CYAN}1. Homebrew packages not in Brewfiles:${NC}"
echo ""

# Check what would be removed
if brew bundle cleanup --file="$SCRIPT_DIR/Brewfile.shared" 2>&1 | grep -q "Skipping"; then
    brew bundle cleanup --file="$SCRIPT_DIR/Brewfile.shared" 2>&1 | grep "Skipping" | sed 's/^/  /'
    HAS_BUNDLE_CLEANUP=true
else
    echo "  (none)"
    HAS_BUNDLE_CLEANUP=false
fi

if [[ -f "$SCRIPT_DIR/aggressive/Brewfile" ]]; then
    if brew bundle cleanup --file="$SCRIPT_DIR/aggressive/Brewfile" 2>&1 | grep -q "Skipping"; then
        brew bundle cleanup --file="$SCRIPT_DIR/aggressive/Brewfile" 2>&1 | grep "Skipping" | sed 's/^/  /'
        HAS_BUNDLE_CLEANUP=true
    fi
fi

# ============================================================================
# 2. Unused dependencies
# ============================================================================

echo ""
echo -e "${CYAN}2. Unused Homebrew dependencies:${NC}"
UNUSED=$(brew autoremove --dry-run 2>&1 | grep "Would uninstall" | wc -l | tr -d ' ')
if [[ "$UNUSED" -gt 0 ]]; then
    brew autoremove --dry-run 2>&1 | grep "Would uninstall" | sed 's/^/  /'
else
    echo "  (none)"
fi

# ============================================================================
# 3. Old Homebrew downloads
# ============================================================================

echo ""
echo -e "${CYAN}3. Old Homebrew downloads:${NC}"
BREW_CACHE=$(brew --cache)
if [[ -d "$BREW_CACHE" ]]; then
    CACHE_SIZE=$(du -sh "$BREW_CACHE" 2>/dev/null | awk '{print $1}')
    echo "  Cache size: $CACHE_SIZE"
else
    echo "  (none)"
fi

# ============================================================================
# 4. Downloads folder (home only)
# ============================================================================

if [[ "$IS_AGGRESSIVE" == true ]]; then
    echo ""
    echo -e "${CYAN}4. Downloads folder (files older than 30 days):${NC}"
    OLD_DOWNLOADS=$(find "$HOME/Downloads" -type f -mtime +30 2>/dev/null | wc -l | tr -d ' ')
    if [[ "$OLD_DOWNLOADS" -gt 0 ]]; then
        OLD_SIZE=$(find "$HOME/Downloads" -type f -mtime +30 -print0 2>/dev/null | du -ch --files0-from=- 2>/dev/null | tail -1 | awk '{print $1}')
        echo "  $OLD_DOWNLOADS files ($OLD_SIZE)"
        find "$HOME/Downloads" -type f -mtime +30 2>/dev/null | head -10 | sed 's/^/    /'
        [[ "$OLD_DOWNLOADS" -gt 10 ]] && echo "    ... and $(($OLD_DOWNLOADS - 10)) more"
    else
        echo "  (none)"
    fi
else
    OLD_DOWNLOADS=0
fi

# ============================================================================
# 5. System caches (home only)
# ============================================================================

if [[ "$IS_AGGRESSIVE" == true ]]; then
    echo ""
    echo -e "${CYAN}5. System caches:${NC}"
    if [[ -d "$HOME/Library/Caches" ]]; then
        CACHE_SIZE=$(du -sh "$HOME/Library/Caches" 2>/dev/null | awk '{print $1}')
        echo "  ~/Library/Caches: $CACHE_SIZE"
    fi
fi

# ============================================================================
# Execute cleanup
# ============================================================================

# Check if there's meaningful work to do
HAS_PACKAGE_WORK=false
[[ "$HAS_BUNDLE_CLEANUP" == true ]] && HAS_PACKAGE_WORK=true
[[ "$UNUSED" -gt 0 ]] && HAS_PACKAGE_WORK=true

HAS_CACHE_WORK=false
if [[ -d "$BREW_CACHE" ]]; then
    # Only count cache as "work" if it's > 100MB
    CACHE_BYTES=$(du -sk "$BREW_CACHE" 2>/dev/null | awk '{print $1}')
    [[ "${CACHE_BYTES:-0}" -gt 102400 ]] && HAS_CACHE_WORK=true
fi

HAS_WORK=false
[[ "$HAS_PACKAGE_WORK" == true ]] && HAS_WORK=true
[[ "$HAS_CACHE_WORK" == true ]] && HAS_WORK=true
[[ "$OLD_DOWNLOADS" -gt 0 ]] && HAS_WORK=true

# Interactive prompt (only when run standalone without flags)
if [[ "$EXECUTE" == false ]]; then
    echo ""
    if [[ "$HAS_WORK" == true ]]; then
        echo "============================================"
        if [[ "$IS_AGGRESSIVE" == true ]]; then
            read -p "Run cleanup now? [y/N] " run_now
        else
            echo "Conservative mode: Only Homebrew cache will be cleaned."
            echo "(Packages shown above are managed by other tools)"
            read -p "Clean Homebrew cache now? [y/N] " run_now
        fi
        if [[ "$run_now" == "y" || "$run_now" == "Y" ]]; then
            EXECUTE=true
            CONFIRMED=true  # User just confirmed, skip secondary prompt
        else
            echo ""
            echo "Run later with: ./cleanup.sh --execute"
            exit 0
        fi
    else
        echo "Nothing to clean!"
        exit 0
    fi
fi

echo ""
echo "============================================"
echo "  Executing Cleanup"
echo "============================================"
echo ""

# Safety confirmation for --execute flag (not needed if --confirmed or --force)
if [[ "$FORCE" == false && "$CONFIRMED" == false ]]; then
    if [[ "$IS_AGGRESSIVE" == true ]]; then
        read -p "This will remove packages, casks (with --zap), and old files. Continue? [y/N] " confirm
    else
        read -p "This will clean Homebrew cache. Continue? [y/N] " confirm
    fi
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        echo "Aborted."
        exit 0
    fi
fi

# 1. Bundle cleanup (--zap on home, conservative on work)
if [[ "$HAS_BUNDLE_CLEANUP" == true ]]; then
    if [[ "$IS_AGGRESSIVE" == true ]]; then
        echo "Removing packages not in Brewfiles (with --zap)..."
        brew bundle cleanup --file="$SCRIPT_DIR/Brewfile.shared" --force --zap || true
        [[ -f "$SCRIPT_DIR/aggressive/Brewfile" ]] && brew bundle cleanup --file="$SCRIPT_DIR/aggressive/Brewfile" --force --zap || true
    else
        echo "Conservative mode: Not removing packages (other tools may manage them)"
        echo "  To manually remove: brew bundle cleanup --file=... --force"
    fi
    echo ""
fi

# 2. Remove unused dependencies (aggressive only)
if [[ "$UNUSED" -gt 0 ]]; then
    if [[ "$IS_AGGRESSIVE" == true ]]; then
        echo "Removing unused dependencies..."
        brew autoremove || true
    else
        echo "Conservative mode: Not auto-removing dependencies"
        echo "  To manually remove: brew autoremove"
    fi
    echo ""
fi

# 3. Clean Homebrew cache (safe for all)
echo "Cleaning Homebrew cache..."
brew cleanup -s || true
echo ""

# 4. Clean old downloads (home only)
if [[ "$IS_AGGRESSIVE" == true && "$OLD_DOWNLOADS" -gt 0 ]]; then
    echo "Old downloads can be reviewed manually in ~/Downloads"
    if [[ "$FORCE" == true ]]; then
        echo "Moving old downloads to trash..."
        find "$HOME/Downloads" -type f -mtime +30 -exec trash {} \; 2>/dev/null || \
        find "$HOME/Downloads" -type f -mtime +30 -delete 2>/dev/null
    fi
    echo ""
fi

# 5. Clean system caches (conservative even on home)
if [[ "$IS_AGGRESSIVE" == true ]]; then
    echo "System caches kept (too risky to auto-clean)"
    echo "  To clean manually: rm -rf ~/Library/Caches/*"
    echo ""
fi

echo "============================================"
echo -e "  ${GREEN}Cleanup Complete!${NC}"
echo "============================================"
echo ""
echo "Final disk usage:"
df -h / | tail -1 | awk '{print "  Used: " $3 " / " $2 " (" $5 ")"}'
