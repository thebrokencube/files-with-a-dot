#!/bin/bash
# lib/paths.sh - Path resolution utilities
#
# Usage: source "$(dirname "$0")/lib/paths.sh"

# Check if a path is managed by dotfiles (points to DOTFILES_DIR)
# Resolves both paths to handle DOTFILES_DIR being a symlink
is_ours() {
    local path="$1"
    local real_path real_dotfiles target
    real_dotfiles=$(realpath "$DOTFILES_DIR" 2>/dev/null || echo "$DOTFILES_DIR")
    if [[ -L "$path" ]]; then
        target=$(readlink "$path")
        [[ "$target" != /* ]] && target="$(dirname "$path")/$target"
        real_path=$(realpath "$target" 2>/dev/null || echo "$target")
    elif [[ -e "$path" ]]; then
        real_path=$(realpath "$path" 2>/dev/null || echo "")
    else
        return 1
    fi
    [[ -n "$real_path" ]] && { [[ "$real_path" == "$real_dotfiles" ]] || [[ "$real_path" == "$real_dotfiles/"* ]]; }
}

# Check if path is a symlink NOT managed by dotfiles
is_foreign_symlink() {
    local path="$1"
    [[ -L "$path" ]] && ! is_ours "$path"
}

# Extract source path from symlink_map line
get_source() {
    local line="$1"
    echo "$line" | cut -d':' -f1
}

# Extract destination path from symlink_map line and expand $HOME
get_dest() {
    local line="$1"
    local dest
    dest=$(echo "$line" | cut -d':' -f2-)
    echo "${dest/\$HOME/$HOME}"
}

# Safely resolve real path
resolve_path() {
    local path="$1"
    realpath "$path" 2>/dev/null || echo ""
}

RETIRED_LINK_PAIRS=(
    "$HOME/.claude/CLAUDE.md|configs/base/claude/.claude/CLAUDE.md"
    "$HOME/.claude/rules/code-comments.md|configs/base/claude/.claude/rules/code-comments.md"
    "$HOME/.claude/rules/dotfiles-awareness.md|configs/base/claude/.claude/rules/dotfiles-awareness.md"
    "$HOME/.claude/rules/isolated-checkouts.md|configs/base/claude/.claude/rules/isolated-checkouts.md"
    "$HOME/.claude/rules/no-memory-files.md|configs/base/claude/.claude/rules/no-memory-files.md"
    "$HOME/.claude/rules/working-discipline.md|configs/base/claude/.claude/rules/working-discipline.md"
    "$HOME/.claude/rules/workflow.md|configs/base/claude/.claude/rules/workflow.md"
    "$HOME/.claude/rules/writing-structure.md|configs/base/claude/.claude/rules/writing-structure.md"
    "$HOME/.claude/rules/concision.md|configs/base/claude/.claude/output-styles/Concise.md"
    "$HOME/.claude/references/isolated-checkouts-runbook.md|configs/base/claude/.claude/references/isolated-checkouts-runbook.md"
)
RETIRED_LINK_CLEANUP_PLAN=()

retired_link_matches() {
    local destination="$1" source="$2" raw_target root resolved_root
    [[ -L "$destination" ]] || return 1
    raw_target=$(readlink "$destination") || return 1

    resolved_root=$(resolve_path "$DOTFILES_DIR")
    for root in "$DOTFILES_DIR" "$resolved_root" "$HOME/.dotfiles"; do
        [[ -n "$root" && "$raw_target" == "$root/$source" ]] && return 0
    done
    return 1
}

plan_retired_link_cleanup() {
    RETIRED_LINK_CLEANUP_PLAN=()
    local pair destination source
    for pair in "${RETIRED_LINK_PAIRS[@]}"; do
        destination="${pair%%|*}"
        source="${pair#*|}"
        if retired_link_matches "$destination" "$source"; then
            RETIRED_LINK_CLEANUP_PLAN+=("$pair")
            ACTIONS+=("Remove retired link ${destination#"$HOME"/}")
        fi
    done
}

cleanup_retired_links() {
    local pair destination source
    for pair in "${RETIRED_LINK_CLEANUP_PLAN[@]}"; do
        destination="${pair%%|*}"
        source="${pair#*|}"
        if retired_link_matches "$destination" "$source"; then
            /bin/rm -f "$destination"
            echo "  Removed retired link ${destination#"$HOME"/}"
        fi
    done
}
