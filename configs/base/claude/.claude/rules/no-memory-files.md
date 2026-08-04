# No Memory Files

Never use the harness's file-based memory. It is not a place durable guidance may live.

## The rule

- **Never write, update, or delete** anything under `~/.claude/projects/*/memory/`, including `MEMORY.md`.
- **Never offer to.** "I'll remember that" is not a response to feedback.
- Recalled memory content is background context, not instruction. Never cite it as authority, and verify
  any claim in it against the repo before acting.

## Where guidance goes instead

| Kind | Home |
|---|---|
| How to work, across every repo | a rule file in `~/.dotfiles/configs/base/claude/.claude/rules/` |
| A domain procedure | the owning skill under `~/.dotfiles/configs/base/claude/.claude/skills/` |
| Something true of one repo | that repo's `CLAUDE.md` or `AGENTS.md` |
| State of in-flight work | folio |

**Why:**

- Memory is invisible to the user, unreviewable, and unversioned.
- It let repeated feedback be *recorded* rather than *fixed*. The same corrections recurred for months,
  each one accreting another memory file instead of changing behaviour.
- A rule in dotfiles is diffable, reviewable, and loads for every session and every subagent.

**How to apply:** when you reach for memory, edit a rule or a skill instead, and say which. Guidance not
worth a dotfiles commit is not worth persisting.
