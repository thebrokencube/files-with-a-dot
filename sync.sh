#!/bin/bash
# sync.sh - Synchronize dotfiles to current machine state
#
# This is the main command for dotfiles management. It intelligently detects
# whether this is a first-time setup or an update, and applies the state from
# symlink_map.txt to your system.
#
# Usage:
#   ./sync.sh                     # Auto-detect mode and sync
#   ./sync.sh --pull              # Force git pull before sync
#   ./sync.sh --dry-run           # Preview changes without applying
#   ./sync.sh --links-only        # Only re-create symlinks
#
# First-time setup (non-interactive):
#   ./sync.sh --machine aggressive --git-name "Name" --git-email "email@example.com"

set -e

# Error handler - show detailed error info (only on actual errors)
trap 'exitcode=$?; if [[ $exitcode -ne 0 ]]; then echo ""; echo "ERROR: Command failed at line $LINENO with exit code $exitcode"; echo "Function: ${FUNCNAME[0]:-main}"; exit $exitcode; fi' ERR

# ============================================================================
# Arguments & Configuration
# ============================================================================

SKIP_BREW=false
SKIP_PULL=false
LINKS_ONLY=false
DRY_RUN=false
NO_BACKUP=false
SKIP_PROMPTS=false
FORCE_PULL=false

MACHINE_ARG=""
PROFILE_ARG=""
GIT_NAME_ARG=""
GIT_EMAIL_ARG=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run) DRY_RUN=true; shift ;;
        --skip-brew) SKIP_BREW=true; shift ;;
        --skip-pull) SKIP_PULL=true; shift ;;
        --pull) FORCE_PULL=true; shift ;;
        --links-only) LINKS_ONLY=true; SKIP_BREW=true; SKIP_PULL=true; shift ;;
        --no-backup) NO_BACKUP=true; shift ;;
        --skip-prompts) SKIP_PROMPTS=true; shift ;;
        --machine) MACHINE_ARG="$2"; shift 2 ;;
        --profile) PROFILE_ARG="$2"; shift 2 ;;
        --git-name) GIT_NAME_ARG="$2"; shift 2 ;;
        --git-email) GIT_EMAIL_ARG="$2"; shift 2 ;;
        --help|-h)
            echo "Usage: ./sync.sh [OPTIONS]"
            echo ""
            echo "Synchronize dotfiles to current machine state."
            echo "Automatically detects first-time setup vs. update mode."
            echo ""
            echo "Options:"
            echo "  --dry-run            Preview changes without applying"
            echo "  --pull               Force git pull before sync"
            echo "  --skip-brew          Skip Homebrew package operations"
            echo "  --skip-pull          Skip git pull"
            echo "  --links-only         Only re-create symlinks (implies --skip-brew --skip-pull)"
            echo "  --no-backup          Skip backing up existing files"
            echo "  --skip-prompts       Skip all interactive prompts"
            echo ""
            echo "First-time setup (non-interactive):"
            echo "  --machine TYPE       Set machine type (aggressive|conservative)"
            echo "  --profile TYPE       Set machine profile (work|personal)"
            echo "  --git-name NAME      Set git user name"
            echo "  --git-email EMAIL    Set git user email"
            echo ""
            echo "Private overlay:"
            echo "  Place private configs in ~/.dotfiles.private/"
            echo "  Run ./init-private.sh to create the structure"
            echo ""
            echo "Debug:"
            echo "  DEBUG=1 ./sync.sh    Run with debug output"
            exit 0
            ;;
        *) echo "Unknown option: $1"; exit 1 ;;
    esac
done

