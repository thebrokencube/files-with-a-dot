# Architecture

Three sections: **Design** (target state), **Decisions** (why), **Migration** (progress).

---

## Design

### Functional Core, Imperative Shell (FCIS)

**lib/** contains reusable functions (the functional core). Scripts are thin orchestrators that source lib/ and sequence calls (the imperative shell).

### lib/ Responsibilities

| Library | Purpose | Key Functions |
|---------|---------|---------------|
| `colors.sh` | Color constants + symbols | `RED`, `GREEN`, `SYM_OK`, etc. |
| `logging.sh` | Structured output | `ok()`, `err()`, `warn()`, `info()`, `section()`, `debug()` |
| `paths.sh` | Path utilities | `get_source()`, `get_dest()`, `is_ours()`, `resolve_path()` |
| `backup.sh` | Backup operations | `init_backup()`, `backup_file()`, `sync_backups()`, `restore_backup()` |
| `symlinks.sh` | Symlink management | `check_symlink()`, `create_symlink()`, `apply_symlinks()` |
| `private.sh` | Private overlay | `has_private_overlay()`, `apply_private_symlinks()`, `migrate_private_overlay()`, `init_private_overlay()`, `private_status()`, `private_sync()`, `private_push()` |
| `config.sh` | Machine config | `read_machine_type()`, `get_dotfiles_version()`, `init_dotfiles_vars()` |
| `prompt.sh` | User interaction | `confirm()`, `choose()`, `require_interactive()` |
| `brew.sh` | Homebrew operations | `build_merged_brewfile()`, `detect_brew_cleanup()`, `detect_brew_autoremove()`, `detect_brew_cache()`, `execute_brew_cleanup()` |
| `shell.sh` | Shell config integration | `get_shell_configs()`, `check_source_line()`, `add_source_line()`, `remove_source_line()`, `integrate_shell_configs()`, `remove_shell_configs()` |

### Script Roles

| Script | Role | Sources |
|--------|------|---------|
| `dot` | CLI entry + TTY detection + orchestration | `lib/config.sh` |
| `sync.sh` | Main sync orchestrator (~300 lines) | All lib/ |
| `cleanup.sh` | System cleanup orchestrator | `lib/{colors,logging,config,prompt,brew,private}.sh` |
| `health.sh` | Diagnostics + setup guide | `lib/{colors,logging,config,prompt}.sh` |
| `uninstall.sh` | Clean removal | `lib/{colors,logging,config,paths,backup,symlinks,shell}.sh` |
| `bootstrap.sh` | Self-contained clone script | None (repo not yet cloned) |

### Non-Interactive Protocol

**`DOTFILES_NONINTERACTIVE=1`** is the single env var controlling non-interactive behavior.

Set automatically by:
- `dot` CLI when stdin is not a TTY
- Programmatic callers: `DOTFILES_NONINTERACTIVE=1 dot sync`

**`--force`** is a per-script local flag. Does NOT propagate via env var. Parents pass `--force` to children explicitly.

**`confirm [-f] PROMPT [DEFAULT]`** behavior:

| Condition | Behavior |
|-----------|----------|
| `-f` flag passed | Always return 0 (yes) |
| `DOTFILES_NONINTERACTIVE=1` | Return DEFAULT (`yes`->0, `no`->1) |
| Not a TTY, no env var | Return DEFAULT + warn on stderr |
| Interactive TTY | Show prompt, read answer |

Auto-appends `[Y/n]` or `[y/N]` based on default. Scripts use:
```bash
confirm ${FORCE:+-f} "Continue?" "no"
```

### Target Topology

```
dot (CLI + TTY detection + orchestration)
 ├── sync: sync.sh -> cleanup.sh  (dot sequences, not sync.sh)
 ├── clean: cleanup.sh
 ├── health: health.sh
 ├── validate: validate.sh
 └── private: lib/private.sh functions (init, sync, push, status, edit)

bootstrap.sh (standalone, self-contained) -> sync.sh only
uninstall.sh (standalone leaf)
```

### Conventions

- **Bash 3.2 compatible** (macOS system bash): no namerefs, no associative arrays
- **`set -e`** in all scripts
- **Source order**: colors -> logging -> (other libs as needed)
- **Error handler**: trap ERR for detailed error info
- **Every commit**: update this checklist + run `dot validate`

---

## Decisions

### D1: FCIS over monolithic scripts

**Problem**: 6 scripts each define their own colors (10 lines), path helpers (20 lines), backup logic (60 lines). sync.sh is 975 lines, mostly duplicated from lib/.

**Decision**: lib/ is the single source of truth for reusable logic. Scripts become thin orchestrators.

**Trade-off**: More files to navigate, but each file has a single responsibility. Duplicated code drops to zero.

### D2: `DOTFILES_NONINTERACTIVE` env var over `--skip-prompts`

**Problem**: 6 different interactivity mechanisms (`--skip-prompts`, `--confirmed`, `--force`, `NONINTERACTIVE=1`, direct `read -p` checks, no guard at all). `read -p` hangs when piped from agents/CI.

**Decision**: Single env var `DOTFILES_NONINTERACTIVE=1`. `--force` stays as a local per-script flag (always-yes, doesn't propagate). `dot` auto-sets the env var when stdin isn't a TTY.

**Trade-off**: Existing `--skip-prompts` flag in sync.sh is removed (breaking change). Worth it because the new protocol is consistent everywhere and auto-detected.

### D3: `dot` orchestrates sync -> cleanup (not sync.sh)

**Problem**: sync.sh calls cleanup.sh directly at the end, mixing concerns. Can't run sync without cleanup. Can't test them independently.

**Decision**: `dot sync` runs sync.sh then cleanup.sh sequentially. sync.sh no longer calls cleanup. Each script is independently runnable and testable.

**Trade-off**: Running `./sync.sh` directly no longer triggers cleanup. But `dot sync` is the intended interface.

### D4: Delete brew-cleanup.sh

**Problem**: brew-cleanup.sh manually parses Brewfiles with regex and compares against installed packages. cleanup.sh uses `brew bundle cleanup` which does the same thing natively and correctly.

**Decision**: Delete brew-cleanup.sh. Extract the Brewfile-merging logic to lib/brew.sh. cleanup.sh uses `brew bundle cleanup` with the merged Brewfile.

### D5: Delete init-private.sh, move to `dot private init`

**Problem**: init-private.sh is a standalone script for a one-time operation. No discoverability via the `dot` CLI.

**Decision**: Move all private overlay logic to lib/private.sh. Add `dot private` subcommand with init/sync/push/status/edit. Delete init-private.sh.

### D6: Remove `dot setup` — fold into sync + health

**Problem**: `dot setup` dispatched to `health.sh --setup`, an interactive walkthrough (5 steps with `read -p` between each). But most steps were manual actions in other apps (open nvim, select iTerm2 profile, run `claude auth`). The walkthrough format didn't add value over just listing the steps. It also coupled setup concerns into the diagnostics script.

**Decision**: In v0.8.0, remove `run_interactive_setup()` and `--setup` from health.sh, remove `dot setup` from the CLI. Instead: sync.sh shows comprehensive "Next steps" after first-time setup, and `dot health` continues to detect pending items via `check_setup_status()`. Until then, setup is guarded with `require_interactive`.

### D7: bootstrap.sh stays self-contained

**Problem**: bootstrap.sh runs before the repo is cloned, so it can't source lib/.

**Decision**: Keep bootstrap.sh self-contained (~60-80 lines). Its only job is: install Xcode CLT, clone repo, exec sync.sh. Homebrew prerequisite check moves to sync.sh (which can use lib/).

