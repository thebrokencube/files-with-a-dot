#!/bin/bash
# lib/private.sh - Private overlay support
#
# Usage: source "$(dirname "$0")/lib/private.sh"
# Requires: DOTFILES_DIR to be set
#
# Management functions (init, status, pull, push, sync) additionally require:
#   lib/colors.sh, lib/logging.sh, lib/config.sh, lib/prompt.sh to be sourced
#   lib/git.sh for push and status operations
#   lib/paths.sh, lib/backup.sh for sync operations

PRIVATE_DIR="$HOME/.dotfiles.private"

# Check if private overlay exists
has_private_overlay() {
    [[ -d "$PRIVATE_DIR" ]]
}

# Check if private overlay is a git repo
has_private_git() {
    [[ -d "$PRIVATE_DIR/.git" ]]
}

# Get private symlink map path
get_private_symlink_map() {
    echo "$PRIVATE_DIR/symlink_map.txt"
}

# Get private Brewfile path
get_private_brewfile() {
    echo "$PRIVATE_DIR/Brewfile"
}

# Check private symlink (for analysis phase)
check_private_symlink() {
    local source="$1"
    local dest="$2"
    local base_dir="${3:-$PRIVATE_DIR}"
    local name
    name=$(basename "$source")
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
    local name
    name=$(basename "$source")
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
    local parent_dir
    parent_dir=$(dirname "$dest")
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
        local source
        source=$(get_source "$line")
        local dest
        dest=$(get_dest "$line")
        create_private_symlink "$source" "$dest" "$base_dir"
    done < "$symlink_map"
}

# ── Migration ─────────────────────────────────────────────────────────────────

