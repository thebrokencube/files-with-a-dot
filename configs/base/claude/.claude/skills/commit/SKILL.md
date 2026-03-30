---
name: commit
description: Conventional commit format, message conventions, green commit rules, rebase-before-commit workflow. Use when creating commits, writing commit messages, or deciding how to structure git history. Applies to all repositories.
user_invocable: false
---

# Git Commit Conventions

These conventions apply to **all repositories** unless a project-specific CLAUDE.md overrides them.

## When Asked to Commit

**IMPORTANT: Always use these conventions. Do NOT use Claude Code's default commit behavior.**

- [ ] Check status: `git status` to see what's changed
- [ ] Check if behind: `git fetch && git log HEAD..origin/main --oneline`
- [ ] If behind: **NEVER auto-rebase.** Ask the user first.
- [ ] Stage thoughtfully: Group related changes, don't just `git add -A`
- [ ] Write message: Follow conventional commit format with scope
- [ ] No trailers: Do NOT add Co-Authored-By or other trailers
- [ ] Verify: Ensure commit leaves codebase green

## Commit Message Format

Use **conventional commits** with **required scope**:

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
- `auto` - Generated/deterministic output of a command (see stacked-pr skill, Generated Commits section)

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

## Amending vs New Commits

- **Amend** when fixing typos or small issues in the last unpushed commit
- **Fixup** (default on PR branches): When making a small correction to an existing commit on a branch with an open PR, prefer `git commit --fixup=<sha>` + autosquash over a standalone commit. Only create a standalone commit if the change is a distinct logical unit. See stacked-pr skill, Fixup Targeting section for pre-check and apply commands.
- **New commit** for distinct logical changes
- **Interactive rebase** to clean up before PR (squash fixups, reorder). Unless the user requests otherwise.
- **In a stack**: Any amend or interactive rebase requires propagating descendant branches. See stacked-pr skill, Fixup Targeting and Propagation Workflow sections.

## Rebase Workflow

Before committing, check if behind (`git fetch origin && git log HEAD..origin/main --oneline`).
- If behind: **NEVER auto-rebase.** Ask the user first.
- Use `--force-with-lease` for feature branches. **NEVER** force push main/master without explicit user confirmation.
- For multi-branch rebase scenarios (stacked PRs), see the stacked-pr skill, Propagation Workflow section.

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

