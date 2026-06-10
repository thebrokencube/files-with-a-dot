# Review Leaf — ARCHITECTURE.md

Type-local dimensions for an ARCHITECTURE.md (repo/system map). Applied with `review-shared.md`.
Audience is humans-first, agents on-demand — no tool auto-loads it. There is no agent-specific
ARCHITECTURE.md convention, so review for usefulness and durability, not spec compliance.

## Dimensions

### No hand-frozen derivable data (the dominant failure mode)
Facts that can be derived — module/owner/path tables, dependency lists, anything generatable
from source or config — must be **generated or pointed-to**, not hand-maintained. Hand-frozen
tables are a snapshot of a moving target; they drift silently and an agent trusting a stale
table is worse than one told to query. **fail** on a large hand-maintained table of derivable
data with no generation source; **warn** when it's small or marked as a snapshot.

### Rules front-loaded over enumeration
The high-value content (layering rules, invariants, how-to-navigate) should come first; long
enumerations belong later or behind a pointer. **warn** when rules are buried under inventory.

### States invariants
A good architecture doc states the load-bearing rules ("dependencies flow downward only",
"cross-pack refs only touch public API") explicitly, so an agent gets them without inferring.
**warn** when invariants are implied but never stated.

### One canonical grouping
Avoid two overlapping tables/views of the same things grouped differently — they disagree at
birth and drift further. **fail** when overlapping tables already contradict; **warn** on
redundant groupings that merely risk it.

### Referenced, not inlined
ARCHITECTURE.md should be *pointed at* by AGENTS.md/CLAUDE.md, not copied into them (always-on
token cost). Conversely, an orphan ARCHITECTURE.md that nothing references is easy to miss —
**warn** to link it.
