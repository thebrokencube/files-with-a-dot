# Container Migration Runbook

One-time migration of a single-home `~/.folio` into the **multi-store container**
model. This is the **mandatory** path forward — single-home is a transitional
bridge slated for removal. Packaged as a runbook (not a `folio migrate` command)
because it is one-time and destructive.

```
BEFORE: ~/.folio                     (colocated git+jj repo, single-home)

AFTER:  ~/.folio/                     (plain umbrella DIRECTORY — not a repo)
        ├── thebrokencube-folio/      (fresh colocated clone = the work store, default)
        ├── folio-vault/              (personal vault store — optional, additive)
        ├── adr/ radr/ ...            (external KB clones — optional, read-only)
        └── stores.yml                (registry; dotfile-managed)
```

**Why clone-beside, not move:** jj workspaces store **absolute** back-pointers to
the repo root. Moving the root (`mv ~/.folio ~/.folio/thebrokencube-folio`)
corrupts all ~25 workspaces. A fresh `git clone` from origin is complete once all
work is pushed; a `cp -a` backup + two-`mv` swap keep every step reversible.

## Prerequisite — binary-first ordering (NON-NEGOTIABLE)

`folio >= 0.0.4` (the v2 binary that understands `stores.yml`) **must already be
on PATH before any `stores.yml` exists**. An older binary treats the umbrella as
the home and silently stops creating workspaces.

```bash
folio --version          # must be >= 0.0.4
# if not: bump cmd/folio/VERSION, dispatch release.yml -f tool=folio, then 'dot sync'
```

## Step 1 — dry-run, then execute against the live tree

The script (`cmd/folio/scripts/migrate-container.sh`) defaults to `--check`
(dry-run, zero mutations). It gates on a clean/pushed state and aborts before the
irreversible swap. Read the dry-run output in full first.

```bash
bash cmd/folio/scripts/migrate-container.sh --check     # dry-run — mutates nothing
bash cmd/folio/scripts/migrate-container.sh --execute   # real: backup → clone → verify → swap
```

What `--execute` does, in order:

