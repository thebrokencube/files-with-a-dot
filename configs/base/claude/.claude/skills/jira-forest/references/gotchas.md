# Jira Gotchas

Common pitfalls when working with Jira via jf/acli.

## Content

- **Relative links**: Jira ADF doesn't support relative links. Use absolute URLs.
- **@mentions**: `@username` in markdown becomes literal text, not a Jira mention. Use `[~accountId:...]` for real mentions.
- **Description size**: Jira descriptions have a ~32KB ADF limit. Large documents may silently truncate.
- **Images**: Inline images in markdown won't render in Jira. Upload attachments separately and reference them.
- **HTML in markdown**: Raw HTML passes through marklassian but Jira strips most HTML tags.

## acli

- **`acli create` with description**: `acli create --from-json` silently drops or malforms inline ADF descriptions. Always create barebones first, then push description separately (this is what `jf create-missing` does).
- **Rate limiting**: Jira Cloud rate limits at ~100 requests/minute. For large forests, expect throttling during bulk sync.
- **Auth token scope**: `JIRA_API_TOKEN` must have write access to the target project. Read-only tokens fail silently on push.

## Frontmatter

- **`jira: TBD`** is case-insensitive — `tbd`, `"TBD"`, `'TBD'` all work.
- **Missing closing `---`**: If frontmatter has no closing fence, jf treats the file as having no frontmatter (content passes through as-is).
- **Order field**: `order: N` controls sibling sort within a directory. Lower values sort first. Nodes without order sort after ordered nodes, alphabetically by filename.

## forest.yml

- **Schema version**: Must be `schema: 1`. Missing or wrong version causes discovery to fail.
- **Defaults cascade**: `defaults.sync`, `defaults.type`, `defaults.project` apply to all nodes unless overridden in frontmatter.
- **`.jf/` directory**: Created automatically for state tracking. Add `.jf/` to `.gitignore` if the forest lives in a git repo.
