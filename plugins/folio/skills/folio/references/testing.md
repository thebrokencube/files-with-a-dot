# Integration Testing

Shell-based integration tests for folio CLI behavior. Use `FOLIO_HOME` to isolate
test state from your real `~/.folio` — no backup/restore needed.

## Setup

All commands respect the `FOLIO_HOME` environment variable. Use it as an inline
prefix per command to isolate test state from `~/.folio` without touching the shell
environment (important if `FOLIO_HOME` is set in your shell config):

```bash
FOLIO_HOME=/tmp/folio-test folio home init
FOLIO_HOME=/tmp/folio-test folio init --name "test-project"
FOLIO_HOME=/tmp/folio-test folio home push -m "feat(test-project): init project"
rm -rf /tmp/folio-test
```

Do not use `export FOLIO_HOME=...` across separate commands — shell state does not
persist between invocations if `FOLIO_HOME` is already set in your shell config.

## Test: home init + project init + push

Covers the first-run flow: initializing FOLIO_HOME, creating a project, and
committing with `folio home push`.

### Steps

```bash
FOLIO_HOME=/tmp/folio-test folio home init
```

Verify:
- [ ] `/tmp/folio-test/{active,archive}` directories created
- [ ] `/tmp/folio-test/.git` exists (git repo initialized)
- [ ] `CLAUDE.md`, `README.md`, `.gitignore` present

```bash
FOLIO_HOME=/tmp/folio-test folio init --name "test-project"
```

Verify:
- [ ] `/tmp/folio-test/active/test-project/folio.yml` exists
- [ ] `folio.yml` contains `project: "test-project"` and schema version 2
- [ ] Path is printed to stdout (not just a success message)

```bash
FOLIO_HOME=/tmp/folio-test folio home push -m "feat(test-project): init project"
```

Verify:
- [ ] Exit code 0
- [ ] Commit created (not "nothing to commit")
- [ ] `FOLIO_HOME=/tmp/folio-test folio home list` shows `test-project` in active section

### Teardown

```bash
rm -rf /tmp/folio-test
```

## Adding New Test Cases

Each test case follows the same structure:

1. **Setup**: choose a unique `/tmp/folio-test-<name>` path
2. **Steps**: `FOLIO_HOME=/tmp/folio-test-<name> folio <cmd>` per command
3. **Verify**: checklist of filesystem or output assertions
4. **Teardown**: `rm -rf /tmp/folio-test-<name>`

Keep tests independent — each uses its own path.
