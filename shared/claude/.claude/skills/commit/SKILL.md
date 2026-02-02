---
name: commit
description: Git commit workflow and conventions. Use when committing code, creating PRs, or managing git history. Applies to all repositories.
user-invocable: false
---

# Git Commit Conventions

These conventions apply to **all repositories** unless a project-specific CLAUDE.md overrides them.

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

### Scope
**Always include a scope.** Choose the most relevant:
- Component/module name: `feat(auth):`, `fix(api):`
- File/area: `refactor(user-model):`, `docs(readme):`
- Feature area: `feat(checkout):`, `fix(payments):`

### Description
- Imperative mood: "add" not "added" or "adds"
- Lowercase, no period at end
- Under 50 chars for the subject line
- Be specific: "fix null pointer in user lookup" not "fix bug"

### Examples
```
feat(auth): add OAuth2 login flow
fix(api): handle null response from payments service
refactor(models): extract validation into concern
docs(api): add rate limiting section
chore(deps): bump rails to 7.1
```

### Tagged Commits (releases on main)

When committing directly to main with a version tag, prepend the version:

```
v1.2.0: feat(auth): add OAuth2 login flow
v0.3.2: fix(api): handle null response
```

The tag message should match the commit message.

## Green Commits

**Every commit should leave the codebase in a working state:**

1. Tests pass (or at least don't introduce new failures)
2. Code compiles/lints
3. App can start

If a change requires multiple commits, structure them so each is independently valid. Never commit "WIP" or broken intermediate states to shared branches.

## Rebase Workflow

**Before committing, check if you're behind:**

```bash
git fetch origin
git log HEAD..origin/main --oneline  # See what's new
```

If behind, **suggest rebasing** (don't auto-rebase):
- "You're X commits behind main. Want to rebase first?"
- User decides based on context (conflicts, urgency, etc.)

**When rebasing:**
```bash
git fetch origin
git rebase origin/main
# Resolve conflicts if any
git push --force-with-lease  # Safe force push for feature branches
```

**Never force push to main/master** without explicit user confirmation.

## Reviewable History

Structure commits for human reviewers:

1. **One logical change per commit** - Don't mix refactoring with features
2. **Tell a story** - Commits should read like a narrative of the change
3. **Tests with their code** - Include tests in the same commit as the code they test
4. **Separate concerns:**
   - Refactoring commits before feature commits
   - Dependency updates in their own commits

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

## When Asked to Commit

1. **Check status**: `git status` to see what's changed
2. **Check if behind**: `git fetch && git log HEAD..origin/main --oneline`
3. **If behind**: Suggest rebase, let user decide
4. **Stage thoughtfully**: Group related changes, don't just `git add -A`
5. **Write message**: Follow conventional commit format with scope
6. **Verify**: Ensure commit leaves codebase green

## Amending vs New Commits

- **Amend** when fixing typos or small issues in the last unpushed commit
- **New commit** for distinct logical changes
- **Interactive rebase** to clean up before PR (squash fixups, reorder)

## Versioning and Tags

Use **semantic versioning**: `vMAJOR.MINOR.PATCH`

- **MAJOR** - Breaking changes
- **MINOR** - New features (backwards compatible)
- **PATCH** - Bug fixes

### Creating Tags

Use annotated tags when a release is ready:
```bash
git tag -a v1.2.0 -m "v1.2.0: brief description"
git push origin v1.2.0
```

### Tag Discipline

- **Tags are immutable** - Once pushed, don't move or delete tags
- **Main branch = tagged** - Every commit on main should be part of a tagged release
- **New changes = new version** - Committing to main means bumping the version
- **Tag after commit** - Commit first, then tag, then push both

