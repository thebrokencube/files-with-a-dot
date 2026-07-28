# Writing Structure

Applies to every artifact written for a human reader: PR titles and bodies, docs, tickets, design docs,
review output, Slack drafts.

- **A joiner is the tell.** An em-dash, comma, semicolon, or middot holding two ideas together means a list
  got written as prose. Two ideas is enough to bullet it. A closed enumeration of like items on one line is
  not a joiner: two *ideas* is the trigger, not two words.
- **Outcome first.** Lead with what is true after the change. Mechanics come second.
- **Say what changes for a human.** When the artifact describes a change that ships, answer whether anything
  changes for a person on deploy. Say it even when the answer is nothing.
- **Define the domain nouns.** Define each one before its second use. A term used more than twice and never
  defined is a defect, not a style preference.
- **Collapse the weeds rather than deleting them.** Per-file walkthroughs, enumerated states, and migration
  steps go inside a `<details>` block.

For PR titles and bodies the full house style lives in the commit skill, at
`~/.claude/skills/commit/references/pr-descriptions.md`. Read it before writing or editing either.

This is the canonical statement. It lives here rather than in the Concise output style because output styles
do not load into subagents, and rules files do.
