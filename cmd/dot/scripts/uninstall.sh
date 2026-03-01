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
BACKUP_DIR="$DOTFILES_DIR/.backup"
BACKUP_MANIFEST="$BACKUP_DIR/manifest"
SYMLINK_MAP="$DOTFILES_DIR/symlink_map.txt"

# Source libraries
# shellcheck source=lib/colors.sh
source "$SCRIPT_DIR/lib/colors.sh"
# shellcheck source=lib/logging.sh
source "$SCRIPT_DIR/lib/logging.sh"
# shellcheck source=lib/config.sh
source "$SCRIPT_DIR/lib/config.sh"
# shellcheck source=lib/paths.sh
source "$SCRIPT_DIR/lib/paths.sh"
# shellcheck source=lib/backup.sh
source "$SCRIPT_DIR/lib/backup.sh"
# shellcheck source=lib/symlinks.sh
source "$SCRIPT_DIR/lib/symlinks.sh"
# shellcheck source=lib/shell.sh
source "$SCRIPT_DIR/lib/shell.sh"

MACHINE_TYPE="$(read_machine_type)"
MACHINE_TYPE="${MACHINE_TYPE:-home}"

# ============================================================================
# Detection phase - analyze current state
# ============================================================================

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
for pair in "${SHELL_CONFIG_PAIRS[@]}"; do
    local_config="$HOME/.${pair%%:*}"
    local_pattern=".${pair##*:}"
    check_source_line "$local_config" "$local_pattern" && WILL_REMOVE+=("Remove source line from $local_config")
done

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
    for file in "$HOME/.gitconfig.local" "$HOME/.env.local" "$HOME/.shell.local" "$DOTFILES_DIR/.machine"; do
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

# ============================================================================
# Report phase - show what was detected
# ============================================================================

echo "--- Already clean ---"
if [[ ${#ALREADY_CLEAN[@]} -eq 0 ]]; then
    echo "  (nothing)"
else
    for item in "${ALREADY_CLEAN[@]}"; do echo -e "  ${GREEN}${SYM_OK}${NC} $item"; done
fi
echo ""

echo "--- Will remove ---"
if [[ ${#WILL_REMOVE[@]} -eq 0 ]]; then
    echo "  (nothing)"
else
    for item in "${WILL_REMOVE[@]}"; do echo -e "  ${CYAN}${SYM_INFO}${NC} $item"; done
fi
echo ""

echo "--- Available backups to restore ---"
if [[ ${#AVAILABLE_RESTORES[@]} -eq 0 ]]; then
    echo "  (none)"
else
    for item in "${AVAILABLE_RESTORES[@]}"; do
        if [[ "$RESTORE_BACKUPS" == true ]]; then
            echo -e "  ${CYAN}${SYM_BACKUP}${NC} Will restore: $item"
        else
            echo -e "  ${YELLOW}${SYM_BACKUP}${NC} $item"
        fi
    done
    [[ "$RESTORE_BACKUPS" != true ]] && echo "" && echo "  Use --restore to restore these files"
fi
echo ""

echo "--- Frictions ---"
if [[ ${#FRICTIONS[@]} -eq 0 ]]; then
    echo -e "  ${GREEN}None!${NC}"
else
    for item in "${FRICTIONS[@]}"; do echo -e "  ${YELLOW}${SYM_WARN}${NC} $item"; done
fi
echo ""

# ============================================================================
# Execute phase - perform the uninstall
# ============================================================================

if [[ ${#WILL_REMOVE[@]} -eq 0 ]]; then
    echo -e "${GREEN}Nothing to uninstall - already clean!${NC}"
    exit 0
fi

if [[ "$DRY_RUN" == true ]]; then
    echo "Dry run complete. Run without --dry-run to remove."
    exit 0
fi

# Remove symlinks
echo "Removing symlinks..."
for target in "${SYMLINKED_FILES[@]}"; do
    is_ours "$target" && rm -rf "$target" && echo "  Removed: $target"
done

# Remove shell config source lines
echo ""
echo "Removing source lines from shell configs..."
remove_shell_configs

# Remove ~/.dotfiles symlink
[[ -L "$HOME/.dotfiles" ]] && is_ours "$HOME/.dotfiles" && rm "$HOME/.dotfiles" && echo "  Removed: ~/.dotfiles"

# Remove local config files
if [[ "$REMOVE_LOCAL" == true ]]; then
    echo ""
    echo "Removing local config files..."
    for file in "$HOME/.gitconfig.local" "$HOME/.env.local" "$HOME/.shell.local" "$DOTFILES_DIR/.machine"; do
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
