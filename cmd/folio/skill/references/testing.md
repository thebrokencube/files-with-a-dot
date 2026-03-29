# Integration Testing

Shell-based integration tests for folio CLI behavior. Use `FOLIO_HOME` to isolate
test state from your real `~/.folio` — no backup/restore needed.

## Setup

All commands respect the `FOLIO_HOME` environment variable. Point it at a temp
directory before running any test loop, and unset it when done.

```bash
export FOLIO_HOME=/tmp/folio-test
# ... run test steps ...
rm -rf /tmp/folio-test
unset FOLIO_HOME
```

## Test: home init + project init + push

Covers the first-run flow: initializing FOLIO_HOME, creating a project, and
committing with `folio home push`.

### Steps

```bash
export FOLIO_HOME=/tmp/folio-test

folio home init
```

Verify:
- [ ] `/tmp/folio-test/{active,archive}` directories created
- [ ] `/tmp/folio-test/.git` exists (git repo initialized)
- [ ] `CLAUDE.md`, `README.md`, `.gitignore` present

```bash
folio init --name "test-project"
```

Verify:
- [ ] `/tmp/folio-test/active/test-project/folio.yml` exists
- [ ] `folio.yml` contains `project: "test-project"` and schema version 2
- [ ] Path is printed to stdout (not just a success message)

```bash
folio home push -m "feat(test-project): init project"
```

Verify:
- [ ] Exit code 0
- [ ] Commit created (not "nothing to commit")
- [ ] `folio home list` shows `test-project` in active section

### Teardown

```bash
rm -rf /tmp/folio-test
unset FOLIO_HOME
```

## Adding New Test Cases

Each test case follows the same structure:

1. **Setup**: `export FOLIO_HOME=/tmp/folio-test-<name>`
2. **Steps**: CLI commands with expected output or exit code
3. **Verify**: Checklist of filesystem or output assertions
4. **Teardown**: `rm -rf /tmp/folio-test-<name> && unset FOLIO_HOME`

Keep tests independent — each uses its own `FOLIO_HOME` path.
