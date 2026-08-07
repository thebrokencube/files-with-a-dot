# Writing — Replace Table

Loaded on demand from `review-shared.md` → Writing Quality, when a finding's fix is a word swap.

A replace list, not a ban list. Every row names what to write instead, or says delete. A negative alone
nudges the unwanted token down slightly; a stated target pulls harder. Where the row says delete, the
word carries no fact and replacing it just relocates the problem.

| Instead of | Write |
|---|---|
| leverage, utilize | use |
| facilitate | help, make possible |
| in order to | to |
| prior to | before |
| due to the fact that | because |
| in the event that | if |
| when it comes to | for |
| enables you to, allows you to | you can |
| is designed to, aims to | delete — say what it does |
| functionality | function, feature |
| streamline | make simpler, make faster |
| plethora, myriad | many |
| out of the box | by default |
| under the hood | internally |
| dive into, delve into | read, examine |
| robust, powerful, comprehensive, performant | delete, or give the measurable property |
| gracefully handles | say what it does — "retries three times, then stops" |
| blazingly fast, state-of-the-art | give the number, or delete |
| simply, just, easily, seamlessly, effortlessly | delete |
| it is worth noting that | delete |
| it's important to, crucially | delete — state the fact |
| as needed, as necessary | state the condition |
| addresses the issue, tackles | corrects the fault, removes the error |
| and/or | pick one, or write "X, or Y, or both" |
| e.g. / i.e. / etc. | for example / that is / name the items |

## Modal swaps

Distinct from the table above, because these change meaning rather than register. From the modal ladder
in `~/.claude/rules/writing-structure.md`.

| Instead of | Write |
|---|---|
| should (a requirement) | must |
| should (a recommendation) | state it as fact, or delete it |
| may, might, could (possibility) | can |
| may (permission) | can |
| would (hypothetical) | restructure — "If X occurs, Y occurs" |

## Scope

Applies to prose in agentic documents and human-facing artifacts. It does not apply to code, identifiers,
CLI flags, file paths, quoted error messages, or product names. Those are exact and stay exact.

Adapted from the SimpleEnglish skill's slop-substitution table, which is that project's own work rather
than part of ASD-STE100.
