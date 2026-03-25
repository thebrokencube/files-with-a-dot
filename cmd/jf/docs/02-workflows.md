# Workflows

## Archetype Workflows

### (a) Ad-hoc: Single-ticket operations

**When:** Push a description to an existing ticket, search for tickets, or create one ticket from a markdown file. No forest required.

```bash
jf push ACME-123 design-doc.md
jf search "login timeout" --project ACME
```

**Graduate to Level 1** when you're managing 3+ tickets from the same directory.

### (b) Small effort: Clone and sync

**When:** An epic with a handful of children. Edit descriptions locally and sync back.

```bash
jf clone ACME-100 --dir ~/work/api-redesign
cd ~/work/api-redesign
jf status                      # check what's stale
jf sync                        # push local changes, pull remote updates
jf tree                        # inspect hierarchy
```

After clone, `~/work/api-redesign/.jf/` contains the forest. All commands run from `~/work/api-redesign/`.

**Graduate to Level 2** when you need TBD nodes or bidirectional conflict handling.

### (c) Large-scale planning: Scaffold and create

**When:** Multiple epics and stories designed locally before creating tickets in Jira.

```bash
mkdir ~/work/q3-roadmap && cd ~/work/q3-roadmap
jf init --project ACME
# Create .jf/ directory structure with README.md parents and child .md files
# Use jira: TBD and type: Epic/Story/Task in frontmatter
jf validate                    # check structure
jf create-missing --dry-run    # preview
jf create-missing              # create tickets
jf push                        # push descriptions
```

### (d) On-call / triage: Investigate and annotate

**When:** Reviewing an epic's state, checking staleness, adding investigation notes.

```bash
jf clone ACME-500 --dir ~/work/triage
cd ~/work/triage
jf status
jf tree
# Edit .jf/*.md files with investigation notes
jf sync
```

## Multi-Forest Management

Each working directory gets its own `.jf/` forest. You can manage multiple independent forests across your workspace.

**Pattern:** One working directory per effort, each with its own `.jf/`.

**Real example:** A project managing 2 forests for different initiatives:

```
~/work/
  written-culture-process/
    .jf/                       <-- RETIRE-3384 epic (written culture process)
      forest.yml
      README.md
      ...
  legacy-docs-cleanup/
    .jf/                       <-- RETIRE-3098 epic (legacy docs cleanup)
      forest.yml
      README.md
      ...
```

Each forest is independent -- sync one without affecting the other. Forests can span different Jira projects (RETIRE, BEN) via per-node `jira:` keys in frontmatter.

## Recovery

### Reset state baselines

If you see a state inconsistency warning:

```bash
jf sync --dry-run              # assess current state
```

Delete `.jf/state.json` to reset all baselines. The next sync treats all nodes as first-sync.

### Rebuild from scratch

When a forest is in an awkward state (broken baselines, corrupted frontmatter):

1. Note the root key from `.jf/forest.yml` or `.jf/README.md` frontmatter
2. Push any local content you want to preserve: `jf sync`
3. Delete the `.jf/` directory (or the whole working directory)
4. Re-clone: `jf clone <ROOT_KEY> --dir <working-dir>`

Clone pulls all descriptions fresh and establishes clean baselines. Jira is the source of truth.

## Coexisting with Other Tools

The `.jf/` dot-folder model keeps forest files isolated from your other files. A working directory can contain both a jf forest and other tooling:

```
my-effort/
  .jf/                         <-- jf forest (Jira sync)
    forest.yml
    README.md
    ...
  reference/                   <-- domain docs, research
  output/                      <-- compiled artifacts
  notes.md                     <-- your own files
```

jf only touches files inside `.jf/`. Everything else in the working directory is yours.
