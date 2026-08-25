#!/bin/bash
# lib/symlinks.sh - Symlink creation and validation
#
# Usage: source "$(dirname "$0")/lib/symlinks.sh"
# Requires: lib/colors.sh, lib/logging.sh, lib/paths.sh, lib/backup.sh

# Check symlink state for analysis phase
check_symlink() {
    local source="$1"
    local dest="$2"
    local name
    name="${dest#"$HOME"/}"
    local source_path="$DOTFILES_DIR/$source"

    if [[ ! -e "$source_path" ]]; then
        return 0
    fi

    if [[ -L "$dest" ]]; then
        if [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path" 2>/dev/null)" ]]; then
            DONE_SYMLINKS+=("$name")
        elif is_ours "$dest"; then
            ACTIONS+=("Relink $name to its current managed source")
        else
            FRICTIONS+=("$dest is a symlink to $(readlink "$dest"), conflicts with $name")
        fi
    elif [[ -e "$dest" ]]; then
        ACTIONS+=("Link $name (existing $dest will be backed up)")
        [[ "$NO_BACKUP" != true ]] && WILL_BACKUP+=("$dest ($name)")
    else
        ACTIONS+=("Link $name")
    fi
    return 0
}

# Create a single symlink
create_symlink() {
    local source="$1"
    local dest="$2"
    local source_path="$DOTFILES_DIR/$source"
    local name
    name="${dest#"$HOME"/}"

    # Skip if source doesn't exist
    if [[ ! -e "$source_path" ]]; then
        return
    fi

    # Skip if already correctly linked
    if [[ -L "$dest" ]] && [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path")" ]]; then
        return
    fi

    # Backup and remove existing
    if [[ -e "$dest" || -L "$dest" ]]; then
        if ! is_ours "$dest"; then
            backup_file "$dest" "symlink_target"
        fi
        rm -rf "$dest"
    fi

    # Create parent directory if needed
    local parent_dir
    parent_dir=$(dirname "$dest")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir"
    fi

    # Create symlink
    ln -s "$source_path" "$dest"
    echo "  $name -> $dest"
}

# Apply all symlinks from symlink_map.txt
apply_symlinks() {
    local symlink_map="$1"

    [[ ! -f "$symlink_map" ]] && return

    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        local source
        source=$(get_source "$line")
        local dest
        dest=$(get_dest "$line")
        create_symlink "$source" "$dest"
    done < "$symlink_map"
}

# Create ~/.dotfiles symlink if needed
create_dotfiles_symlink() {
    if [[ "$DOTFILES_DIR" == "$HOME/.dotfiles" ]]; then
        return 0
    elif [[ ! -L "$HOME/.dotfiles" && ! -e "$HOME/.dotfiles" ]]; then
        echo "Creating ~/.dotfiles symlink..."
        ln -s "$DOTFILES_DIR" "$HOME/.dotfiles"
    fi
}
