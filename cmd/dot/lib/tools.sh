#!/usr/bin/env bash
# Install dendrik-built CLI tools (folio, jf, dendrik) from GitHub Releases.
#
# Replaces the old model of committing binaries to the repo and symlinking them.
# Each tool's version is pinned by cmd/<tool>/VERSION; we fetch the matching
# release asset for the host platform, and fall back to a local `go build` if the
# asset is unavailable and Go is installed.
#
# Replacement-safe: overwrites a prior real binary OR a now-dangling symlink left
# over from the pre-migration committed-binary layout (rm -f before install), so a
# `dot pull` on a machine that still has the old symlinks migrates cleanly.

DOT_TOOLS=(folio jf dendrik)
DOT_TOOLS_REPO="thebrokencube/files-with-a-dot"

# dot_host_platform sets DOT_OS / DOT_ARCH to release-asset naming (empty if unsupported).
dot_host_platform() {
    case "$(uname -s)" in
        Darwin) DOT_OS=darwin ;;
        Linux)  DOT_OS=linux ;;
        *)      DOT_OS="" ;;
    esac
    case "$(uname -m)" in
        arm64 | aarch64) DOT_ARCH=arm64 ;;
        x86_64 | amd64)  DOT_ARCH=amd64 ;;
        *)               DOT_ARCH="" ;;
    esac
}

# install_tool <tool>: install ~/.local/bin/<tool> from the pinned release, or build it.
install_tool() {
    local tool="$1"
    local bindir="$HOME/.local/bin"
    local dest="$bindir/$tool"
    local vfile="$DOTFILES_DIR/cmd/$tool/VERSION"

    [[ -f "$vfile" ]] || { warn "$tool: no VERSION file — skipping"; return 1; }
    local ver
    ver="$(tr -d '[:space:]' < "$vfile")"
    mkdir -p "$bindir"

    local url="https://github.com/${DOT_TOOLS_REPO}/releases/download/${tool}/v${ver}/${tool}-${DOT_OS}-${DOT_ARCH}"
    local tmp
    tmp="$(mktemp)"

    if [[ -n "$DOT_OS" && -n "$DOT_ARCH" ]] && curl -fsSL "$url" -o "$tmp" 2>/dev/null; then
        rm -f "$dest"          # drop a prior binary or a now-dangling old symlink
        mv "$tmp" "$dest"
        chmod +x "$dest"
        ok "$tool $ver (release $DOT_OS/$DOT_ARCH)"
        return 0
    fi
    rm -f "$tmp"

    # Fallback: build from source when the release asset is missing and Go is present.
    if command -v go > /dev/null 2>&1; then
        local out
        out="$(mktemp)"
        if (cd "$DOTFILES_DIR/cmd/$tool" && go build -trimpath -buildvcs=false \
            -ldflags "-buildid= -X main.version=${ver}" -o "$out" .) 2> /dev/null; then
            rm -f "$dest"
            mv "$out" "$dest"
            chmod +x "$dest"
            ok "$tool $ver (built locally — no release asset for $DOT_OS/$DOT_ARCH)"
            return 0
        fi
        rm -f "$out"
    fi

    warn "$tool $ver: no release asset ($DOT_OS/$DOT_ARCH) and no local Go build — install Go or check the release"
    return 1
}

# install_tools installs every dendrik-built CLI tool.
install_tools() {
    dot_host_platform
    for tool in "${DOT_TOOLS[@]}"; do
        install_tool "$tool"
    done
}
