# Isolated Checkouts — Runbook

Procedures pulled out of `~/.claude/rules/isolated-checkouts.md` so they cost nothing until needed. The
rule states the law; this file has the commands. Read it when cleaning up, or when dotfiles work needs a
workspace.

## Cleaning up

- `folio fleet workarea list` — every area folio placed, plus every area it didn't. The second group is
  read straight from `jj workspace list` / `git worktree list`, so a hand-made checkout and a registration
  whose directory was deleted both show up.
- `folio fleet workarea reap [--force]` — removes areas folio placed. It never touches an area it didn't
  place; those are reported with the command to run.
- `folio home workspace cleanup <path>` — a KB session workspace. Only ever your own; the list does not
  say which workspaces belong to live sessions.

**Deleting the directory is not enough.** jj and git keep the registration, so the area lingers as
`dangling` in `workarea list` until you `jj workspace forget <name>` or `git worktree prune`.

## `~/.dotfiles` — done by hand

`folio fleet workarea` refuses the `dot` kind, so dotfiles isolation is manual, and still required. The
default `~/.dotfiles` workspace is where the user and other sessions land, so moving its `@` shifts the
working copy under them.

```
jj workspace add --name fwad-<topic> -r main /tmp/fwad-<topic>
```

Work entirely under that path, and `jj workspace forget fwad-<topic>` when the track is done. A jj
workspace has no `.git` — only the main checkout is colocated — so push the bookmark from the workspace
(`jj git push -b <bookmark>`) and run `gh pr create` from `~/.dotfiles`.

Already made edits in the default workspace? Move them out rather than continuing:

```
cd ~/.dotfiles && jj describe -m "<message>"          # name the change so you can find it
jj workspace add --name fwad-<topic> -r main /tmp/fwad-<topic>
jj new main                                            # default lets go of the change
cd /tmp/fwad-<topic> && jj edit <change-id>            # workspace picks it up, files intact
```
