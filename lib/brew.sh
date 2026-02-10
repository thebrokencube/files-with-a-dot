#!/bin/bash
# lib/brew.sh - Homebrew operations (Brewfile merging, cleanup detection/execution)
#
# Usage: source "$(dirname "$0")/lib/brew.sh"
# Requires: DOTFILES_DIR, MACHINE_TYPE, MACHINE_PROFILE to be set
# Requires: lib/private.sh to be sourced first

# ── Brewfile merging ─────────────────────────────────────────────────────────

# Build a merged Brewfile from all sources into a temp file.
# Sets: MERGED_BREWFILE (path to temp file)
# Registers an EXIT trap to clean up the temp file.
build_merged_brewfile() {
    MERGED_BREWFILE=$(mktemp)
    trap 'rm -f "$MERGED_BREWFILE"' EXIT

    echo "# Merged Brewfile for cleanup (auto-generated)" > "$MERGED_BREWFILE"
    echo "" >> "$MERGED_BREWFILE"

    # 1. Brewfile.shared (always)
    if [[ -f "$DOTFILES_DIR/Brewfile.shared" ]]; then
        echo "# From Brewfile.shared" >> "$MERGED_BREWFILE"
        cat "$DOTFILES_DIR/Brewfile.shared" >> "$MERGED_BREWFILE"
        echo "" >> "$MERGED_BREWFILE"
    fi

    # 2. aggressive/Brewfile (if aggressive mode)
    if [[ "$MACHINE_TYPE" == "aggressive" && -f "$DOTFILES_DIR/aggressive/Brewfile" ]]; then
        echo "# From aggressive/Brewfile" >> "$MERGED_BREWFILE"
        cat "$DOTFILES_DIR/aggressive/Brewfile" >> "$MERGED_BREWFILE"
        echo "" >> "$MERGED_BREWFILE"
    fi

    # 3. Private Brewfiles
    if has_private_overlay; then
        local private_brewfile
        private_brewfile="$(get_private_brewfile)"
        if [[ -f "$private_brewfile" ]]; then
            echo "# From private/Brewfile" >> "$MERGED_BREWFILE"
            cat "$private_brewfile" >> "$MERGED_BREWFILE"
            echo "" >> "$MERGED_BREWFILE"
        fi

        # 4. Profile-specific Brewfile
        if [[ -n "$MACHINE_PROFILE" ]]; then
            local profile_brewfile
            profile_brewfile="$(get_profile_brewfile "$MACHINE_PROFILE")"
            if [[ -f "$profile_brewfile" ]]; then
                echo "# From private/$MACHINE_PROFILE/Brewfile" >> "$MERGED_BREWFILE"
                cat "$profile_brewfile" >> "$MERGED_BREWFILE"
                echo "" >> "$MERGED_BREWFILE"
            fi
        fi
    fi
}

# ── Detection (no output, set globals) ───────────────────────────────────────

# Detect packages/casks not in the merged Brewfile.
# Requires: MERGED_BREWFILE to be set (call build_merged_brewfile first).
# Sets: BREW_CLEANUP_OUTPUT
# Returns: 0 if there are packages to clean, 1 if clean
detect_brew_cleanup() {
    BREW_CLEANUP_OUTPUT=$(brew bundle cleanup --file="$MERGED_BREWFILE" 2>&1 || true)
    echo "$BREW_CLEANUP_OUTPUT" | grep -q "Would uninstall"
}

# Detect unused Homebrew dependencies.
# Sets: BREW_AUTOREMOVE_OUTPUT, BREW_UNUSED_COUNT
# Returns: 0 if there are unused deps, 1 if clean
detect_brew_autoremove() {
    BREW_AUTOREMOVE_OUTPUT=$(brew autoremove --dry-run 2>&1 || true)
    BREW_UNUSED_COUNT=$(echo "$BREW_AUTOREMOVE_OUTPUT" | grep -c "Would uninstall" || true)
    [[ "$BREW_UNUSED_COUNT" -gt 0 ]]
}

# Detect old Homebrew cache.
# Sets: BREW_CACHE_DIR, BREW_CACHE_SIZE, BREW_CACHE_BYTES
# Returns: 0 if cache is significant (>100MB), 1 otherwise
# shellcheck disable=SC2034  # Globals consumed by callers
detect_brew_cache() {
    BREW_CACHE_DIR=$(brew --cache)
    if [[ -d "$BREW_CACHE_DIR" ]]; then
        BREW_CACHE_SIZE=$(du -sh "$BREW_CACHE_DIR" 2>/dev/null | awk '{print $1}')
        BREW_CACHE_BYTES=$(du -sk "$BREW_CACHE_DIR" 2>/dev/null | awk '{print $1}')
        [[ "${BREW_CACHE_BYTES:-0}" -gt 102400 ]]
    else
        BREW_CACHE_SIZE="0B"
        BREW_CACHE_BYTES=0
        return 1
    fi
}

# ── Execution ────────────────────────────────────────────────────────────────

# Execute brew bundle cleanup.
# Args: $1 = is_aggressive ("true"/"false")
# Requires: MERGED_BREWFILE to be set.
execute_brew_bundle_cleanup() {
    local is_aggressive="${1:-false}"
    if [[ "$is_aggressive" == "true" ]]; then
        brew bundle cleanup --file="$MERGED_BREWFILE" --force --zap || true
    else
        echo "Conservative mode: Not removing packages (other tools may manage them)"
        echo "  To manually remove: brew bundle cleanup --file=<merged> --force"
    fi
}

# Execute brew autoremove.
# Args: $1 = is_aggressive ("true"/"false")
execute_brew_autoremove() {
    local is_aggressive="${1:-false}"
    if [[ "$is_aggressive" == "true" ]]; then
        brew autoremove || true
    else
        echo "Conservative mode: Not auto-removing dependencies"
        echo "  To manually remove: brew autoremove"
    fi
}

# Execute brew cache cleanup (safe for all modes).
execute_brew_cache_cleanup() {
    brew cleanup -s || true
}