# Detect and migrate legacy structures into the current private overlay format.
# Handles: profile subdirs (work/, personal/), local/ config files, .profile file.
# Safe to call on every sync — returns immediately if nothing to migrate.
migrate_private_overlay() {
    has_private_overlay || return 0

    local did_migrate=false

    # ── Legacy profile dirs (work/, personal/) ────────────────────────────
    if [[ -d "$PRIVATE_DIR/work" || -d "$PRIVATE_DIR/personal" ]]; then
        echo "Migrating private overlay from legacy profile structure..."
        did_migrate=true

        # Find which profile dir has content (only one will in practice)
        local active_profile=""
        for profile in work personal; do
            local pdir="$PRIVATE_DIR/$profile"
            [[ ! -d "$pdir" ]] && continue
            local real_files
            real_files=$(find "$pdir" -not -name '.gitkeep' -not -path "$pdir" -not -type d 2>/dev/null | head -1)
            if [[ -n "$real_files" ]]; then
                active_profile="$profile"
                break
            fi
        done

        if [[ -n "$active_profile" ]]; then
            local pdir="$PRIVATE_DIR/$active_profile"
            echo "  Active profile: $active_profile"

            # Merge symlink_map.txt
            if [[ -f "$pdir/symlink_map.txt" ]]; then
                local has_entries
                has_entries=$(grep -cv '^\s*#\|^\s*$' "$pdir/symlink_map.txt" 2>/dev/null || echo 0)
                if [[ "$has_entries" -gt 0 ]]; then
                    echo "" >> "$PRIVATE_DIR/symlink_map.txt"
                    echo "# Migrated from $active_profile/" >> "$PRIVATE_DIR/symlink_map.txt"
                    grep -v '^\s*#\|^\s*$' "$pdir/symlink_map.txt" >> "$PRIVATE_DIR/symlink_map.txt"
                    echo "  Merged $has_entries symlink(s) from $active_profile/symlink_map.txt"
                fi
            fi

            # Merge Brewfile
            if [[ -f "$pdir/Brewfile" ]]; then
                local has_entries
                has_entries=$(grep -cv '^\s*#\|^\s*$' "$pdir/Brewfile" 2>/dev/null || echo 0)
                if [[ "$has_entries" -gt 0 ]]; then
                    echo "" >> "$PRIVATE_DIR/Brewfile"
                    echo "# Migrated from $active_profile/" >> "$PRIVATE_DIR/Brewfile"
                    grep -v '^\s*#\|^\s*$' "$pdir/Brewfile" >> "$PRIVATE_DIR/Brewfile"
                    echo "  Merged $has_entries package(s) from $active_profile/Brewfile"
                fi
            fi

            # Move skills
            if [[ -d "$pdir/skills" ]]; then
                mkdir -p "$PRIVATE_DIR/skills"
                for skill in "$pdir/skills"/*/; do
                    [[ -d "$skill" ]] || continue
                    local skill_name
                    skill_name=$(basename "$skill")
                    [[ "$skill_name" == ".gitkeep" ]] && continue
                    if [[ ! -d "$PRIVATE_DIR/skills/$skill_name" ]]; then
                        mv "$skill" "$PRIVATE_DIR/skills/$skill_name"
                        echo "  Moved skill: $skill_name"
                    fi
                done
            fi
        fi

        rm -rf "$PRIVATE_DIR/work" "$PRIVATE_DIR/personal"
        [[ -d "$PRIVATE_DIR/shared" ]] && rmdir "$PRIVATE_DIR/shared" 2>/dev/null || true
        echo "  Removed legacy profile directories"
    fi

    # ── Legacy local/ config files → adopt into private overlay ───────────
    local local_dir="$DOTFILES_DIR/local"
    if [[ -d "$local_dir" ]]; then
        for pair in "gitconfig.local:$HOME/.gitconfig.local" "env.local:$HOME/.env.local" "shell.local:$HOME/.shell.local"; do
            local name="${pair%%:*}"
            local home_path="${pair##*:}"
            if [[ -f "$local_dir/$name" && ! -f "$PRIVATE_DIR/$name" ]]; then
                mv "$local_dir/$name" "$PRIVATE_DIR/$name"
                echo "  Adopted $name from local/ into private overlay"
                did_migrate=true
            fi
        done
        # Clean up local/ (shell.managed moved to aggressive/)
        rm -f "$local_dir/shell.managed" 2>/dev/null
        rmdir "$local_dir" 2>/dev/null || true
    fi

    # ── Ensure symlink_map.txt has correct entries for local config files ─
    local private_map="$PRIVATE_DIR/symlink_map.txt"
    if [[ -f "$private_map" ]]; then
        for pair in "gitconfig.local:\$HOME/.gitconfig.local" "env.local:\$HOME/.env.local" "shell.local:\$HOME/.shell.local"; do
            local name="${pair%%:*}"
            local dest="${pair##*:}"
            local correct_entry="${name}:${dest}"
            if grep -q "^${name}:" "$private_map" 2>/dev/null; then
                # Entry exists — fix destination if it points somewhere stale
                local current_entry
                current_entry=$(grep "^${name}:" "$private_map" | head -1)
                if [[ "$current_entry" != "$correct_entry" ]]; then
                    # Replace stale entry with correct one
                    local escaped_old
                    escaped_old=$(printf '%s\n' "$current_entry" | sed 's/[[\.*^$()+?{|]/\\&/g')
                    sed -i '' "s|${escaped_old}|${correct_entry}|" "$private_map"
                    echo "  Fixed $name destination in private symlink_map.txt"
                    did_migrate=true
                fi
            else
                echo "$correct_entry" >> "$private_map"
                echo "  Added $name to private symlink_map.txt"
                did_migrate=true
            fi
        done
    fi

    # ── Legacy .profile file ──────────────────────────────────────────────
    if [[ -f "$DOTFILES_DIR/.profile" ]]; then
        rm -f "$DOTFILES_DIR/.profile"
        echo "  Removed legacy .profile file"
        did_migrate=true
    fi

    [[ "$did_migrate" == true ]] && echo "" || true
}

# ── Management functions ─────────────────────────────────────────────────────
# These require colors.sh, logging.sh, config.sh, prompt.sh to be sourced.

# Initialize a private dotfiles overlay from the template.
# Args: $1 = --no-git (optional, skip git init)
# Requires: confirm() from lib/prompt.sh
init_private_overlay() {
    local init_git=true
    [[ "${1:-}" == "--no-git" ]] && init_git=false

    echo "============================================"
    echo "  Initialize Private Dotfiles Overlay"
    echo "============================================"
    echo ""
    echo "Location: $PRIVATE_DIR"
    echo ""

    # Check if already exists
    if [[ -d "$PRIVATE_DIR" ]]; then
        warn "$PRIVATE_DIR already exists"
        if ! confirm "Overwrite? This will delete existing content." "no"; then
            echo "Aborted."
            return 1
        fi
        rm -rf "$PRIVATE_DIR"
    fi

    # Copy template
    info "Creating directory structure..."
    cp -R "$DOTFILES_DIR/configs/templates/private" "$PRIVATE_DIR"

    # Create .gitkeep for empty directories
    touch "$PRIVATE_DIR/skills/.gitkeep"

    # Create empty files if not from template
    [[ ! -f "$PRIVATE_DIR/symlink_map.txt" ]] && echo "# Private symlinks" > "$PRIVATE_DIR/symlink_map.txt"
    [[ ! -f "$PRIVATE_DIR/Brewfile" ]] && echo "# Private packages" > "$PRIVATE_DIR/Brewfile"

    # Adopt existing local config files into the overlay
    local adopted=false
    local local_dir="$DOTFILES_DIR/local"
    # Check both local/ and $HOME for each file
    for pair in "gitconfig.local:$HOME/.gitconfig.local" "env.local:$HOME/.env.local" "shell.local:$HOME/.shell.local"; do
        local name="${pair%%:*}"
        local home_path="${pair##*:}"
        local source=""
        # Prefer local/ (legacy) over $HOME
        if [[ -f "$local_dir/$name" ]]; then
            source="$local_dir/$name"
        elif [[ -f "$home_path" && ! -L "$home_path" ]]; then
            source="$home_path"
        fi
        if [[ -n "$source" && ! -f "$PRIVATE_DIR/$name" ]]; then
            cp "$source" "$PRIVATE_DIR/$name"
            echo "  Adopted $name from $source"
            adopted=true
        fi
    done
    if [[ "$adopted" == true ]]; then
        echo ""
    fi

    # Initialize git repo
    if [[ "$init_git" == true ]]; then
        echo ""
        info "Initializing git repository..."
        (
            cd "$PRIVATE_DIR" || return
            git init

            cat > .gitignore << 'GITIGNORE'
# OS files
.DS_Store
*.swp
*~

# Secrets (if any accidentally added)
*.secret
*.key
GITIGNORE

            git add -A
            git commit -m "Initial private dotfiles overlay"
        )

        echo ""
        ok "Git repository initialized."
        echo ""
        echo "To push to a remote:"
        echo "  cd $PRIVATE_DIR"
        echo "  git remote add origin <your-private-repo-url>"
        echo "  git push -u origin main"
    fi

    echo ""
    echo "============================================"
    ok "Private overlay created!"
    echo "============================================"
    echo ""
    echo "Next steps:"
    echo "  1. Add your private configs (symlinks, Brewfile, skills)"
    echo "  2. Run: dot sync"
    echo "  3. Optionally push to a private git remote"
}

# Show private overlay status.
private_status() {
    echo -e "${BOLD:-}Private Overlay${NC:-}"
    echo ""

    if ! has_private_overlay; then
        warn "No private overlay found at $PRIVATE_DIR"
        echo "  Run: dot private init"
        return
    fi

    ok "Location: $PRIVATE_DIR"

    # Git info
    if has_private_git; then
        local branch
        branch="$(git -C "$PRIVATE_DIR" branch --show-current 2>/dev/null || echo "unknown")"
        echo -e "  Branch:  $branch"

        parse_git_status "$PRIVATE_DIR"
        if [[ ${#_GIT_MODIFIED[@]} -eq 0 && ${#_GIT_ADDED[@]} -eq 0 && ${#_GIT_DELETED[@]} -eq 0 ]]; then
            ok "Clean (no uncommitted changes)"
        else
            warn "Uncommitted changes:"
            [[ ${#_GIT_MODIFIED[@]} -gt 0 ]] && echo "    Modified: $(IFS=', '; echo "${_GIT_MODIFIED[*]}")"
            [[ ${#_GIT_ADDED[@]} -gt 0 ]]    && echo "    New:      $(IFS=', '; echo "${_GIT_ADDED[*]}")"
            [[ ${#_GIT_DELETED[@]} -gt 0 ]]  && echo "    Deleted:  $(IFS=', '; echo "${_GIT_DELETED[*]}")"
        fi

        # Remote status
        if git -C "$PRIVATE_DIR" remote get-url origin &>/dev/null; then
            ok "Remote: $(git -C "$PRIVATE_DIR" remote get-url origin)"
        else
            info "No remote configured"
        fi
    else
        info "Not a git repository"
    fi

    # Symlink map
    echo ""
    local private_map="$PRIVATE_DIR/symlink_map.txt"
    if [[ -f "$private_map" ]]; then
        local count
        count=$(grep -cv '^\s*#\|^\s*$' "$private_map" 2>/dev/null || echo 0)
        info "Symlinks: $count"
        while IFS= read -r line || [[ -n "$line" ]]; do
            [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
            local name dest
            name="$(basename "$(echo "$line" | cut -d: -f1)")"
            dest="$(echo "$line" | cut -d: -f2-)"
            echo "    $name → $dest"
        done < "$private_map"
    fi

    # Skills
    if [[ -d "$PRIVATE_DIR/skills" ]]; then
        local skill_count
        skill_count=$(find "$PRIVATE_DIR/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l | tr -d ' ')
        if [[ "$skill_count" -gt 0 ]]; then
            info "Skills: $skill_count"
        fi
    fi
}

# Pull latest changes from private overlay remote.
private_pull() {
    if ! has_private_overlay; then
        err "No private overlay found. Run: dot private init"
        return 1
    fi
    if ! has_private_git; then
        err "Private overlay is not a git repository"
        return 1
    fi
    info "Pulling latest..."
    git -C "$PRIVATE_DIR" pull
}

# Commit and push private overlay changes.
# Args: $1 = commit message (optional, auto-generated if omitted)
# Requires: lib/git.sh to be sourced
private_push() {
    if ! has_private_overlay; then
        err "No private overlay found. Run: dot private init"
        return 1
    fi
    if ! has_private_git; then
        err "Private overlay is not a git repository"
        return 1
    fi

    git_push_with_preview "$PRIVATE_DIR" "private" "${1:-}"
}

# Re-apply private symlinks.
# Requires: lib/paths.sh and lib/backup.sh to be sourced (for create_private_symlink).
private_sync() {
    if ! has_private_overlay; then
        err "No private overlay found. Run: dot private init"
        return 1
    fi

    section "Applying private symlinks..."

    local private_map
    private_map="$(get_private_symlink_map)"
    apply_private_symlinks "$private_map" "$PRIVATE_DIR"

    # Private skills
    if [[ -d "$PRIVATE_DIR/skills" ]]; then
        for skill in "$PRIVATE_DIR/skills"/*/; do
            [[ -d "$skill" ]] || continue
            local skill_name
            skill_name=$(basename "$skill")
            [[ "$skill_name" == ".gitkeep" ]] && continue
            local dest="$HOME/.claude/skills/$skill_name"
            mkdir -p "$(dirname "$dest")"
            create_private_symlink "skills/$skill_name" "$dest" "$PRIVATE_DIR"
        done
    fi

    ok "Private symlinks applied."
}

