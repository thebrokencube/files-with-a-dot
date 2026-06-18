# Review Leaf — CLAUDE.md

Type-local dimensions for a CLAUDE.md. Applied with `review-shared.md`.

## First: branch on scope (by path)

- **User-global** (`~/.claude/CLAUDE.md`, any `.claude/CLAUDE.md` under a user home) and
  **enterprise/managed-policy** CLAUDE.md → **EXEMPT** from the AGENTS.md-push below. Personal
  cross-project memory and org policy are legitimately Claude-specific.
- **Project / repo** (`./CLAUDE.md`, a repo's `.claude/CLAUDE.md`) → subject to the AGENTS.md-push.
- **Ambiguous** — a dotfiles repo whose root IS a user home `.claude/` (e.g. `~/.dotfiles`
  symlinked into `~/.claude/`, so a CLAUDE.md is both project and user-global), or you cannot
  observe whether it's enterprise/managed → **state the inferred scope and ask** (per
  shared "never guess ambiguously"); do not apply the push by default.

## Dimensions

### Right Layer? (the dominant finding, project-scoped)
Content that is NOT genuinely Claude-specific — project conventions, build/test commands,
code-org rules — should move to **AGENTS.md**, leaving a thin `@AGENTS.md` pointer. Phrase as
dendrik's opinion, not a spec violation. **Max severity = warn** — never fail a project CLAUDE.md
solely for hoarding portable content.

### Always-on budget
CLAUDE.md is loaded into context every session. Every line costs per session — is each line
worth its standing cost? **warn** on reference knowledge or rarely-needed detail that belongs in
an on-demand file.

### No volatile specifics
As an always-read doc, CLAUDE.md carries direction/architecture, not volatile specifics (exact
counts, per-layer numbers, enumerated lists) that a code file or single reference already owns.
**warn** on a volatile specific; point at the one derived source (see shared "Volatile specifics
/ denormalization").

### No duplication of AGENTS.md
A CLAUDE.md that copies AGENTS.md content (instead of pointing at it) is drift waiting to
happen. **warn** — recommend a pointer.

### Correct scope tier
Content matches its tier (enterprise = org policy; project = repo conventions; user = personal
cross-project). Flag project-specific rules in a user-global file and vice versa.

### Standing instructions, not reference knowledge
CLAUDE.md holds standing behavioral instructions, not lookup material (schemas, long tables,
deep docs) — those go in references the agent loads on demand.
