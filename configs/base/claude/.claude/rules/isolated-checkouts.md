# Isolated Checkouts

An isolated checkout is a second working copy of a repo you already have — a jj workspace or a git
worktree — so parallel work doesn't fight over one working directory.

**Never create one beside the repo it came from.** `jj workspace add ../myrepo-feature` and
`git worktree add ../myrepo-feature` are the wrong call every time. So is anything inside the repo
(`.claude/worktrees/`), which the repo's own tooling then has to gitignore.

## The two roots

| Working on | Root | Command |
|---|---|---|
| A code repo | `~/.folio/.worktrees/<store>/<slug>` | `folio fleet workarea open <store> <branch>` |
| A folio KB store | `/tmp/folio-ws-<id>` | `folio home workspace create` |

`<store>` is the name in `~/.folio/stores.yml` (`gdl-groot`, `zenpayroll`, …), not the directory
name. Run `folio stores list` to see them.

Two roots because the lifetimes differ. A code work area is durable and survives the session. A KB
session workspace is ephemeral.

## Why the CLI rather than raw jj/git

`open` probes the repo to pick the isolation tier, and records the area in a ledger. That ledger is
what lets `reap` clean up correctly later — removal differs per tier, and `rm -rf` on a git worktree
strands a registration that then blocks re-checkout of that branch anywhere else. A repo that can
isolate neither way is refused with a message rather than guessed at; work it in place.

## Fetch before you branch

`jj git fetch` / `git fetch origin` first, and base new work on `origin/main`, not local `main`. Two of four
parallel worktrees once branched off a local main that was 252 commits stale, and recovery was cherry-picks.

## cwd persists between Bash calls — always `cd` in the same call

Two colocated jj repos are usually in play: the code repo and a folio workspace. cwd carries over silently,
so a bare `jj git push` after a `folio observe` pushes the code branch to the **folio** remote. That has
happened.

- Prefix every jj/git command with an explicit `cd <repo> && …` in the *same* call.
- Before any `jj git push`, confirm `git remote get-url origin` is the repo you meant.
- Run `folio home workspace create` from a neutral cwd (`cd /tmp`, `cd ~`). Inside a jj repo it creates a
  workspace *of that repo* instead — recover with `jj workspace forget <name>` plus `/bin/rm -rf <path>`.

## Never operate from a checkout you don't own

A repo's main checkout, and every other session's work area, belong to someone else. Two rules follow,
and both are easy to break while doing something that feels read-only.

**Target the repo with `-R`, don't `cd` into it.** `jj -R <repo-path> <cmd>` runs against the repo
without entering it. Needed for repo-level operations — `bookmark create`, `git push`,
`workspace forget`, `abandon` — which act on the repo, not on a working copy. (`git -C` is banned
separately; for git-only repo-level work, prefer the folio CLI or ask.)

**Pass `--ignore-working-copy` for anything that isn't yours.** Plain `jj log`/`jj status` snapshots
the working copy it runs in. Run it inside another session's workspace and you have silently committed
their in-flight edits into their `@`.

**Never rewrite a commit that is another workspace's `@`.** `jj describe -r <other-workspace>@` on a
workspace whose working copy is stale creates a **divergent change** — two commits sharing one change
ID — which the other session then has to untangle. Bookmark the commit instead; a bookmark adds a ref
and rewrites nothing.

## Cleaning up

- `folio fleet workarea list` — every area folio placed, plus every area it didn't. The second group
  is read straight from `jj workspace list` / `git worktree list`, so a hand-made checkout and a
  registration whose directory was deleted both show up.
- `folio fleet workarea reap [--force]` — removes areas folio placed. It never touches an area it
  didn't place; those are reported with the command to run.
- `folio home workspace cleanup <path>` — a KB session workspace. Only ever your own; the list does
  not say which workspaces belong to live sessions.

**Deleting the directory is not enough.** jj and git keep the registration, so the area lingers as
`dangling` in `workarea list` until you `jj workspace forget <name>` or `git worktree prune`.

## `~/.dotfiles` is the exception

`folio fleet workarea` refuses the `dot` kind, so dotfiles isolation is still done by hand — and it
is still required. The default `~/.dotfiles` workspace is where the user and other sessions land, so
moving its `@` shifts the working copy under them.

```
jj workspace add --name fwad-<topic> -r main /tmp/fwad-<topic>
```

Work entirely under that path, and `jj workspace forget fwad-<topic>` when the track is done. A jj
workspace has no `.git` — only the main checkout is colocated — so push the bookmark from the
workspace (`jj git push -b <bookmark>`) and run `gh pr create` from `~/.dotfiles`.

Already made edits in the default workspace? Move them out rather than continuing:

```
cd ~/.dotfiles && jj describe -m "<message>"          # name the change so you can find it
jj workspace add --name fwad-<topic> -r main /tmp/fwad-<topic>
jj new main                                            # default lets go of the change
cd /tmp/fwad-<topic> && jj edit <change-id>            # workspace picks it up, files intact
```
