# Workflow

- **Scope confirmation required.** Before reorganizing files, archiving work tracks, restructuring directories, or moving content between folio artifact types, confirm the scope first. Small edits within a file are fine.
- **Skill-first for domain operations.** Before performing Jira, git commit, rebase, or other domain operations, invoke the owning skill (`/jf`, `/commit`, `/stacked-pr`). Do not use raw MCP tools or CLI commands for these domains — even if tool schemas are already loaded. The skill enforces conventions (parking lot checks, commit format, field defaults) that raw tools skip.
- **Respect the invoked skill's tooling.** When a `/skill` command is invoked, use that skill's intended tools. Don't substitute with raw commands, MCP calls, or subagents unless the skill's approach explicitly fails.
- **Keep it simple.** Start with the simplest implementation that works. No extra error handling, wrapper classes, or abstraction unless asked.
- **Show, don't summarize.** When asked to see a plan, document, or file contents, display the actual content.
- **NEVER chain commands.** No `&&`, `;`, `||`, pipes, or command substitution. One operation per Bash call. This is absolute — hooks and permissions break on chained commands.
- **Use `cd` for cross-repo work.** When you need to run commands in another directory, `cd <path>` in its own Bash call first. This is required, not optional. Never use `git -C` — hooks cannot pattern-match it.
- **Always use `/commit` skill for commits.** Never assume commit message conventions. The skill defines the format, trailers policy, and workflow.
- **One-liner commits only.** Use `git commit -m "message"`. Never use heredoc, multi-line strings, or `-m` with newlines.
- **Planning**: Use `/folio plan` for non-trivial tasks. Do not call EnterPlanMode directly. Skip planning for trivial changes.
- **Observation management**: Always batch `folio observe resolve "#N" "#N2" ...` to avoid index shifting between calls.