MACHINE_TYPE="${MACHINE_ARG:-${DOTFILES_MACHINE:-}}"
MACHINE_PROFILE="${PROFILE_ARG:-${DOTFILES_PROFILE:-}}"
GIT_NAME="${GIT_NAME_ARG:-${DOTFILES_GIT_NAME:-}}"
GIT_EMAIL="${GIT_EMAIL_ARG:-${DOTFILES_GIT_EMAIL:-}}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_LINK="$HOME/.dotfiles"
DOTFILES_DIR="$SCRIPT_DIR"
BACKUP_DIR="$SCRIPT_DIR/.backup"
BACKUP_MANIFEST="$BACKUP_DIR/manifest"
SYMLINK_MAP="$SCRIPT_DIR/symlink_map.txt"
MACHINE_FILE="$SCRIPT_DIR/.machine"
PROFILE_FILE="$SCRIPT_DIR/.profile"
PRIVATE_DIR="$HOME/.dotfiles.private"

RED='\033[0;31m'
YELLOW='\033[0;33m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m'

# Source libraries (optional - functions defined below if libs don't exist)
[[ -f "$SCRIPT_DIR/lib/paths.sh" ]] && source "$SCRIPT_DIR/lib/paths.sh"

# Private overlay helpers
has_private_overlay() {
    [[ -d "$PRIVATE_DIR" ]]
}

get_profile_dir() {
    echo "$PRIVATE_DIR/$1"
}

# State arrays
ACTIONS=()
ALREADY_DONE=()
FRICTIONS=()
WILL_BACKUP=()

# ============================================================================
# Helper Functions
# ============================================================================

is_ours() {
    local path="$1"
    [[ -e "$path" ]] || return 1
    local real_path=$(realpath "$path" 2>/dev/null || echo "")
    [[ -n "$real_path" && "$real_path" == "$SCRIPT_DIR"* ]]
}

is_foreign_symlink() {
    local path="$1"
    [[ -L "$path" ]] && ! is_ours "$path"
}

get_source() {
    local line="$1"
    echo "$line" | cut -d':' -f1
}

get_dest() {
    local line="$1"
    local dest=$(echo "$line" | cut -d':' -f2-)
    echo "${dest/\$HOME/$HOME}"
}

# ============================================================================
# Backup System
# ============================================================================

init_backup() {
    [[ "$NO_BACKUP" == true ]] && return
    mkdir -p "$BACKUP_DIR"
    if [[ ! -f "$BACKUP_MANIFEST" ]]; then
        echo "# Dotfiles backup manifest" > "$BACKUP_MANIFEST"
        echo "# Format: original_path|backup_name|timestamp|type" >> "$BACKUP_MANIFEST"
    fi
}

backup_file() {
    local original="$1"
    local type="${2:-unknown}"

    [[ "$NO_BACKUP" == true ]] && return 0
    [[ ! -e "$original" ]] && return 0

    if [[ -L "$original" ]]; then
        local target=$(realpath "$original" 2>/dev/null || echo "")
        [[ "$target" == "$SCRIPT_DIR"* ]] && return 0
    fi

    local backup_name=$(echo "$original" | sed "s|$HOME/||" | tr '/' '__')
    local backup_path="$BACKUP_DIR/$backup_name"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    if [[ -e "$backup_path" ]]; then
        diff -q "$original" "$backup_path" &>/dev/null && return 0
        backup_name="${backup_name}.${timestamp}"
        backup_path="$BACKUP_DIR/$backup_name"
    fi

    if [[ -d "$original" ]]; then
        cp -R "$original" "$backup_path"
    else
        cp "$original" "$backup_path"
    fi

    echo "${original}|${backup_name}|${timestamp}|${type}" >> "$BACKUP_MANIFEST"
    echo -e "  ${CYAN}Backed up:${NC} $original"
}