1. **Preconditions** — version ≥ 0.0.4, `~/.folio` is colocated git+jj, no existing `stores.yml`.
2. **Push gate** — abort unless `main == main@origin` and there are no unpushed non-empty changes (a fresh clone would lose them).
3. **Backup** — `cp -a ~/.folio ~/.folio.backup-<stamp>`.
4. **Forget workspaces** — `jj workspace forget` every workspace (incl. this session's).
5. **Clone-beside** — clone origin into `~/.folio.new-<stamp>/thebrokencube-folio`, then `jj git init --colocate`.
6. **Verification gate** — abort *before* the swap unless the clone is colocated and its active-project count matches the backup.
7. **Swap** — two `mv`s: `~/.folio` → `~/.folio.old-<stamp>`, staging → `~/.folio`.
8. **Bootstrap `stores.yml`** — minimal registry so folio works immediately.
9. **Smoke test** — `folio home list` from the umbrella.

If the push gate fails: `folio home push` from each workspace that has unpushed
work, then retry. The backup also captures anything uncommitted as a safety net.

## Step 2 — make `stores.yml` dotfile-managed (reproducible)

The script writes a **bootstrap** `stores.yml`. Replace it with the
dotfile-managed one so it is reproducible and survives `dot sync`. `dot` merges
the base + private overlay as YAML (`yq`: maps deep-merge, scalars overlay-win).

**Base** (`~/.dotfiles/configs/base/folio/stores.base.yml`) — shared; valid on its
own (a personal one-store container, no work store):

```yaml
schema: 2
default: folio-vault
stores:
  folio-vault: { path: ~/.folio/folio-vault, kind: folio, remote: git@github.com:thebrokencube/folio-vault.git }
```

**Private overlay** (`~/.dotfiles.private/folio-stores.yml`) — work machine only;
adds the Gusto stores and overrides the default:

```yaml
default: thebrokencube-folio
stores:
  thebrokencube-folio: { path: ~/.folio/thebrokencube-folio, kind: folio, remote: git@github.com:Gusto/thebrokencube-folio.git }
  adr:  { path: ~/.folio/adr,  kind: external, remote: <adr-remote> }      # fill in real remotes
  radr: { path: ~/.folio/radr, kind: external, remote: <radr-remote> }
  carrier-platform-llm-wiki: { path: ~/.folio/carrier-platform-llm-wiki, kind: external, remote: <llm-wiki-remote> }
```

**Register** in `~/.dotfiles/managed_map.txt`:

```
configs/base/folio/stores.base.yml+folio-stores.yml:$HOME/.folio/stores.yml
```

Then preview + apply (separate dotfiles commits per repo, conventional commits):

```bash
cd ~/.dotfiles && dot sync --dry-run   # confirm the merged stores.yml round-trips
dot validate                           # pre-existing warnings in unrelated files are expected
dot sync                               # deploys ~/.folio/stores.yml (supersedes bootstrap)
```

Decoupling check: a machine with **base only** (no private overlay) must still
produce a valid `stores.yml` (just `folio-vault` + `default: folio-vault`).

## Step 3 — (optional) clone vault + external KBs as siblings

```bash
git clone git@github.com:thebrokencube/folio-vault.git ~/.folio/folio-vault
git clone <adr-remote>  ~/.folio/adr        # external (read-only to folio)
git clone <radr-remote> ~/.folio/radr
```

External stores: `folio home pull <store>` refreshes them; folio **never** pushes
them — contribute via each repo's own PR flow.

## Step 4 — re-create this session's workspace

The migration forgot every workspace. Before further `/folio` work:

```bash
folio home workspace create   # resolves the default store via the registry
```

## Step 5 — verify, then clean up

```bash
folio home list                         # lists projects in the default store
folio home push thebrokencube-folio -m "test(folio): post-migration smoke"   # (or a real change)
folio home pull adr                     # external pull works
# from inside ~/.folio/folio-vault/... a bare command acts on folio-vault (cwd override)
```

After a confidence period, remove the safety copies:

```bash
rm -rf ~/.folio.old-<stamp> ~/.folio.backup-<stamp>
```

## Second machine (personal) — one-store container

The personal machine is a **one-store container** (umbrella + `stores.yml` with
only `folio-vault`). It never clones the Gusto work repo, and it must have **no**
`~/.dotfiles.private/folio-stores.yml` — then `dot sync` builds `stores.yml` from
the base half alone (`folio-vault` + `default: folio-vault`).

**Binary first** (same rule): `git pull` + `dot sync` so folio ≥ 0.0.4 is on PATH.
Confirm `folio --version`. That `dot sync` may also deploy `stores.yml` onto the
still-single-home `~/.folio` (folio then points at a not-yet-existing nested
store) — harmless: the migration auto-removes that stray `stores.yml` and
regenerates it post-swap.

Then pick by the current state of that machine's `~/.folio`:

- **Already a single-home folio repo** (origin = `folio-vault`): run the same
  migration — the store name auto-derives from origin, so no flag needed:
  ```bash
  bash cmd/folio/scripts/migrate-container.sh --check     # → store: folio-vault
  bash cmd/folio/scripts/migrate-container.sh --execute
  ```
- **Empty / absent**: bootstrap fresh instead of migrating:
  ```bash
  mkdir -p ~/.folio
  git clone git@github.com:thebrokencube/folio-vault.git ~/.folio/folio-vault
  ( cd ~/.folio/folio-vault && jj git init --colocate )
  ```

Either way, finish by replacing the bootstrap with the dotfile-managed registry
(`dot sync` deploys the base-only `stores.yml`; if the bootstrap is flagged as
drift, `rm ~/.folio/stores.yml` then re-`dot sync`), and re-create the session
workspace: `folio home workspace create`.

## Rollback

Before deleting `~/.folio.old-<stamp>`:

```bash
rm -rf ~/.folio && mv ~/.folio.old-<stamp> ~/.folio
```

This restores the original colocated repo. Workspaces were forgotten; re-create
with `folio home workspace create` (it falls back to single-home with no
`stores.yml`). If `stores.yml` was already dotfile-managed, also remove the
`managed_map.txt` entry and re-`dot sync`, or `git checkout` it away.
