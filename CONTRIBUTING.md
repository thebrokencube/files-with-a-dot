# Contributing

Thanks for helping with files-with-a-dot — a dotfiles manager that also ships Go CLI tools
(`folio`, `jf`, `dendrik`, `dot`) as cross-harness plugins. This is the contributor how-to. For
the plugin/marketplace model see [AGENTS.md](AGENTS.md); for releases (maintainer-only) see
`pkg/dendrik/conventions/release.md`.

## Setup

```sh
git clone git@github.com:thebrokencube/files-with-a-dot.git
cd files-with-a-dot
```

You need Go (per each tool's `go.mod`) to build the CLIs, and the repo uses a `go.work` across
the tool modules. `dot validate` and `dendrik lint` are the local gates — both build from source.

## Make a change

Work by layer; verify with that layer's gate. `<tool-root>` is the tool's module directory —
`cmd/<tool>/` today, or its own repo if a tool is extracted later.

| Layer | Where | Verify |
|---|---|---|
| A tool's Go CLI | `<tool-root>/` | `cd <tool-root> && make check` (fmt + vet + test) |
| A tool's skill / agentic doc | `<tool-root>/skill/` | `dendrik lint <tool-root>`; `/dendrik review <file>` |
| Repo config / dotfiles | `configs/`, scripts, manifests | `dot validate` |

Editing the plugin registry? Edit `plugins.json` only, then run `scripts/marketplace-generate`
(never hand-edit the generated catalogs) — CI fails on drift.

## Submit a PR

1. Branch off `main`; use conventional commits (`type(scope): description` — the `/commit` skill
   enforces the format).
2. Open a PR to `main`. CI must pass — per-tool `make check`, `pkg-dendrik`, and
   `version-consistency` — and `main` is protected, so a red PR can't merge.
3. Keep the change scoped; one logical change per PR. Releases are maintainer-only
   (`workflow_dispatch`) — don't bump versions or cut a release in a contribution PR.

## Test a plugin locally before release

```
/plugin marketplace add ./
/plugin install <tool>@files-with-a-dot
```

Installs from your working copy so you can exercise the skill and its `bin/setup` before anything
ships. (Claude Code is the golden path; Cursor/Codex are best-effort.)