sync_backups() {
    [[ ! -f "$BACKUP_MANIFEST" ]] && return

    local updated=0
    while IFS='|' read -r original backup_name timestamp type; do
        [[ "$original" =~ ^#.*$ || -z "$original" ]] && continue

        backup_path="$BACKUP_DIR/$backup_name"
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

# ============================================================================
# State Detection
# ============================================================================

detect_first_time() {
    [[ ! -f "$MACHINE_FILE" ]]
}

# ============================================================================
# Analysis & Planning
# ============================================================================

analyze_state() {
    local is_first_time="$1"

    cd "$SCRIPT_DIR" || {
        echo "Error: Cannot access dotfiles directory: $SCRIPT_DIR"
        exit 1
    }

    # Check git status
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking git status..."
    if [[ "$SKIP_PULL" == false ]]; then
        if ! git rev-parse --is-inside-work-tree &>/dev/null; then
            FRICTIONS+=("Not a git repository")
        elif ! git diff --quiet 2>/dev/null || ! git diff --cached --quiet 2>/dev/null; then
            FRICTIONS+=("Uncommitted changes in repo - commit or stash before pulling")
        else
            git fetch --quiet 2>/dev/null || true
            LOCAL=$(git rev-parse HEAD 2>/dev/null || echo "")
            REMOTE=$(git rev-parse @{u} 2>/dev/null || echo "")
            if [[ -n "$LOCAL" && -n "$REMOTE" && "$LOCAL" != "$REMOTE" ]]; then
                BEHIND=$(git rev-list --count HEAD..@{u} 2>/dev/null || echo "0")
                if [[ "$BEHIND" -gt 0 ]]; then
                    ACTIONS+=("Pull $BEHIND commit(s) from remote")
                else
                    ALREADY_DONE+=("Git repo up to date")
                fi
            else
                ALREADY_DONE+=("Git repo up to date")
            fi
        fi
    fi

    # Check ~/.dotfiles
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking ~/.dotfiles..."
    if [[ "$SCRIPT_DIR" == "$DOTFILES_LINK" ]]; then
        ALREADY_DONE+=("~/.dotfiles (repo location)")
    elif [[ -L "$DOTFILES_LINK" ]]; then
        if is_ours "$DOTFILES_LINK"; then
            ALREADY_DONE+=("~/.dotfiles symlink")
        else
            FRICTIONS+=("~/.dotfiles is a symlink to $(readlink "$DOTFILES_LINK"), not this repo")
        fi
    elif [[ -d "$DOTFILES_LINK" ]]; then
        FRICTIONS+=("~/.dotfiles is a directory (not this repo or symlink to it) - remove or rename it")
    elif [[ -e "$DOTFILES_LINK" ]]; then
        FRICTIONS+=("~/.dotfiles exists but is not a symlink or directory")
    else
        ACTIONS+=("Create ~/.dotfiles symlink")
    fi

    # Check required tools
    command -v git &>/dev/null || FRICTIONS+=("Required tool not found: git")
    command -v brew &>/dev/null || FRICTIONS+=("Homebrew not installed - run bootstrap.sh first")

    # Check symlink map exists
    [[ ! -f "$SYMLINK_MAP" ]] && FRICTIONS+=("symlink_map.txt not found")

    # Check shell configs
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking shell configs..."
    check_shell_config "$HOME/.zshrc" ".zshrc.dotfiles"
    check_shell_config "$HOME/.zprofile" ".zprofile.dotfiles"
    check_shell_config "$HOME/.bashrc" ".bashrc.dotfiles"
    check_shell_config "$HOME/.bash_profile" ".bash_profile.dotfiles"

    # Check all symlink map entries
    [[ "${DEBUG:-}" == "1" ]] && echo "  Checking symlink map entries..."
    if [[ -f "$SYMLINK_MAP" ]]; then
        while IFS= read -r line || [[ -n "$line" ]]; do
            [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
            local source=$(get_source "$line")
            local dest=$(get_dest "$line")
            check_symlink "$source" "$dest" || true
        done < "$SYMLINK_MAP"
    fi

    # Check brew
    if [[ "$SKIP_BREW" == false ]]; then
        if ! command -v brew &>/dev/null; then
            FRICTIONS+=("Homebrew not installed")
        else
            ACTIONS+=("Install/update Homebrew packages")
        fi
    fi

    # Check backups
    [[ -f "$BACKUP_MANIFEST" ]] && ACTIONS+=("Sync backups (update if files changed)")

    # Ensure function always returns success
    return 0
}

check_shell_config() {
    local config="$1"
    local source_file="$2"

    if [[ ! -e "$config" ]]; then
        ACTIONS+=("Create $config with source line")
    elif is_foreign_symlink "$config"; then
        FRICTIONS+=("$config is a symlink to $(readlink "$config") - can't modify")
    elif grep -qF "$source_file" "$config" 2>/dev/null; then
        ALREADY_DONE+=("$config sources $source_file")
    else
        ACTIONS+=("Append source line to $config")
        [[ "$NO_BACKUP" != true ]] && WILL_BACKUP+=("$config (shell config)")
    fi
    return 0
}

check_symlink() {
    local source="$1"
    local dest="$2"
    local name=$(basename "$source")
    local source_path="$SCRIPT_DIR/$source"

    if [[ ! -e "$source_path" ]]; then
        return 0
    fi

    if [[ -L "$dest" ]]; then
        if [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path" 2>/dev/null)" ]]; then
            ALREADY_DONE+=("$name")
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

# ============================================================================
# Reporting
# ============================================================================

report_state() {
    echo "--- Already in place ---"
    if [[ ${#ALREADY_DONE[@]} -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${ALREADY_DONE[@]}"; do echo -e "  ${GREEN}✓${NC} $item"; done
    fi
    echo ""

    echo "--- Will do ---"
    if [[ ${#ACTIONS[@]} -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${ACTIONS[@]}"; do echo -e "  ${CYAN}→${NC} $item"; done
    fi
    echo ""

    echo "--- Will backup ---"
    if [[ ${#WILL_BACKUP[@]} -eq 0 ]]; then
        echo "  (nothing)"
    else
        for item in "${WILL_BACKUP[@]}"; do echo -e "  ${YELLOW}⟳${NC} $item"; done
    fi
    echo ""

    echo "--- Frictions (cannot proceed) ---"
    if [[ ${#FRICTIONS[@]} -eq 0 ]]; then
        echo -e "  ${GREEN}None!${NC}"
    else
        for item in "${FRICTIONS[@]}"; do echo -e "  ${RED}✗${NC} $item"; done
    fi
    echo ""
}

# ============================================================================
# Git Operations
# ============================================================================

handle_git_pull() {
    local is_first_time="$1"

    [[ "$SKIP_PULL" == true ]] && return 0
    [[ "$is_first_time" == true ]] && return 0

    if [[ "$FORCE_PULL" == true ]]; then
        echo "Pulling latest dotfiles..."
        git pull --rebase
        echo ""
        return 0
    fi

    # Auto-pull only if behind
    local behind=$(git rev-list --count HEAD..@{u} 2>/dev/null || echo "0")
    if [[ "$behind" -gt 0 ]]; then
        echo "Pulling latest dotfiles..."
        git pull --rebase
        echo ""
    fi
}

# ============================================================================
# Symlink Operations
# ============================================================================

create_dotfiles_symlink() {
    if [[ "$SCRIPT_DIR" == "$DOTFILES_LINK" ]]; then
        return 0
    elif [[ ! -L "$DOTFILES_LINK" && ! -e "$DOTFILES_LINK" ]]; then
        echo "Creating ~/.dotfiles symlink..."
        ln -s "$SCRIPT_DIR" "$DOTFILES_LINK"
        DOTFILES_DIR="$DOTFILES_LINK"
    else
        DOTFILES_DIR="$DOTFILES_LINK"
    fi
}

apply_symlinks() {
    echo "Creating symlinks..."

    # Public symlinks
    while IFS= read -r line || [[ -n "$line" ]]; do
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
        local source=$(get_source "$line")
        local dest=$(get_dest "$line")
        create_symlink "$source" "$dest"
    done < "$SYMLINK_MAP"

    # Private overlay symlinks
    if has_private_overlay; then
        echo ""
        echo "Applying private overlay..."

        # Shared private symlinks
        local private_map="$PRIVATE_DIR/symlink_map.txt"
        if [[ -f "$private_map" ]]; then
            while IFS= read -r line || [[ -n "$line" ]]; do
                [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
                local source=$(get_source "$line")
                local dest=$(get_dest "$line")
                create_private_symlink "$source" "$dest" "$PRIVATE_DIR"
            done < "$private_map"
        fi

        # Profile-specific symlinks
        local profile_dir="$(get_profile_dir "$MACHINE_PROFILE")"
        local profile_map="$profile_dir/symlink_map.txt"
        if [[ -f "$profile_map" ]]; then
            echo "  Applying $MACHINE_PROFILE profile..."
            while IFS= read -r line || [[ -n "$line" ]]; do
                [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
                local source=$(get_source "$line")
                local dest=$(get_dest "$line")
                create_private_symlink "$source" "$dest" "$profile_dir"
            done < "$profile_map"
        fi

        # Profile-specific skills
        local skills_dir="$profile_dir/skills"
        if [[ -d "$skills_dir" ]]; then
            for skill in "$skills_dir"/*/; do
                [[ -d "$skill" ]] || continue
                local skill_name=$(basename "$skill")
                [[ "$skill_name" == ".gitkeep" ]] && continue
                local dest="$HOME/.claude/skills/$skill_name"
                mkdir -p "$(dirname "$dest")"
                create_private_symlink "skills/$skill_name" "$dest" "$profile_dir"
            done
        fi
    fi

    echo ""
}

create_symlink() {
    local source="$1"
    local dest="$2"
    local source_path="$DOTFILES_DIR/$source"
    local name=$(basename "$source")

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
    local parent_dir=$(dirname "$dest")
    if [[ ! -d "$parent_dir" ]]; then
        mkdir -p "$parent_dir"
    fi

    # Create symlink
    ln -s "$source_path" "$dest"
    echo "  $name -> $dest"
}

create_private_symlink() {
    local source="$1"
    local dest="$2"
    local base_dir="$3"
    local source_path="$base_dir/$source"
    local name=$(basename "$source")

    # Skip if source doesn't exist
    if [[ ! -e "$source_path" ]]; then
        return
    fi

    # Skip if already correctly linked
    if [[ -L "$dest" ]] && [[ "$(realpath "$dest" 2>/dev/null)" == "$(realpath "$source_path")" ]]; then
        return
    fi

    # Backup and remove existing (can override public symlinks)
    if [[ -e "$dest" || -L "$dest" ]]; then
        # Only backup if it's not one of our symlinks
        if ! is_ours "$dest"; then
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

# ============================================================================
# Shell Integration
# ============================================================================

integrate_shell_configs() {
    echo "Integrating shell configs..."

    add_source_line "$HOME/.zprofile" "$HOME/.zprofile.dotfiles"
    add_source_line "$HOME/.zshrc" "$HOME/.zshrc.dotfiles"
    add_source_line "$HOME/.bash_profile" "$HOME/.bash_profile.dotfiles"
    add_source_line "$HOME/.bashrc" "$HOME/.bashrc.dotfiles"
    echo ""
}

add_source_line() {
    local target="$1"
    local source_file="$2"
    local source_line="[ -f \"$source_file\" ] && source \"$source_file\""

    [[ ! -f "$target" ]] && echo "# Shell configuration" > "$target"

    if ! grep -qF "$source_file" "$target" 2>/dev/null; then
        backup_file "$target" "shell_config"
        echo "" >> "$target"
        echo "# Added by dotfiles" >> "$target"
        echo "$source_line" >> "$target"
        echo "  Added to $target"
    fi
    return 0
}

# ============================================================================
# Homebrew Management
# ============================================================================

install_brew_packages() {
    local is_first_time="$1"

    [[ "$SKIP_BREW" == true ]] && return 0

    echo "Installing Homebrew packages..."
    brew bundle --file="$DOTFILES_DIR/Brewfile.shared" --verbose || echo "  (some packages may have failed)"

    if [[ "$MACHINE_TYPE" == "aggressive" && -f "$DOTFILES_DIR/aggressive/Brewfile" ]]; then
        echo ""
        echo "Installing aggressive-mode packages..."
        brew bundle --file="$DOTFILES_DIR/aggressive/Brewfile" --verbose || echo "  (some packages may have failed)"
    fi

    # Private overlay Brewfiles
    if has_private_overlay; then
        # Shared private Brewfile
        local private_brewfile="$PRIVATE_DIR/Brewfile"
        if [[ -f "$private_brewfile" ]]; then
            echo ""
            echo "Installing private packages..."
            brew bundle --file="$private_brewfile" --verbose || echo "  (some packages may have failed)"
        fi

        # Profile-specific Brewfile
        local profile_brewfile="$(get_profile_dir "$MACHINE_PROFILE")/Brewfile"
        if [[ -f "$profile_brewfile" ]]; then
            echo ""
            echo "Installing $MACHINE_PROFILE profile packages..."
            brew bundle --file="$profile_brewfile" --verbose || echo "  (some packages may have failed)"
        fi
    fi

    # Update mode: upgrade existing packages
    if [[ "$is_first_time" == false ]]; then
        echo ""
        echo "Upgrading installed packages..."
        brew upgrade || true
    fi

    if command -v mise &>/dev/null; then
        echo ""
        echo "Updating mise tools..."
        mise upgrade || echo "  (no updates available)"
    fi

    if [[ "$MACHINE_TYPE" == "aggressive" ]]; then
        echo ""
        echo "Running brew cleanup..."
        brew cleanup || echo "  (no cleanup needed)"
    fi
    echo ""
}

# ============================================================================
# Local Config Setup
# ============================================================================

setup_local_configs() {
    local is_first_time="$1"
    local local_dir="$DOTFILES_DIR/local"

    echo "Setting up local config..."

    # Create local directory if needed
    mkdir -p "$local_dir"

    # Migrate existing files from home directory to local/
    if [[ -f "$HOME/.gitconfig.local" && ! -f "$local_dir/gitconfig.local" ]]; then
        mv "$HOME/.gitconfig.local" "$local_dir/gitconfig.local"
        echo "  Migrated ~/.gitconfig.local to local/"
    fi
    if [[ -f "$HOME/.env.local" && ! -f "$local_dir/env.local" ]]; then
        mv "$HOME/.env.local" "$local_dir/env.local"
        echo "  Migrated ~/.env.local to local/"
    fi

    # Git config
    if [[ ! -f "$local_dir/gitconfig.local" ]]; then
        if [[ -n "$GIT_NAME" && -n "$GIT_EMAIL" ]]; then
            cat > "$local_dir/gitconfig.local" << EOF
[user]
    name = $GIT_NAME
    email = $GIT_EMAIL
EOF
            echo "  Created local/gitconfig.local"
        elif [[ "$is_first_time" == true && "$SKIP_PROMPTS" != true ]]; then
            echo ""
            read -p "Git name: " GIT_NAME
            read -p "Git email: " GIT_EMAIL
            if [[ -n "$GIT_NAME" && -n "$GIT_EMAIL" ]]; then
                cat > "$local_dir/gitconfig.local" << EOF
[user]
    name = $GIT_NAME
    email = $GIT_EMAIL
EOF
                echo "  Created local/gitconfig.local"
            fi
        else
            echo "  Skipping local/gitconfig.local (no values provided)"
        fi
    else
        echo "  local/gitconfig.local exists"
    fi

    # Env config
    if [[ ! -f "$local_dir/env.local" ]]; then
        cat > "$local_dir/env.local" << 'EOF'
# Local environment variables (not in git)
# Add your environment variables here
# Example: GITHUB_TOKEN=ghp_...
# Example: ANTHROPIC_API_KEY=sk-ant-...
EOF
        echo "  Created local/env.local"
    else
        echo "  local/env.local exists"
    fi

    # Shell local config
    # shell.managed - aggressive mode only, comes from repo (tracked in git)
    if [[ "$MACHINE_TYPE" == "aggressive" ]]; then
        if [[ -f "$local_dir/shell.managed" ]]; then
            echo "  local/shell.managed exists (repo-managed)"
        else
            echo "  Warning: local/shell.managed not found - run git pull"
        fi
    fi

    # shell.local - never overwritten, user's private customizations
    if [[ ! -f "$local_dir/shell.local" ]]; then
        cat > "$local_dir/shell.local" << 'EOF'
# ============================================================================
# Private local shell config (never touched by sync)
# ============================================================================
# Add your private aliases, functions, and machine-specific config here.
# This file is never modified by dotfiles sync.
#
# Examples:
# alias myalias='my command'
# export MY_VAR="value"
EOF
        echo "  Created local/shell.local"
    else
        echo "  local/shell.local exists (not modified)"
    fi

    echo ""
}

# ============================================================================
# First-Time Setup
# ============================================================================

prompt_first_time_config() {
    if [[ -n "$MACHINE_TYPE" ]]; then
        if [[ "$MACHINE_TYPE" != "aggressive" && "$MACHINE_TYPE" != "conservative" ]]; then
            echo "Error: Invalid machine type '$MACHINE_TYPE'. Must be 'aggressive' or 'conservative'."
            exit 1
        fi
        echo "$MACHINE_TYPE" > "$MACHINE_FILE"
    elif [[ "$SKIP_PROMPTS" == true ]]; then
        MACHINE_TYPE="aggressive"
        echo "$MACHINE_TYPE" > "$MACHINE_FILE"
        echo "Machine type: $MACHINE_TYPE (default)"
    else
        echo ""
        echo "Machine type:"
        echo "  1) aggressive - Source of truth, aggressive cleanup (personal machines)"
        echo "  2) conservative - Minimal changes, show cleanup opportunities only (work machines)"
        read -p "Enter choice [1-2]: " choice
        case "$choice" in
            1) MACHINE_TYPE="aggressive" ;;
            2) MACHINE_TYPE="conservative" ;;
            *) MACHINE_TYPE="aggressive" ;;
        esac
        echo "$MACHINE_TYPE" > "$MACHINE_FILE"
    fi

    # Profile selection
    if [[ -n "$MACHINE_PROFILE" ]]; then
        if [[ "$MACHINE_PROFILE" != "work" && "$MACHINE_PROFILE" != "personal" ]]; then
            echo "Error: Invalid profile '$MACHINE_PROFILE'. Must be 'work' or 'personal'."
            exit 1
        fi
        echo "$MACHINE_PROFILE" > "$PROFILE_FILE"
    elif [[ "$SKIP_PROMPTS" == true ]]; then
        MACHINE_PROFILE="personal"
        echo "$MACHINE_PROFILE" > "$PROFILE_FILE"
        echo "Profile: $MACHINE_PROFILE (default)"
    else
        echo ""
        echo "Machine profile:"
        echo "  1) personal - Personal machine"
        echo "  2) work - Work machine"
        read -p "Enter choice [1-2]: " profile_choice
        case "$profile_choice" in
            1) MACHINE_PROFILE="personal" ;;
            2) MACHINE_PROFILE="work" ;;
            *) MACHINE_PROFILE="personal" ;;
        esac
        echo "$MACHINE_PROFILE" > "$PROFILE_FILE"
    fi
}

# ============================================================================
# Main Execution
# ============================================================================

VERSION=$(cd "$DOTFILES_DIR" && git describe --tags --always 2>/dev/null || echo "unknown")
echo "============================================"
echo "  Dotfiles Sync ($VERSION)"
echo "============================================"
echo ""
echo "Repo: $DOTFILES_DIR"
[[ "$DRY_RUN" == true ]] && echo -e "${CYAN}Mode: DRY RUN${NC}"
echo ""

# Detect mode
IS_FIRST_TIME=$(detect_first_time && echo true || echo false)

if [[ "$IS_FIRST_TIME" == true ]]; then
    echo "Mode: First-time setup"
    [[ -f "$MACHINE_FILE" ]] && MACHINE_TYPE=$(cat "$MACHINE_FILE")
    [[ -f "$PROFILE_FILE" ]] && MACHINE_PROFILE=$(cat "$PROFILE_FILE")
else
    echo "Mode: Update"
    MACHINE_TYPE=$(cat "$MACHINE_FILE")
    # Use --profile arg if provided, otherwise read from file
    if [[ -n "$PROFILE_ARG" ]]; then
        MACHINE_PROFILE="$PROFILE_ARG"
        # Save the new profile
        echo "$MACHINE_PROFILE" > "$PROFILE_FILE"
    elif [[ -f "$PROFILE_FILE" ]]; then
        MACHINE_PROFILE=$(cat "$PROFILE_FILE")
    else
        MACHINE_PROFILE="personal"
    fi
fi
echo "Machine type: ${MACHINE_TYPE:-unknown}"
echo "Profile: ${MACHINE_PROFILE:-personal}"
if has_private_overlay; then
    echo "Private: $PRIVATE_DIR"
fi
echo ""

echo "Detecting current state..."
echo ""

# Analyze state
analyze_state "$IS_FIRST_TIME"

# Report
report_state

# Gates
if [[ ${#FRICTIONS[@]} -gt 0 ]]; then
    echo -e "${RED}Cannot proceed due to frictions above.${NC}"
    exit 1
fi

if [[ ${#ACTIONS[@]} -eq 0 ]]; then
    echo -e "${GREEN}Nothing to do!${NC}"
    exit 0
fi

if [[ "$DRY_RUN" == true ]]; then
    echo "Dry run complete. Run without --dry-run to apply changes."
    exit 0
fi

if [[ "$SKIP_PROMPTS" != true ]]; then
    read -p "Proceed with sync? [Y/n] " confirm
    [[ "$confirm" == "n" || "$confirm" == "N" ]] && echo "Aborted." && exit 2
fi

# Execute sync
echo ""
echo "============================================"
echo "  Executing Sync"
echo "============================================"
echo ""

# First-time setup
if [[ "$IS_FIRST_TIME" == true ]]; then
    prompt_first_time_config
    echo ""
fi

# Initialize backup system
init_backup

# Git pull
handle_git_pull "$IS_FIRST_TIME"

# Sync backups
if [[ -f "$BACKUP_MANIFEST" ]]; then
    echo "Syncing backups..."
    sync_backups
    echo ""
fi

# Create ~/.dotfiles symlink if needed
create_dotfiles_symlink

# Migrate from ~/.claude directory symlink to granular linking
if [[ -L "$HOME/.claude" && -d "$HOME/.claude" ]]; then
    target=$(readlink "$HOME/.claude")
    if [[ "$target" == *"shared/claude/.claude"* || "$target" == *".dotfiles"*"/.claude"* ]]; then
        echo "Migrating ~/.claude from directory link to granular links..."
        rm "$HOME/.claude"  # Remove symlink (not the directory it points to)
        mkdir -p "$HOME/.claude/skills"
        echo ""
    fi
fi

# Apply symlinks
apply_symlinks

# Integrate shell configs
integrate_shell_configs

# Install/update brew packages
install_brew_packages "$IS_FIRST_TIME"

# Setup local configs
setup_local_configs "$IS_FIRST_TIME"

# Setup mise
if command -v mise &>/dev/null; then
    echo "Setting up mise..."
    mise trust "$DOTFILES_DIR" 2>/dev/null || true
    mise use --global node@lts 2>/dev/null || true
    echo ""
fi

echo "============================================"
echo -e "  ${GREEN}Sync complete!${NC}"
echo "============================================"
echo ""

# Prompt to run cleanup
if [[ "$SKIP_PROMPTS" != true ]]; then
    read -p "Run cleanup? [y/N] " cleanup_response
    if [[ "$cleanup_response" == "y" || "$cleanup_response" == "Y" ]]; then
        echo ""
        "$DOTFILES_DIR/cleanup.sh" --force
    fi
fi

echo ""
echo "Next steps:"
echo "  - Open nvim and run :Lazy update for plugins"
if [[ "$IS_FIRST_TIME" == true ]]; then
    echo "  - Run ./health.sh to verify"
fi
echo ""

# Prompt to reload shell
if [[ "$SKIP_PROMPTS" != true ]]; then
    read -p "Reload shell now? [Y/n] " reload_response
    if [[ "$reload_response" != "n" && "$reload_response" != "N" ]]; then
        exec $SHELL -l
    fi
fi
