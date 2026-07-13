---
name: commit
description: Conventional commit format, message conventions, green commit rules, VCS history structuring for both git and jj (Jujutsu), and PR title/description house style. Use when creating commits, amending commits, writing commit messages, describing jj changes (jj describe, jj new, jj split, jj squash), fixup commits, interactive rebase, stacked branch propagation, deciding how to decompose work into commits, or writing a PR title/body (including whether to use a repo PR template or the house style, and the stack list for stacked PRs). Applies to all repositories — always use this skill instead of Claude Code's default commit behavior.
user_invocable: false
---

# Commit Conventions

These conventions apply to **all repositories** unless a project-specific CLAUDE.md overrides them.

## VCS Detection

```
Check `jj root 2>/dev/null`:
  exit 0 → jj repo, use jj sections below
  non-zero → git repo, use git sections below
```

**Colocated repos** (both `.jj` and `.git` present) use jj conventions since `jj root` succeeds. jj is the preferred VCS when available.

## Commit Message Format

Use **conventional commits** with **required scope** (applies to both git and jj):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types
- `feat` - New feature or capability
- `fix` - Bug fix
- `docs` - Documentation only
- `refactor` - Code change that neither fixes a bug nor adds a feature
- `test` - Adding or updating tests
- `chore` - Maintenance, dependencies, tooling
- `style` - Formatting, whitespace (no code change)
- `perf` - Performance improvement
- `auto` - Generated/deterministic output of a command (see Generated Commits in the stacking references)

### Scope
**Always include a scope.** Choose the most relevant:
- Component/module name: `feat(auth):`, `fix(api):`
- File/area: `refactor(user-model):`, `docs(readme):`
- Feature area: `feat(checkout):`, `fix(payments):`

### Description
- Imperative mood: "add" not "added" or "adds"
- Lowercase, no period at end
- Under 50 chars for the subject line (`auto` commits exempt when the command is longer)
- Be specific: "fix null pointer in user lookup" not "fix bug"

### Examples
```
feat(auth): add OAuth2 login flow
fix(api): handle null response from payments service
refactor(models): extract validation into concern
docs(api): add rate limiting section
chore(deps): bump rails to 7.1
auto(codegen): graphqlme
```

## Green Commits

**Every commit MUST leave the codebase in a working state:**

1. Tests pass (or at least don't introduce new failures)
2. Code compiles/lints
3. App can start

If a change requires multiple commits, structure them so each is independently valid. NEVER commit "WIP" or broken intermediate states to shared branches.

## Reviewable History

- One logical change per commit; don't mix refactoring with features
- Tests in same commit as code they test
- Refactors before features; deps in own commits

### Bad Example
```
feat(checkout): add stripe integration, fix tests, update deps, refactor utils
```

### Good Example
```
chore(deps): add stripe gem
refactor(payments): extract payment processor interface
feat(checkout): add stripe payment processor with tests
```

## Git: Single Change

**IMPORTANT: Always use these conventions. Do NOT use Claude Code's default commit behavior.**

- [ ] Check status: `git status` to see what's changed
- [ ] Check if behind: `git fetch origin` then `git log HEAD..origin/main --oneline`
- [ ] If behind: **NEVER auto-rebase.** Ask the user first — rebasing behind the user's back can silently break stacked branches or discard in-progress work.
- [ ] Stage thoughtfully: Group related changes, don't just `git add -A`
- [ ] Write message: Follow conventional commit format with scope
- [ ] No trailers: Do NOT add Co-Authored-By or other trailers
- [ ] Verify: Ensure commit leaves codebase green

### Amending vs New Commits
- **Amend** when fixing typos or small issues in the last unpushed commit
- **Fixup** (default on PR branches): `git commit --fixup=<sha>` + autosquash. Only standalone commit if the change is a distinct logical unit. See `references/git-stacking.md` Fixup Targeting section.
- **New commit** for distinct logical changes
- **Interactive rebase** to clean up before PR (squash fixups, reorder)
- **In a stack**: Any amend or interactive rebase requires propagating descendant branches. See `references/git-stacking.md` Propagation Workflow.

### Rebase
- If behind: **NEVER auto-rebase.** Ask the user first — rebasing behind the user's back can silently break stacked branches or discard in-progress work.
- Use `--force-with-lease` for feature branches. **NEVER** force push main/master without explicit user confirmation — this rewrites shared history and can break everyone's local state.

## jj: Single Change

- [ ] Check status: `jj status`
- [ ] Describe the change: `jj describe -m "type(scope): description"`
- [ ] Start next change: `jj new` (or `jj new -m "..."` to describe upfront)
- [ ] No trailers in descriptions
- [ ] Verify: Ensure change leaves codebase green

### Key Operations
- **Describe**: `jj describe -m "type(scope): description"` — set/update the commit message
- **New change**: `jj new` — start a new child change (current change is "committed")
- **Split**: `jj split` — break a change into two (interactive)
- **Squash**: `jj squash` — fold current change into parent
- **Rebase**: `jj rebase -d main` — move change onto new base

No staging area — the working copy IS the change. Files are auto-snapshotted.

### Amending
- Edit files directly — they auto-snapshot into the current change
- To fix a previous change: `jj edit <change-id>`, make edits, then `jj new` to return to tip
- To squash a fix into a specific target: `jj squash --into <change-id>`

## Git: Stacking

-> Read references/git-stacking.md for propagation, fixup targeting, generated commits, conflict resolution, push strategy, and common failure modes.

## jj: Stacking

-> Read references/jj-stacking.md for building stacks, bookmark management, push semantics, PR creation, review cycles, and post-merge cleanup.

## PR Descriptions

A house style for PR bodies applies to **all repos**. **If the repo has a PR template (`.github/pull_request_template.md`), ask the user whether to use the template or the house style** — don't silently pick (a settled per-repo preference in CLAUDE.md/memory overrides the ask).

-> Read references/pr-descriptions.md for the house style (title, What/How/Notes structure, diagram conventions, and the stack list for stacked PRs).