# ── Managed files ─────────────────────────────────────────────────────────────
# Merge base + private overlay JSON files and write to destination.
# Reads a managed_map.txt with format: base[+overlay]:destination

apply_managed_files() {
    local map_file="$1"
    [[ ! -f "$map_file" ]] && return 0

    # Require jq for JSON merging
    if ! command -v jq &>/dev/null; then
        warn "jq not found — skipping managed files (install jq via Homebrew)"
        return 0
    fi

    section "Merging managed files..."

    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip comments and blank lines
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

        # Parse format: base[+overlay]:destination
        local spec="${line%%:*}"
        local dest="${line#*:}"
        dest=$(eval echo "$dest")  # Expand $HOME

        local base_rel overlay_rel
        if [[ "$spec" == *"+"* ]]; then
            base_rel="${spec%%+*}"
            overlay_rel="${spec#*+}"
        else
            base_rel="$spec"
            overlay_rel=""
        fi

        local base_path="$DOTFILES_DIR/$base_rel"
        local overlay_path=""
        [[ -n "$overlay_rel" ]] && overlay_path="$PRIVATE_DIR/$overlay_rel"

        local have_base=false have_overlay=false
        [[ -f "$base_path" ]] && have_base=true
        [[ -n "$overlay_path" && -f "$overlay_path" ]] && have_overlay=true

        if [[ "$have_base" == false && "$have_overlay" == false ]]; then
            warn "Skipping $(basename "$dest") — no base or overlay found"
            continue
        fi

        # Ensure parent directory exists
        mkdir -p "$(dirname "$dest")"

        if [[ "$have_base" == true && "$have_overlay" == true ]]; then
            jq -s '.[0] * .[1]' "$base_path" "$overlay_path" > "$dest"
            ok "Merged base + private → $dest"
        elif [[ "$have_base" == true ]]; then
            cp "$base_path" "$dest"
            ok "Copied base → $dest (no private overlay)"
        else
            cp "$overlay_path" "$dest"
            ok "Copied private → $dest (no base)"
        fi
    done < "$map_file"
}
