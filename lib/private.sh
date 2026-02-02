#!/bin/bash
# lib/private.sh - Private overlay support
#
# Usage: source "$(dirname "$0")/lib/private.sh"
# Requires: DOTFILES_DIR to be set

PRIVATE_DIR="$HOME/.dotfiles.private"

# Check if private overlay exists
has_private_overlay() {
    [[ -d "$PRIVATE_DIR" ]]
}

# Check if private overlay is a git repo
has_private_git() {
    [[ -d "$PRIVATE_DIR/.git" ]]
}

# Get profile-specific directory path
get_profile_dir() {
    local profile="$1"
    echo "$PRIVATE_DIR/$profile"
}

# Check if profile directory exists in private overlay
has_profile_dir() {
    local profile="$1"
    [[ -d "$(get_profile_dir "$profile")" ]]
}

# Get private symlink map path (shared)
get_private_symlink_map() {
    echo "$PRIVATE_DIR/symlink_map.txt"
}

# Get profile-specific symlink map path
get_profile_symlink_map() {
    local profile="$1"
    echo "$(get_profile_dir "$profile")/symlink_map.txt"
}

# Get private Brewfile path (shared)
get_private_brewfile() {
    echo "$PRIVATE_DIR/Brewfile"
}

# Get profile-specific Brewfile path
get_profile_brewfile() {
    local profile="$1"
    echo "$(get_profile_dir "$profile")/Brewfile"
}

# Check private symlink (for analysis phase)
check_private_symlink() {
    local source="$1"
    local dest="$2"
    local base_dir="${3:-$PRIVATE_DIR}"
    local name=$(basename "$source")
    local source_path="$base_dir/$source"

    if [[ ! -e "$source_path" ]]; then
        return 0
    fi

    if [[ -L "$dest" ]]; then
        if [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path" 2>/dev/null)" ]]; then
            ALREADY_DONE+=("$name (private)")
        else
            # Check if it's linked to public version - private should override
            if is_ours "$dest"; then
                ACTIONS+=("Override $name with private version")
            else
                FRICTIONS+=("$dest is a symlink to $(readlink "$dest"), conflicts with private $name")
            fi
        fi
    elif [[ -e "$dest" ]]; then
        ACTIONS+=("Link private $name (existing $dest will be backed up)")
        [[ "$NO_BACKUP" != true ]] && WILL_BACKUP+=("$dest (private $name)")
    else
        ACTIONS+=("Link private $name")
    fi
    return 0
}

# Create a private symlink (can override public symlinks)
create_private_symlink() {
    local source="$1"
    local dest="$2"
    local base_dir="${3:-$PRIVATE_DIR}"
    local name=$(basename "$source")
    local source_path="$base_dir/$source"

    # Skip if source doesn't exist
    if [[ ! -e "$source_path" ]]; then
        return
    fi

    # Skip if already correctly linked
    if [[ -L "$dest" ]] && [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path")" ]]; then
        return
    fi

    # Backup and remove existing (even if it's our public symlink)
    if [[ -e "$dest" || -L "$dest" ]]; then
        # Only backup if it's not one of our symlinks
        if ! is_ours "$dest" && [[ ! -L "$dest" || "$(realpath "$dest" 2>/dev/null)" != "$source_path" ]]; then
            backup_file "$dest" "private_symlink_target"
        fi
        rm -rf "$dest"
    fi

    # Create parent directory if needed
    local parent_dir=$(dirname "$dest")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir"
    fi

    # Create symlink
    ln -s "$source_path" "$dest"
    echo "  $name -> $dest (private)"
}

# Apply private symlinks from a symlink_map file
apply_private_symlinks() {
    local symlink_map="$1"
    local base_dir="${2:-$PRIVATE_DIR}"

    [[ ! -f "$symlink_map" ]] && return

    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        local source=$(get_source "$line")
        local dest=$(get_dest "$line")
        create_private_symlink "$source" "$dest" "$base_dir"
    done < "$symlink_map"
}
