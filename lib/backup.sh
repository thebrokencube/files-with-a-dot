#!/bin/bash
# lib/backup.sh - Backup and restore operations
#
# Usage: source "$(dirname "$0")/lib/backup.sh"
# Requires: lib/colors.sh, lib/logging.sh, lib/paths.sh
# Requires: DOTFILES_DIR to be set
#
# Callers may set NO_BACKUP=true before calling to disable all backup operations.
# All other variables (BACKUP_DIR, BACKUP_MANIFEST) are managed internally.

# Initialize backup system (sets defaults, creates dirs)
init_backup() {
    BACKUP_DIR="${BACKUP_DIR:-$DOTFILES_DIR/.backup}"
    BACKUP_MANIFEST="${BACKUP_MANIFEST:-$BACKUP_DIR/manifest}"

    [[ "${NO_BACKUP:-false}" == true ]] && return
    mkdir -p "$BACKUP_DIR"
    if [[ ! -f "$BACKUP_MANIFEST" ]]; then
        echo "# Dotfiles backup manifest" > "$BACKUP_MANIFEST"
        echo "# Format: original_path|backup_name|timestamp|type" >> "$BACKUP_MANIFEST"
    fi
}

# Backup a file before replacing
backup_file() {
    local original="$1"
    local type="${2:-unknown}"

    [[ "${NO_BACKUP:-false}" == true ]] && return 0
    [[ ! -e "$original" ]] && return 0

    # Don't backup our own symlinks
    if [[ -L "$original" ]]; then
        local target
        target=$(realpath "$original" 2>/dev/null || echo "")
        [[ "$target" == "$DOTFILES_DIR"* ]] && return 0
    fi

    local backup_name
    backup_name=$(echo "$original" | sed "s|$HOME/||" | tr '/' '__')
    local backup_path="${BACKUP_DIR}/$backup_name"
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # If backup exists and is identical, skip
    if [[ -e "$backup_path" ]]; then
        diff -q "$original" "$backup_path" &>/dev/null && return 0
        backup_name="${backup_name}.${timestamp}"
        backup_path="${BACKUP_DIR}/$backup_name"
    fi

    if [[ -d "$original" ]]; then
        cp -R "$original" "$backup_path"
    else
        cp "$original" "$backup_path"
    fi

    echo "${original}|${backup_name}|${timestamp}|${type}" >> "$BACKUP_MANIFEST"
    echo -e "  ${CYAN}Backed up:${NC} $original"
}

# Sync backups if original files have changed
sync_backups() {
    [[ ! -f "${BACKUP_MANIFEST:-}" ]] && return

    local updated=0
    while IFS='|' read -r original backup_name timestamp type; do
        [[ "$original" =~ ^#.*$ || -z "$original" ]] && continue

        backup_path="${BACKUP_DIR}/$backup_name"
        if [[ -e "$original" && -e "$backup_path" ]]; then
            if ! diff -q "$original" "$backup_path" &>/dev/null 2>&1; then
                if [[ -d "$original" ]]; then
                    rm -rf "$backup_path"
                    cp -R "$original" "$backup_path"
                else
                    cp "$original" "$backup_path"
                fi
                ((updated++))
            fi
        fi
    done < "$BACKUP_MANIFEST"

    if [[ $updated -gt 0 ]]; then
        echo "  Updated $updated backup(s)"
    fi
}

# Restore a backed up file
restore_backup() {
    local original="$1"
    local backup_name
    backup_name=$(echo "$original" | sed "s|$HOME/||" | tr '/' '__')
    local backup_path="${BACKUP_DIR}/$backup_name"

    if [[ -e "$backup_path" ]]; then
        if [[ -d "$backup_path" ]]; then
            cp -R "$backup_path" "$original"
        else
            cp "$backup_path" "$original"
        fi
        return 0
    fi
    return 1
}
