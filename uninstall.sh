#!/bin/bash
# uninstall.sh - Remove dotfiles symlinks and optionally local config
#
# Usage:
#   ./uninstall.sh                 # Remove symlinks only
#   ./uninstall.sh --all           # Also remove local config files
#   ./uninstall.sh --restore       # Restore backed up files after uninstall
#   ./uninstall.sh --dry-run       # Show what would be removed
#
# Exit codes:
#   0 - Success
#   1 - Friction encountered

set -e

REMOVE_LOCAL=false
DRY_RUN=false
RESTORE_BACKUPS=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --all) REMOVE_LOCAL=true; shift ;;
        --dry-run) DRY_RUN=true; shift ;;
        --restore) RESTORE_BACKUPS=true; shift ;;
        --help|-h)
            echo "Usage: ./uninstall.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --all       Also remove local config files (.gitconfig.local, .env.local)"
            echo "  --restore   Restore backed up files after uninstall"
            echo "  --dry-run   Show what would be removed without removing"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="$SCRIPT_DIR"
MACHINE_FILE="$DOTFILES_DIR/.machine"
BACKUP_DIR="$DOTFILES_DIR/.backup"
BACKUP_MANIFEST="$BACKUP_DIR/manifest"
SYMLINK_MAP="$DOTFILES_DIR/symlink_map.txt"

