# Getting Started

## Prerequisites

- **Node.js** -- required for markdown / ADF (Atlassian Document Format) conversion
- **acli** -- Atlassian CLI (`brew install acli`)
- **Jira auth** -- run `acli auth login` to authenticate

Verify with:

```bash
jf setup
```

## The .jf/ Dot-Folder Model

jf organizes Jira tickets into tree-like structures called **forests**. A forest is a directory of `.md` files (with YAML frontmatter) that maps to a Jira hierarchy.

jf uses a `.jf/` dot-folder (analogous to `.git/`) to store all forest content -- markdown files, `forest.yml`, `state.json`. Your working directory stays clean.

**Key rules:**

- `--dir` points to the **working directory** (parent of `.jf/`), not the forest root
- jf walks up from your current directory looking for `.jf/forest.yml`
- File paths in command output are working-dir-relative: `.jf/README.md`, `.jf/PROJ-101.md`

### Single-level forest (epic + stories)

An epic with children -- the most common shape:

```
my-effort/
  .jf/
    forest.yml
    state.json
    README.md                  <-- epic (root)
    PROJ-101.md                <-- story: clean push (mutable)
    PROJ-102.md                <-- story: table content (mutable)
    PROJ-103.md                <-- story: code block (read-only, pull-only)
    PROJ-104.md                <-- story: nested list (read-only, pull-only)
    PROJ-105.md                <-- story: empty node (blocked)
```

`jf tree` output:

```
my-effort
  [jf Test] lint pipeline (PROJ-6000)
    clean push (PROJ-101)
    table content (PROJ-102)
    code block (PROJ-103)
    nested list (PROJ-104)
    empty node (PROJ-105)
```

### Multi-level forest (initiative > project > epic > stories)

Deeper hierarchies use directories. Nodes with children become subdirectories with `README.md` as the parent; leaf nodes are `<KEY>.md` files:

```
my-effort/
  .jf/
    forest.yml
    state.json
    README.md                         <-- initiative (root)
    child-project/
      README.md                       <-- project (has children)
      child-epic/
        README.md                     <-- epic (has children)
        PROJ-1003.md                  <-- leaf story A
        PROJ-1004.md                  <-- leaf story B
```

`jf tree` output:

```
my-effort
  [jf Test] multi-level (PROJ-1000)
    child project (PROJ-1001)
      child epic (PROJ-1002)
        leaf story A (PROJ-1003)
        leaf story B (PROJ-1004)
```

## Quick Start

### Clone an existing Jira hierarchy

```bash
jf clone PROJ-123 --dir ~/work/my-effort
cd ~/work/my-effort
jf tree
jf status
jf sync
```

Clone creates `~/work/my-effort/.jf/` with all the scaffolded nodes. You work from `~/work/my-effort/` -- all jf commands find `.jf/forest.yml` automatically.

### Start a greenfield forest

```bash
mkdir ~/work/my-effort && cd ~/work/my-effort
jf init --project PROJ
```

This creates `.jf/forest.yml`. Then create `.md` files inside `.jf/` with frontmatter (see [Reference](03-reference.md)):

```bash
jf push
```

### Ad-hoc (no forest)

Push or pull a single file without any forest setup:

```bash
jf push ACME-123 notes.md
jf pull ACME-456 output.md
```

## Level Progression

jf supports incremental adoption. Start at Level 0 and graduate as your needs grow.

### Level 0: Ad-hoc (no forest)

Push a single markdown file to a Jira ticket description. No `forest.yml` needed.

```bash
jf push ACME-123 notes.md
jf pull ACME-123 output.md
```

**Graduate to Level 1** when you're managing 3+ tickets from the same directory.

### Level 1: Persistent forest

Create a forest to manage multiple nodes from a single directory.

```bash
mkdir ~/work/my-effort && cd ~/work/my-effort
jf init --project ACME
# Create .jf/feature.md with frontmatter:
#   ---
#   jira: ACME-456
#   ---
jf sync
```

**Graduate to Level 2** when you need TBD nodes or hierarchies.

### Level 2: Hierarchies and TBD nodes

Plan a hierarchy before tickets exist. Use `jira: TBD` (to be determined) in frontmatter as a placeholder, then create them in Jira.

```bash
# Create .jf/README.md (parent) and .jf/child.md files with jira: TBD
jf create-missing --dry-run    # preview what would be created
jf create-missing              # create tickets, rewrite TBD -> real keys
jf push --subtree ACME-100     # push only a branch of the tree
```

**Graduate to Level 3** when others are editing the same tickets in Jira.

### Level 3: Bidirectional sync with conflict resolution

When others edit tickets in Jira, pull their changes and merge with yours. Bidirectional sync is the default -- no `sync:` field needed.

```bash
jf sync                        # detects conflicts, skips them by default
jf sync --resolve local        # local wins on conflict
jf sync --resolve remote       # remote wins on conflict
```
