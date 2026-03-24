# Testing

Clone-based integration testing for the jf engine. Create a Jira hierarchy via MCP, clone it locally, write test content, then validate sync behavior.

## Clone-Based Test Setup (from scratch)

Creates a test forest from absolute zero — Jira hierarchy via MCP, then `jf clone`, then validate. No pre-existing state required.

**Cloud ID**: Read from `~/.jf.yml` at `cloud_id`. Do NOT call `getAccessibleAtlassianResources`.

### Step 1: Create Jira Hierarchy

Create an epic (or initiative for multi-level) with children via MCP. Each child gets a specific content state to exercise different engine behaviors. Record every created key — you'll need them for verification in later steps.

**Single-level** (epic + children):

1. Create the epic (PROJ requires T-Shirt Size for Epics):
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Epic", summary: "[jf Test] <test-name>",
  additional_fields: {"customfield_XXXXX": {"value": "Small"}} })
```
Record the returned key (e.g., `PROJ-6000`).

2. Create children under it:
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Story", summary: "[jf Test] clean push",
  parent: "<epic-key>" })
```
Repeat for each child in the table below. Record each returned key.

3. Set descriptions on children that need remote content:
```
editJiraIssue({ cloudId: "<cloud-id>", issueIdOrKey: "<child-key>",
  contentFormat: "adf",
  fields: {
    description: { type: "doc", version: 1, content: [{
      type: "paragraph", content: [{ type: "text", text: "Test content for clean push." }]
    }] }
  }
})
```

4. Verify tickets exist before cloning:
```
getJiraIssue({ cloudId: "<cloud-id>", issueIdOrKey: "<epic-key>" })
```
Confirm the epic has the expected number of children.

| Child | Summary | Remote Description | Purpose |
|-------|---------|-------------------|---------|
| clean | `[jf Test] clean push` | substantive paragraph | Mutable — push+pull |
| table | `[jf Test] table content` | substantive paragraph | Mutable — push+pull |
| codeblock | `[jf Test] code block` | empty | Write local code block after clone — read-only, demotes to pull-only |
| nested-list | `[jf Test] nested list` | empty | Write local nested list after clone — read-only, demotes to pull-only |
| empty | `[jf Test] empty node` | empty | Empty — blocked by Tier 3 |

**Multi-level** (initiative → project name → epic → stories):

Tests hierarchy depth beyond one level. Clone scaffolds nested directories: nodes with children become subdirectories with `README.md` as the parent; leaf nodes become `<KEY>.md` files.

**Jira hierarchy constraint**: Parent-child relationships require adjacent hierarchy levels (Initiative 3 → Project Name 2 → Epic 1 → Story 0). You cannot skip levels (e.g., Initiative → Epic fails with "does not belong to appropriate hierarchy").

**Project-specific required fields**: Epics in PROJ require `customfield_XXXXX` (T-Shirt Size Estimate). Pass via `additional_fields: {"customfield_XXXXX": {"value": "Small"}}`. Check required fields with `getJiraIssueTypeMetaWithFields` before creating.

1. Create the initiative (root):
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Initiative", summary: "[jf Test] multi-level" })
```
Record key (e.g., `PROJ-1000`).

2. Create a project name under the initiative:
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Project Name", summary: "[jf Test] child project",
  parent: "<initiative-key>" })
```
Record key (e.g., `PROJ-1001`).

3. Create an epic under the project name:
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Epic", summary: "[jf Test] child epic",
  parent: "<project-key>",
  additional_fields: {"customfield_XXXXX": {"value": "Small"}} })
```
Record key (e.g., `PROJ-1002`).

4. Create stories under the epic:
```
createJiraIssue({ cloudId: "<cloud-id>", projectKey: "PROJ",
  issueTypeName: "Story", summary: "[jf Test] leaf story A",
  parent: "<epic-key>" })
