# Review Workflow

Read by `/folio review [scope]`. Assumes you've already read SKILL.md for orientation and tooling resolution.

## Scope

- No arg or `local` = local checks only (cross-references, staleness, structural integrity)
- `external` = also fetch and compare external targets (normalize formats before diff)
- Specific target ID = validate just that target externally

## Steps

1. Run `folio validate`. Report errors.
2. Run `folio status`. Report all targets.
3. For each `cross_references` entry: read source_of_truth and each also_appears_in, flag differences. Descriptive references: report as not machine-checkable.
4. (External scope only) Fetch external targets via tooling.yml pull method, compare against local. For format differences (ADF vs markdown), compare structural elements not literal text.

## Output Format

```
## Status
- [target-id]: clean / stale / missing / unknown

## Cross-References
- [fact]: consistent / warning / error

## External Validation (if requested)
- [system] [id]: matches / differs
```

Review only reports. It does not fix anything.
