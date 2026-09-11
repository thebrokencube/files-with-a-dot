# Verification Procedure

Verify every commit, not only the final tip. Run the repository's CI-equivalent command, then review the complete changed artifact for stale names and inconsistent vocabulary. A rename needs case-insensitive, line-wrap-aware, and identifier-versus-prose searches.

A test must fail when its guard is removed, then exercise the guard through its real caller. Do not write a test that only proves wiring. Use supplied screenshots and user-reported failures as evidence; do not rediscover them.

Self-verify first. Request one bounded cold review only when a real context blind spot remains. A sign-off or decision gate is standing authorization for one such review; take it without asking. Give that reviewer one question, require file-and-line evidence, and confirm every finding. Delegated work must name the deliverable, its constraints, a concise return contract, and an explicit instruction to return the result; treat silence as a delivery failure, not an empty finding set.

At a gate, show the artifact rather than a recap. State only source-grounded facts and use observable, repository-native evidence.
