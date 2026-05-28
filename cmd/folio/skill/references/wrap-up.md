# Wrap-up

End-of-session workflow. Invoked via explicit `/folio wrap-up` subcommand only.

**Do NOT trigger on freeform phrases** like "I'm done" or "let's wrap up." Only the literal `/folio wrap-up` command or "wrap-up" as ARGUMENTS to `/folio` activates this workflow.

## Workflow

Each step is independently safe — the user can skip or abandon at any point without leaving broken state. Archiving without successor tracks is fine (create them next session). Handoff without archive is fine (archive next session).

1. **Scan** — identify projects touched this session.
   Primary: check mtime on work track directories for recent changes (same mechanism as Stale Detection in `references/lifecycle.md`, inverted — look for dirs modified today or within the last few hours).
   Fallback: ask the user which projects were touched.

2. **Retro gate** (soft) — present a session summary, then ask if it warrants a retro.

   **2a. Session summary.** Before asking the retro question, compile and present:
   - Projects touched (from step 1)
   - Key actions taken (commits, artifacts created, bugs fixed, decisions made)
   - Things that went sideways (failed approaches, rework, user corrections)
   - Open threads (unresolved questions, known gaps, deferred work)

   Keep it to 5-10 bullet points. The goal is shared context — the user shouldn't
   have to recall the session from memory.

   **2b. Retro question.** "Worth capturing as a retro, or just move on?"
   If yes: `folio new retro <topic>`, pre-fill from the summary, commit via `folio home push`.
   If no: proceed.

3. **Archive gate** (hard) — run lifecycle derivation on all touched projects. Tracks matching rules 5 or 6 (all done, or all done + retro) are archive candidates.
   Present all eligible tracks as a single batch across all projects:
   ```
   Ready to archive:
     - SRM / "Legacy Launch Prep" (retro complete)
     - Folio / "CLI Cleanup" (all tracks done, no retro — skipping retro)
   Archive all? (yes/no)
   ```
   One confirmation for the entire batch. On yes, run `folio archive` per track.

4. **Successor tracks** — for archived tracks where work continues, offer to scaffold a successor:
   "Continue [track-name] next session? I'll scaffold the next track."
   If yes: `folio new <type> <topic>` with `depends_on` pointing to the archived track's key artifact (retro, design doc, or spike). The `depends_on` is a manual folio.yml edit — acceptable for source declarations.

5. **Update folio state** — no separate handoff file. Persist what changed this session:
   - For each touched project, update the design doc (if one exists) with new pinned
     constraints, convergence-status changes, and open questions.
   - Log follow-on work as observations: `folio observe "type(scope): one-line"`.
   - Commit via `folio home push`.
   A new session resumes by running `/folio <project>` — folio surfaces what's current.
   Do NOT write a handoff file to `/tmp` or anywhere else.

6. **Workspace cleanup** — `folio home push` then `folio home workspace cleanup <path>`.

## Plan-Design Delegation

If the current session has an active plan-design session (design doc exists but no committed brief), wrap-up delegates to plan-design Session Exit rather than running its own competing flow:

| Wrap-up step | Delegates to |
|---|---|
| Step 1 (scan) | Wrap-up owns this |
| Step 2 (retro) | plan-design Session Exit Step 2 (Retro) |
| Step 3 (archive) | Wrap-up owns this — plan-design has no archive concept |
| Steps 4-5 (successor + state update) | plan-design Session Exit Step 3 (Update State and Commit) |
| Step 6 (cleanup) | Wrap-up owns this |

## Edge Cases

**Multi-project sessions**: Steps 1 and 3 operate across all touched projects. The archive gate (step 3) batches all eligible tracks into a single confirmation — not per-project.

**Session-touch detection**: mtime-based detection works regardless of VCS (git or jj). The 14-day stale threshold from lifecycle.md is for flagging old tracks; session-touch detection looks for *recent* changes (today/hours). These are the same mechanism with different thresholds.

**Partial completion**: Each step is designed to be safe if the session ends mid-flow. No step creates state that requires a subsequent step to clean up.