MACHINE_TYPE="home"
[[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")

# Get destination path from symlink map line (with $HOME expanded)
get_dest() {
    local line="$1"
    local dest=$(echo "$line" | cut -d':' -f2-)
    echo "${dest/\$HOME/$HOME}"
}

RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "============================================"
echo "  Dotfiles Uninstall"
echo "============================================"
echo ""
echo "Repo: $DOTFILES_DIR"
echo "Machine: $MACHINE_TYPE"
[[ "$DRY_RUN" == true ]] && echo -e "${CYAN}Mode: DRY RUN${NC}"
[[ "$REMOVE_LOCAL" == true ]] && echo -e "${YELLOW}Will also remove local config files${NC}"
echo ""

echo "Detecting current state..."
echo ""

FRICTIONS=()
WILL_REMOVE=()
ALREADY_CLEAN=()
AVAILABLE_RESTORES=()

is_ours() {
    local path="$1"
    if [[ -L "$path" ]]; then
        local target=$(readlink "$path" 2>/dev/null || echo "")
        [[ "$target" == *"files-with-a-dot"* || "$target" == "$SCRIPT_DIR"* ]]
    elif [[ -e "$path" ]]; then
        local real=$(realpath "$path" 2>/dev/null || echo "")
        [[ "$real" == *"files-with-a-dot"* || "$real" == "$SCRIPT_DIR"* ]]
    else
        return 1
    fi
}

# Check symlinks from symlink_map.txt
SYMLINKED_FILES=()
if [[ -f "$SYMLINK_MAP" ]]; then
    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        dest=$(get_dest "$line")
        SYMLINKED_FILES+=("$dest")
        is_ours "$dest" && WILL_REMOVE+=("Remove symlink: $dest")
    done < "$SYMLINK_MAP"
fi

# Check shell config source lines
check_source_line() {
    local config="$1"
    local pattern="$2"
    [[ -f "$config" ]] && grep -qF "$pattern" "$config" 2>/dev/null && WILL_REMOVE+=("Remove source line from $config")
}

check_source_line "$HOME/.zshrc" ".zshrc.dotfiles"
check_source_line "$HOME/.zprofile" ".zprofile.dotfiles"
check_source_line "$HOME/.bashrc" ".bashrc.dotfiles"
check_source_line "$HOME/.bash_profile" ".bash_profile.dotfiles"

# Check ~/.dotfiles symlink
if [[ -L "$HOME/.dotfiles" ]]; then
    if is_ours "$HOME/.dotfiles"; then
        WILL_REMOVE+=("Remove ~/.dotfiles symlink")
    else
        FRICTIONS+=("~/.dotfiles points to $(readlink "$HOME/.dotfiles"), not this repo")
    fi
elif [[ -e "$HOME/.dotfiles" ]]; then
    FRICTIONS+=("~/.dotfiles exists but is not a symlink")
else
    ALREADY_CLEAN+=("~/.dotfiles symlink")
fi

# Check local config files
if [[ "$REMOVE_LOCAL" == true ]]; then
    for file in "$HOME/.gitconfig.local" "$HOME/.env.local" "$DOTFILES_DIR/.machine"; do
        [[ -f "$file" ]] && WILL_REMOVE+=("Remove local config: $file")
    done
fi

# Check for available backups
if [[ -f "$BACKUP_MANIFEST" ]]; then
    while IFS='|' read -r original backup_name timestamp type; do
        [[ "$original" =~ ^#.*$ || -z "$original" ]] && continue
        [[ -f "$BACKUP_DIR/$backup_name" || -d "$BACKUP_DIR/$backup_name" ]] && AVAILABLE_RESTORES+=("$original (backed up $timestamp)")
    done < "$BACKUP_MANIFEST"
fi

# Report state
echo "--- Already clean ---"
if [[ ${#ALREADY_CLEAN[@]} -eq 0 ]]; then
    echo "  (nothing)"
else
    for item in "${ALREADY_CLEAN[@]}"; do echo -e "  ${GREEN}✓${NC} $item"; done
fi
echo ""

echo "--- Will remove ---"
if [[ ${#WILL_REMOVE[@]} -eq 0 ]]; then
    echo "  (nothing)"
else
    for item in "${WILL_REMOVE[@]}"; do echo -e "  ${CYAN}→${NC} $item"; done
fi
echo ""

echo "--- Available backups to restore ---"
if [[ ${#AVAILABLE_RESTORES[@]} -eq 0 ]]; then
    echo "  (none)"
else
    for item in "${AVAILABLE_RESTORES[@]}"; do
        if [[ "$RESTORE_BACKUPS" == true ]]; then
            echo -e "  ${CYAN}⟳${NC} Will restore: $item"
        else
            echo -e "  ${YELLOW}⟳${NC} $item"
        fi
    done
    [[ "$RESTORE_BACKUPS" != true ]] && echo "" && echo "  Use --restore to restore these files"
fi
echo ""

echo "--- Frictions ---"
if [[ ${#FRICTIONS[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}None!${NC}"
else
    for item in "${FRICTIONS[@]}"; do echo -e "  ${YELLOW}⚠${NC} $item"; done
fi
echo ""

# Gates
if [[ ${#WILL_REMOVE[@]} -eq 0 ]]; then
    echo -e "${GREEN}Nothing to uninstall - already clean!${NC}"
    exit 0
fi

if [[ "$DRY_RUN" == true ]]; then
    echo "Dry run complete. Run without --dry-run to remove."
    exit 0
fi

# Execute uninstall
echo "Removing symlinks..."
for target in "${SYMLINKED_FILES[@]}"; do
    is_ours "$target" && rm -rf "$target" && echo "  Removed: $target"
done

echo ""
echo "Removing source lines from shell configs..."

remove_source_line() {
    local target_file="$1"
    local pattern="$2"

    if [[ -f "$target_file" ]] && grep -qF "$pattern" "$target_file" 2>/dev/null; then
        grep -v "$pattern" "$target_file" | grep -v "# Added by dotfiles" > "$target_file.tmp"
        mv "$target_file.tmp" "$target_file"
        echo "  Cleaned: $target_file"
    fi
}

remove_source_line "$HOME/.zprofile" ".zprofile.dotfiles"
remove_source_line "$HOME/.zshrc" ".zshrc.dotfiles"
remove_source_line "$HOME/.bash_profile" ".bash_profile.dotfiles"
remove_source_line "$HOME/.bashrc" ".bashrc.dotfiles"

[[ -L "$HOME/.dotfiles" ]] && is_ours "$HOME/.dotfiles" && rm "$HOME/.dotfiles" && echo "  Removed: ~/.dotfiles"

if [[ "$REMOVE_LOCAL" == true ]]; then
    echo ""
    echo "Removing local config files..."
    for file in "$HOME/.gitconfig.local" "$HOME/.env.local" "$DOTFILES_DIR/.machine"; do
        [[ -f "$file" ]] && rm "$file" && echo "  Removed: $file"
    done
fi

# Restore backups
if [[ "$RESTORE_BACKUPS" == true && ${#AVAILABLE_RESTORES[@]} -gt 0 ]]; then
    echo ""
    echo "Restoring backed up files..."

    while IFS='|' read -r original backup_name timestamp type; do
        [[ "$original" =~ ^#.*$ || -z "$original" ]] && continue

        backup_path="$BACKUP_DIR/$backup_name"
        if [[ -e "$backup_path" ]]; then
            mkdir -p "$(dirname "$original")"
            if [[ -d "$backup_path" ]]; then
                cp -R "$backup_path" "$original"
            else
                cp "$backup_path" "$original"
            fi
            echo "  Restored: $original"
        fi
    done < "$BACKUP_MANIFEST"
fi

echo ""
echo "============================================"
echo -e "  ${GREEN}Uninstall complete!${NC}"
echo "============================================"
echo ""
[[ ${#AVAILABLE_RESTORES[@]} -gt 0 && "$RESTORE_BACKUPS" != true ]] && echo "Backups available - run with --restore to restore original files" && echo ""
echo "To reinstall, run: $DOTFILES_DIR/sync.sh"
