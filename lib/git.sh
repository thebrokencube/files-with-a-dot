#!/bin/bash
# lib/git.sh - Shared git helpers for push and status workflows
#
# Usage: source "$(dirname "$0")/lib/git.sh"
# Requires: lib/colors.sh, lib/logging.sh, lib/prompt.sh to be sourced

# Parse git status --short into _GIT_MODIFIED, _GIT_ADDED, _GIT_DELETED arrays.
# Args: $1 = repo dir
parse_git_status() {
    local repo_dir="$1"
    local status_output
    status_output="$(git -C "$repo_dir" status --short 2>/dev/null)"

    _GIT_MODIFIED=()
    _GIT_ADDED=()
    _GIT_DELETED=()

    while IFS= read -r line; do
        [[ -z "$line" ]] && continue
        local code="${line:0:2}"
        local file="${line:3}"
        case "$code" in
            " M"|"M "|"MM"|"AM") _GIT_MODIFIED+=("$file") ;;
            "??"|"A "|" A")      _GIT_ADDED+=("$file") ;;
            " D"|"D ")           _GIT_DELETED+=("$file") ;;
            *)                   _GIT_MODIFIED+=("$file") ;;
        esac
    done <<< "$status_output"
}

# Full preview/confirm/commit/push workflow.
# Args: $1 = repo dir, $2 = scope (for commit message prefix), $3 = message (optional)
git_push_with_preview() {
    local repo_dir="$1"
    local scope="$2"
    local user_msg="${3:-}"

    # Check if there are changes
    if git -C "$repo_dir" diff --quiet 2>/dev/null && \
       git -C "$repo_dir" diff --cached --quiet 2>/dev/null && \
       [[ -z "$(git -C "$repo_dir" ls-files --others --exclude-standard 2>/dev/null)" ]]; then
        ok "Nothing to push (working tree clean)"
        return 0
    fi

    # ── Show changed files summary ────────────────────────────────────────
    parse_git_status "$repo_dir"

    echo ""
    if [[ ${#_GIT_MODIFIED[@]} -gt 0 ]]; then
        info "Modified: $(IFS=', '; echo "${_GIT_MODIFIED[*]}")"
    fi
    if [[ ${#_GIT_ADDED[@]} -gt 0 ]]; then
        info "New:      $(IFS=', '; echo "${_GIT_ADDED[*]}")"
    fi
    if [[ ${#_GIT_DELETED[@]} -gt 0 ]]; then
        info "Deleted:  $(IFS=', '; echo "${_GIT_DELETED[*]}")"
    fi
    echo ""

    # ── Show diff (interactive only) ──────────────────────────────────────
    if _is_interactive; then
        local diff_output
        diff_output="$(git -C "$repo_dir" diff --stat 2>/dev/null; git -C "$repo_dir" diff --cached --stat 2>/dev/null)"
        diff_output+=$'\n'"$(git -C "$repo_dir" diff 2>/dev/null; git -C "$repo_dir" diff --cached 2>/dev/null)"
        # Show untracked file content too
        local untracked
        untracked="$(git -C "$repo_dir" ls-files --others --exclude-standard 2>/dev/null)"
        if [[ -n "$untracked" ]]; then
            while IFS= read -r f; do
                diff_output+=$'\n'"diff --git a/$f b/$f"
                diff_output+=$'\n'"new file"
                diff_output+=$'\n'"$(cat "$repo_dir/$f" 2>/dev/null)"
            done <<< "$untracked"
        fi

        if [[ -n "$diff_output" ]]; then
            echo "$diff_output"
            echo ""
        fi
    fi

    # ── Confirm (interactive only) ────────────────────────────────────────
    if ! confirm "Commit and push?"; then
        echo "Aborted."
        return 1
    fi

    # ── Build commit message ──────────────────────────────────────────────
    local msg
    if [[ -n "$user_msg" ]]; then
        msg="$user_msg"
    else
        local all_files=("${_GIT_MODIFIED[@]}" "${_GIT_ADDED[@]}" "${_GIT_DELETED[@]}")
        local basenames=()
        for f in "${all_files[@]}"; do
            basenames+=("$(basename "$f")")
        done

        if [[ ${#basenames[@]} -le 3 ]]; then
            msg="chore($scope): update $(IFS=', '; echo "${basenames[*]}")"
        else
            local file_list
            file_list="$(printf '%s\n' "${basenames[@]}")"
            msg="$(printf 'chore(%s): update dotfiles\n\n%s' "$scope" "$file_list")"
        fi
    fi

    # ── Stage, commit, push ───────────────────────────────────────────────
    info "Committing: $msg"
    (
        cd "$repo_dir" || return
        git add -A
        git commit -m "$msg"
    )

    if git -C "$repo_dir" remote get-url origin &>/dev/null; then
        info "Pushing to remote..."
        git -C "$repo_dir" push
        ok "Pushed."
    else
        warn "No remote configured. Committed locally only."
        echo "  To add a remote: cd $repo_dir && git remote add origin <url>"
    fi
}
