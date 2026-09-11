# Global Policy Migration Ledger

This is an audit ledger, not an installation registry. Track 2 consumes only the retired-projection table.

## Statement disposition

| Former source | Imperative or rule | Canonical destination |
| --- | --- | --- |
| `CLAUDE.md` | do not use file memory | `agents/AGENTS.md` — bullet 1 |
| `CLAUDE.md` | do not narrate code comments | `agents/AGENTS.md` — bullet 2 |
| `CLAUDE.md` | edit generated-file sources; sync after commit | `agents/AGENTS.md` — bullet 3; `skills/dotfiles/SKILL.md` — Use the existing mechanisms |
| `CLAUDE.md` | use PATH-resolved CLI tools | `skills/dotfiles/SKILL.md` — Use the existing mechanisms |
| `CLAUDE.md` | no commit trailers; conventional, cohesive commits; stack procedure | `agents/AGENTS.md` — bullet 9; `skills/commit/SKILL.md` — Commit Conventions |
| `rules/no-memory-files.md` | never mutate, offer, or cite harness memory | `agents/AGENTS.md` — bullet 1 |
| `rules/no-memory-files.md` | verify recalled claims; persist guidance by scope | `agents/AGENTS.md` — bullet 1; `plugins/folio/skills/folio/SKILL.md` — Lifecycle Model |
| `rules/code-comments.md` | one factual, line-local comment at most; no narration, banners, context, or reviewer debate | `agents/AGENTS.md` — bullet 2; `plugins/dendrik/skills/dendrik/references/code-comments.md` — Code Comments |
| `rules/code-comments.md` | use chat only for discarded alternatives; tests need no comments | `plugins/dendrik/skills/dendrik/references/code-comments.md` — Code Comments |
| `rules/dotfiles-awareness.md` | commit dotfile changes separately; dry-run and validate | `skills/dotfiles/SKILL.md` — Post-Edit Workflow |
| `rules/dotfiles-awareness.md` | build/release CLI artifacts through their repository workflow | `skills/dotfiles/SKILL.md` — Build Artifacts and Releasing a CLI tool |
| `rules/isolated-checkouts.md` | work in isolated checkouts, never beside or inside the source checkout | `plugins/folio/skills/folio/SKILL.md` — Isolation and Cleanup |
| `rules/isolated-checkouts.md` | fetch before branching; protect other workspaces; target foreign repos without entering them | `plugins/folio/skills/folio/SKILL.md` — Isolation and Cleanup |
| `rules/isolated-checkouts.md` | clean registrations rather than deleting directories; use the dotfiles exception workflow | `plugins/folio/skills/folio/SKILL.md` — Isolation and Cleanup |
| `references/isolated-checkouts-runbook.md` | list/reap work areas; clean only own KB workspace; use the manual dotfiles workspace procedure | `plugins/folio/skills/folio/SKILL.md` — Isolation and Cleanup |
| `rules/working-discipline.md` | ask one claim-first question; keep scope reviewable; split growth; edit only flagged content; weigh intent and leverage, not usage count | `agents/AGENTS.md` — bullet 5 |
| `rules/working-discipline.md` | correct root cause; inspect siblings; accept supplied screenshots | `agents/AGENTS.md` — bullets 5 and 6 |
| `rules/working-discipline.md` | validate each commit; mutation-test guards; match CI; complete rename and cohesion sweeps | `plugins/dendrik/skills/dendrik/references/verification.md` — Verification Procedure |
| `rules/working-discipline.md` | self-verify; take one cold read at a sign-off gate without asking; require subagents to return a bounded result | `plugins/dendrik/skills/dendrik/references/verification.md` — Verification Procedure |
| `rules/working-discipline.md` | show artifacts at gates; ground claims; use observable access; answer confirmations directly | `agents/AGENTS.md` — bullets 7 and 8 |
| `rules/workflow.md` | confirm reorganization scope; use the owning skill; use simple implementation | `agents/AGENTS.md` — bullets 4 and 5 |
| `rules/workflow.md` | use jf configuration and forest preflight; use jf before Jira MCP | `plugins/jf/skills/jf/SKILL.md` — Preflight Checklist and When to Use MCP Directly |
| `rules/workflow.md` | use isolated work areas; keep cwd/repository targeting explicit | `plugins/folio/skills/folio/SKILL.md` — Isolation and Cleanup |
| `rules/workflow.md` | use commit conventions and one-line commit messages | `agents/AGENTS.md` — bullet 9; `skills/commit/SKILL.md` — Commit Conventions |
| `rules/workflow.md` | plan multi-file work with Folio; create source-declared artifacts; batch observations | `plugins/folio/skills/folio/SKILL.md` — /folio plan, /folio observe, and Source Declaration Checklist |
| `rules/workflow.md` | use the documentation workflow for docs; avoid internal planning references in shared artifacts | `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Shared Artifacts |
| `rules/workflow.md` | review agentic documents before commit | `plugins/dendrik/skills/dendrik/SKILL.md` — Lens: Review |
| `rules/writing-structure.md` | structure human-facing artifacts as discrete ideas; state deploy impact; define terms; collapse detail | `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Mechanics |
| `rules/writing-structure.md` | link identifiers with title; use direct modal, condition-first, outcome-first language | `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Mechanics |
| `rules/writing-structure.md` | choose artifact shape for sketches, leadership, user-facing documents, Jira, and peer review | `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Audience |
| `rules/writing-structure.md` | include writing mechanics in subagent prompts | `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Delegation |
| `output-styles/Concise.md` | answer first; keep replies, narration, and files proportionate; remove praise, recaps, filler, and meta commentary | `agents/AGENTS.md` — bullet 8 |
| `output-styles/Concise.md` | use one thought per line; include caveats only when actionable; change the artifact rather than defend taste | `agents/AGENTS.md` — bullet 8; `plugins/dendrik/skills/dendrik/references/writing-guidance.md` — Mechanics |

## Moved-tree extraction ledger

| Old path | New path | Type/mode | SHA-256 before move | Authorized transformation | Verification |
| --- | --- | --- | --- | --- | --- |
| `configs/base/claude/.claude/skills/commit/SKILL.md` | `skills/commit/SKILL.md` | regular/0644 | `30f6c31893bf169c930a903fe4d61e01f233cd8812234fc8eef3588b0ef8aed7` | remove `user_invocable: false`; replace the host-specific default reference with “host defaults” | regular/0644; SHA-256 `84b9c99f761f5626af87ce0d31d42228cda1f8ebae21a56bc1e8dde470247a83`; authorized-only diff verified |
| `configs/base/claude/.claude/skills/commit/references/git-stacking.md` | `skills/commit/references/git-stacking.md` | regular/0644 | `3b97b6e45164f3cae98d62c8ae1d31ab01e29de93ea16db25ed94828e088fa2b` | byte-for-byte move | regular/0644; matching SHA-256 verified |
| `configs/base/claude/.claude/skills/commit/references/jj-stacking.md` | `skills/commit/references/jj-stacking.md` | regular/0644 | `6d5efa1125dc0b423d0f50ec7be3b5ea4e0e941b0fe503b1ddefa1fd12d34cf6` | byte-for-byte move | regular/0644; matching SHA-256 verified |
| `configs/base/claude/.claude/skills/commit/references/pr-descriptions.md` | `skills/commit/references/pr-descriptions.md` | regular/0644 | `01dc36eb66c2b1218856da0af673c3e0fc070b9d3f3e6f3e73cfef215a50be63` | byte-for-byte move | regular/0644; matching SHA-256 verified |

## Retired projections for Track 2

| Destination | Expected raw source target | Disposition |
| --- | --- | --- |
| `$HOME/.claude/CLAUDE.md` | `configs/base/claude/.claude/CLAUDE.md` | replace with the core policy projection |
| `$HOME/.claude/rules/code-comments.md` | `configs/base/claude/.claude/rules/code-comments.md` | remove source and map row |
| `$HOME/.claude/rules/dotfiles-awareness.md` | `configs/base/claude/.claude/rules/dotfiles-awareness.md` | remove source and map row |
| `$HOME/.claude/rules/isolated-checkouts.md` | `configs/base/claude/.claude/rules/isolated-checkouts.md` | remove source and map row |
| `$HOME/.claude/rules/no-memory-files.md` | `configs/base/claude/.claude/rules/no-memory-files.md` | remove source and map row |
| `$HOME/.claude/rules/working-discipline.md` | `configs/base/claude/.claude/rules/working-discipline.md` | remove source and map row |
| `$HOME/.claude/rules/workflow.md` | `configs/base/claude/.claude/rules/workflow.md` | remove source and map row |
| `$HOME/.claude/rules/writing-structure.md` | `configs/base/claude/.claude/rules/writing-structure.md` | remove source and map row |
| `$HOME/.claude/rules/concision.md` | `configs/base/claude/.claude/output-styles/Concise.md` | remove only the rule projection; retain the native output-style projection |
| `$HOME/.claude/references/isolated-checkouts-runbook.md` | `configs/base/claude/.claude/references/isolated-checkouts-runbook.md` | remove source and map row |
