# Code Comments

Default to no comment. Add one short, line-local comment only when it records a surprising external behavior, a constraint a future edit could break, or why the obvious alternative is wrong. The code must not already express that fact.

Do not narrate the next line, explain a name, add banners, preserve decision history, anticipate reviewers, or move discarded comments into a PR. Tests need no explanatory comments. If an alternative is worth reporting, state it once in the task result rather than in the source.

When reviewing a change, remove comments that fail this test. A comment that is retained must identify the exact line and the fact it protects.
