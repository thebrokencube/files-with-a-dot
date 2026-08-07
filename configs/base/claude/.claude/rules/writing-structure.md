# Writing Structure

Applies to every artifact written for a human reader: PR titles and bodies, docs, tickets, design docs,
review output, Slack drafts.

- **A joiner is the tell.** An em-dash, comma, semicolon, or middot holding two ideas together means a list
  got written as prose. Two ideas is enough to bullet it. A closed enumeration of like items on one line is
  not a joiner: two *ideas* is the trigger, not two words.
- **Say what changes for a human.** When the artifact describes a change that ships, answer whether anything
  changes for a person on deploy. Say it even when the answer is nothing.
- **Define the domain nouns.** Define each one before its second use. A term used more than twice and never
  defined is a defect, not a style preference.
- **Collapse the weeds rather than deleting them.** Per-file walkthroughs, enumerated states, and migration
  steps go inside a `<details>` block.

- **Link every identifier, never cite it bare.** A ticket key, PR number or build id gets a link and its
  title on first mention: `[RETIRE-16417](…/browse/RETIRE-16417) — [GLOPPY] Roll out the AI reviewer`. A bare
  five-digit key is indistinguishable from an invented one, so it reads as a hallucination and derails the
  task into proving provenance. The Jira base URL comes from `site:` in `~/.jf.yml`.

## Mechanics

Quote the bullets below verbatim into a subagent prompt. The built-in `Explore` and `Plan` agent types
never load this file, so the prompt is the only surface that reaches them.

- **Use must, can, will.** Never write *should*. A model reads it as optional.
- **Condition first, then the command.** "If the build fails, read the log." Never the reverse.
- **One noun per concept.** Do not rotate synonyms once a term is introduced.
- **Outcome first.** Lead with what is true after the change. Mechanics come second.
- **Do not fuse an instruction with its rationale.** A lens for new writing, not a retrofit.

Absorbed from ASD-STE100 Simplified Technical English, cited as provenance. Never cite one of its rule
numbers. A model without the catalog invents them.

## Match the artifact to its audience

| Artifact | Shape |
|---|---|
| Sketch, explainer, option comparison | **Diagram-first.** A matrix, flowchart, timeline or inline SVG carries each idea; text is labels and one-line captions. Prose cards and bullet lists are the failure mode. Pick the strongest visual the content admits — for a comparison, the differing cell *is* the story. |
| Leadership or non-engineering reader | Goal, three bullets, a link out to the epic. Add only when asked. Not tables of tickets, not per-stream diagrams. |
| User-facing (onboarding, UX flow, a doc to dip into) | The newcomer's plain-language journey, not the maintainer's seat. Leading with file names and CLI flags hides the idea and offloads decisions that are not the reader's. |
| Jira ticket | Brief Context / Goal / Scope, as if briefing a teammate. Attach references rather than inlining them, and invent nothing the request did not state. |
| Peer review of a person | Draft notes, not prose. The voice wanted is grounded and willing to admit uncertainty, and a polished draft reads as inauthentic — leave the final wording to the user. |

For PR titles and bodies the full house style lives in the commit skill, at
`~/.claude/skills/commit/references/pr-descriptions.md`. Read it before writing or editing either.

This is the canonical statement. It lives in a rules file because every subagent except the built-in
`Explore` and `Plan` types loads the rules hierarchy, and those two are what the Mechanics block covers.
