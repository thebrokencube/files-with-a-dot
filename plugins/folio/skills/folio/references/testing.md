# Integration Testing

Shell-based integration tests for folio CLI behavior. Use a temporary `FOLIO_HOME` content root to isolate
test state from the real store. For container-mode tests, use a separate temporary
`FOLIO_UMBRELLA` containing `stores.yml`; never treat the umbrella as the test content root.

## Setup

Legacy isolated test:

```bash
FOLIO_HOME=/tmp/folio-test folio home init
FOLIO_HOME=/tmp/folio-test folio init --name "test-project"
FOLIO_HOME=/tmp/folio-test folio home push -m "feat(test-project): init project"
rm -rf /tmp/folio-test
```

Container-mode test:

```bash
FOLIO_UMBRELLA=/tmp/folio-control FOLIO_HOME=/tmp/folio-test folio home list
```

Use inline prefixes per command. Do not export either variable across separate
commands — shell state may already contain a user's control or content root.

## Freshness fixtures

Digest snapshot primitives belong in isolated `t.TempDir()` or `/tmp` fixtures. Test
unstamped outputs as `unknown`, changed inputs as `stale`, and missing outputs as
`missing`. Exercise `folio touch` to record a digest without changing output bytes or
modification times, and exercise `--final` only with a direct external output plus a
local review copy.

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
- [ ] `AGENTS.md`, `CLAUDE.md`, `README.md`, `.gitignore` present

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
