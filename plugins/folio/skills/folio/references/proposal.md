# Proposal Format

Default shape for any folio target that asks readers to approve a decision: proposals, tech specs, RFCs and
decision requests, whatever the destination. Compose applies it without an opt-in. The target's `how` supplies
the audience, the stakeholders, the decision requested and any deviations. If `how` names a different set of
sections, `how` wins; the writing rules, linking rules and audit still apply inside those sections.

Not for launch plans, reference docs, glossaries, status updates, Jira issues (see `/jf`), PR bodies (see the
commit skill) or folio's own plan design docs.

Load the Dendrik skill and follow the Mechanics in its writing guidance. The rules below add only what a
proposal needs.

## Shape

```
# Proposal: {what changes, in a few words}

**Authors:** {names}
**Stakeholders:** {teams or roles who must read it}
**Last Updated:** {YYYY-MM-DD}
**Status:** {Draft | In review | Approved}   ← optional

## TL;DR
## Context
## Decision
## Impact
## Timeline
## Alternatives
## References
---
## Feedback
```

If the destination renders the page title from a property, drop the H1 and let the property carry it.

## Section rules

| Section | Rule |
|---|---|
| TL;DR | One sentence, at most 35 words, naming the change and the reason. Needing a second sentence means the decision is still open. |
| Context | 3–7 bullets, one fact each: the current behavior, what it costs, who feels it. Give numbers where you have them. Skip history that changes no one's decision. |
| Decision | Bullets saying what changes, written as done ("Remove X", not "We would remove X"). Then `Key components:`, a bullet list of `**Name**: role`. End with `**Decision requested:**` and the single thing the reader approves or rejects. |
| Impact | Two labeled lists, `What improves:` and `Open risks:`. Risks are the parts not yet proven. An empty risk list means nobody looked. |
| Timeline | A numbered list of phases, each `**Phase**: outcome. Gate: {the observable condition before the next phase}`. Dates or tickets only if they exist. |
| Alternatives | One bullet per option actually weighed: `**Option.** Rejected — reason.` |
| References | Links grouped under bold labels (`Code:`, `Docs:`, `Tickets:`). Every link opens for the intended reader. |
| Feedback | Always last, after a `---` divider. |

Delete any section that would say only "None" or "N/A". TL;DR, Decision and Feedback always stay.

## Linking claims

- Link every claim about how a system behaves to its source: "the sweep runs every 6 hours", "the client
  retries only on 5xx", "the flag defaults to off".
- Sources: a code permalink pinned to a commit (never a branch, which moves), a doc, a dashboard or log query,
  a PR or an issue.
- Link every number to the query or report that produced it.
- If a claim has no linkable source, state how you verified it ("reproduced locally on 2030-01-10"), or cut it.
- Put the link on the words that make the claim. References lists the main sources; it does not replace
  inline links.

## Diagrams

- At most one.
- Only for content that branches or has states: a decision tree, a state machine, a flow that forks. Write a
  linear sequence as a numbered list.
- No before-and-after pairs and no rollout diagram; the Timeline list already shows the order.
- Put a one-line caption under it saying what the reader should notice.

## Length

- Body (TL;DR through Alternatives): 600 words or fewer. Aim for 400.
- Past 600, move background into a linked doc; do not squeeze the Decision.

## Audit

Run before the compose review gate.

1. From the TL;DR and `Decision requested` alone, can a reader new to the topic say what is being asked?
2. Does any bullet carry two ideas? Split it. Do two bullets make one point? Merge them.
3. Does any sentence hedge ("might", "probably", "should help")? Check the claim and state it, or delete it.
4. Does every behavioral claim and every number link to where a reader can check it?
5. Does each section open with its conclusion and say something no earlier section said?
6. Does any bullet restate its heading or the TL;DR? Cut it.
7. Is there at most one diagram, and does it show branching or state?
8. Is the body within the word budget? Report the count at the review gate.
9. For a shared destination: are local paths, planning ids and agent names gone, and does every link open for
   the intended reader?

## Feedback table

Stance is 👍 (approve as written) or 😬 (needs changes before approval).

Notion destination (see publish.md for rendering constraints):

```
## Feedback

Add a row: your name, 👍 or 😬, and what you need changed.

<table fit-page-width="true" header-row="true">
<tr>
<td>Reviewer</td>
<td>Stance</td>
<td>Feedback</td>
</tr>
</table>
```

Other destinations use a markdown table with the same three columns.

## Example (synthetic)

```
# Proposal: Send Invoice Emails From a Queue

**Authors:** A. Writer
**Stakeholders:** Billing engineers · Support leads
**Last Updated:** 2030-01-15

## TL;DR

Move invoice emails out of the checkout request and onto a retrying background queue, so a slow mail provider no longer fails checkouts.

## Context

- Checkout [sends the invoice email inline](https://example.com/billing/blob/3f9c2ab/checkout/complete.rb#L42), before responding.
- When the mail provider is slow, checkout [times out after 10 s](https://example.com/billing/blob/3f9c2ab/config/timeouts.yml#L8) and the customer sees an error.
- Last quarter, [1.2% of checkouts failed this way](https://example.com/dashboards/checkout-timeouts?range=2029-Q4); every failure was a paid order.
- A failed send [is never retried](https://example.com/billing/blob/3f9c2ab/mailers/invoice.rb#L17); support re-sends by hand.

## Decision

- Checkout records an `invoice_email` job and returns.
- A worker sends the email and retries with backoff for 24 h, per the [queue retry policy](https://example.com/docs/queue-retries).
- After the last retry, the job moves to a dead-letter list that support can see.

Key components:

- **Email job**: invoice id and recipient; nothing else.
- **Send worker**: sends once per job; retries only on provider errors.
- **Dead-letter view**: failed jobs, with a one-click re-send.

**Decision requested:** approve moving invoice sends out of checkout; the retry window is open to change.

## Impact

What improves:

- Checkout no longer depends on the mail provider being up.
- Support stops re-sending invoices by hand.

Open risks:

- Invoices can arrive minutes late while the queue is backed up.
- The worker has not been load-tested at peak checkout volume.

## Timeline

1. **Shadow**: the worker sends to a test inbox alongside the inline send. Gate: a week with matching counts.
2. **Switch**: checkout stops sending inline. Gate: dead-letter rate below 0.1% for a week.
3. **Clean up**: delete the inline send path.

## Alternatives

- **Raise the checkout timeout.** Rejected — it hides the failure and slows every checkout.
- **Switch mail providers.** Rejected — any provider can be slow; the coupling stays.

## References

**Code:** [checkout](https://example.com/billing/tree/3f9c2ab/checkout), [invoice mailer](https://example.com/billing/tree/3f9c2ab/mailers)
**Docs:** [Queue retry policy](https://example.com/docs/queue-retries)
**Tickets:** [BILL-123](https://example.com/issues/BILL-123)

---

## Feedback

| Reviewer | Stance | Feedback |
|---|---|---|
```