```
Repeat for a second leaf story. Set descriptions on at least one using `editJiraIssue` with `contentFormat: "adf"` (same pattern as single-level, but you must specify `contentFormat: "adf"` when passing raw ADF in the description field).

5. Verify hierarchy before cloning:
```
getJiraIssue({ cloudId: "<cloud-id>", issueIdOrKey: "<initiative-key>" })
```

Clone:
```bash
jf clone <INITIATIVE_KEY> --dir ~/.jf/test/<test-name>
```

Expected directory structure (`slugify` strips the `[jf Test]` bracket prefix, then lowercases and hyphenates):
```
~/.jf/test/<test-name>/multi-level/
├── forest.yml
├── README.md                          ← initiative (root)
├── child-project/
│   ├── README.md                      ← project name (has children → directory)
│   └── child-epic/
│       ├── README.md                  ← epic (has children → directory)
│       ├── PROJ-1003.md             ← leaf story A
│       └── PROJ-1004.md             ← leaf story B
```

Verify:
- [ ] Initiative is root `README.md` with correct frontmatter
- [ ] Project name and epic with children became subdirectories (not flat `.md` files)
- [ ] Leaf stories are `<KEY>.md` files inside the epic's directory
- [ ] `jf tree` shows the full hierarchy (4 levels)
- [ ] `jf status` reports correct node count (5)
- [ ] No `sync:` in any frontmatter

**Depth limiting**: `--depth 1` clones only the initiative and its direct children (the project name), stopping before epics and stories:
```bash
jf clone <INITIATIVE_KEY> --dir ~/.jf/test/<test-name>-shallow --depth 1
```
- [ ] Only 2 nodes scaffolded (initiative + project name)
- [ ] Project name is a leaf `<KEY>.md` file (no children fetched → no subdirectory)

### Step 2: Clone

`jf clone` creates a subdirectory named by slugifying the root ticket's summary. Slugify strips any leading `[bracket prefix]`, lowercases, and replaces non-alphanumeric runs with hyphens (e.g., `[jf Test] derived-sync` → `derived-sync/`). The `--dir` flag sets the parent directory where this subdirectory is created.

```bash
jf clone <EPIC_KEY> --dir ~/.jf/test/<test-name>
```

The forest root (containing `forest.yml`) is at `~/.jf/test/<test-name>/<slugified-summary>/`. Use this path for all subsequent `--dir` flags (`status`, `sync`, etc. expect the forest root, not the parent).

Verify scaffolded output:
- [ ] No `sync:` in any frontmatter (check each .md file's YAML block)
- [ ] No `defaults.sync` in forest.yml
- [ ] Descriptions pulled for children with remote content (clean, table children have body text)
- [ ] Children with empty descriptions have frontmatter only (codeblock, nested-list, empty)

With explicit override (for testing override path):
```bash
jf clone <EPIC_KEY> --dir ~/.jf/test/<test-name>-pull --sync pull
```
- [ ] `sync: pull` appears in all frontmatter and forest.yml

### Step 3: Write Test Content

After clone, write local content to nodes that need lint-failing content. These will be
detected as read-only and demoted to pull-only by the engine.

**Code block node** — find the .md file for the "codeblock" child (filename is `<KEY>.md` inside the cloned directory). Append below the frontmatter:

```
## Code Example

` ` `go
func main() { fmt.Println("hello") }
` ` `
```

(The code fence above uses spaced backticks for display — write actual triple backticks when editing the file.)

**Nested list node** — find the .md file for the "nested-list" child. Append below the frontmatter:

```
## Structure

- Item one
  - Sub-item A
  - Sub-item B
- Item two
```

Leave the "empty" node untouched (no content below frontmatter).

### Step 4: Validate

Use the forest root path from Step 2:

```bash
jf status --dir ~/.jf/test/<test-name>/<slugified-summary>
```
Expected output pattern:
```
Effective direction:
  N push+pull (mutable)     # clean content nodes
  N pull-only (N read-only demoted)  # code block, nested list nodes
  N empty                   # empty nodes
```

```bash
jf sync --dry-run --dir ~/.jf/test/<test-name>/<slugified-summary>
```
Verify:
- Mutable nodes with local changes → PUSH
- Read-only nodes → demoted, pull-only behavior
- Empty nodes → BLOCKED(empty)
- Explicit `sync:pull` nodes (if any) → never push

If push fails for a mutable node, check that the content uses only supported markdown (see USAGE.md "Lint and Mutability"). Rewrite using supported constructs and retry.

### Step 5: Teardown

```bash
rm -rf ~/.jf/test/<test-name>
```
Jira tickets can be parked via the `/jf` skill's park workflow (see lifecycle.md) — this is a skill-level operation, not a CLI command.

### Rebuild From Scratch

When a forest gets into an awkward state (broken baselines, stale state.json, corrupted frontmatter), the fastest recovery is:

1. Note the root key: check `forest.yml` or the root README.md frontmatter
2. Delete the local forest directory
3. Re-clone: `jf clone <ROOT_KEY> --dir <parent-dir>`
4. Clone pulls all descriptions fresh and establishes clean baselines

This works because Jira is the source of truth for remote content, and clone re-scaffolds everything from the hierarchy. Local-only content that hasn't been pushed is lost — push first if needed.
