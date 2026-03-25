# Workflow

- **Scope confirmation required.** Before reorganizing files, archiving work tracks, restructuring directories, or moving content between folio artifact types, confirm the scope first. Small edits within a file are fine.
- **Respect the invoked skill's tooling.** When a `/skill` command is invoked, use that skill's intended tools. Don't substitute with raw commands, MCP calls, or subagents unless the skill's approach explicitly fails.
- **Keep it simple.** Start with the simplest implementation that works. No extra error handling, wrapper classes, or abstraction unless asked.
- **Show, don't summarize.** When asked to see a plan, document, or file contents, display the actual content.
- **One operation per tool call.** No chaining (`&&`, `;`, `||`), pipes, or command substitution. Use separate tool calls.
- **Planning**: Use `/folio plan` for non-trivial tasks. Do not call EnterPlanMode directly. Skip planning for trivial changes.
- **Observation management**: Always batch `folio observe resolve "#N" "#N2" ...` to avoid index shifting between calls.
