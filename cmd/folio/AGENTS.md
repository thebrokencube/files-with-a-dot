# folio

Use `dendrik.NewFlagSet`, `dendrik.Parse`, and `dendrik.WriteResult` for commands. Mutating commands support `--dry-run`; Folio home mutations use the home commands, not raw VCS calls.

Keep lifecycle validation in internal packages and test filesystem behavior with temporary homes. Run `make check && make build` after Go changes; the in-tree binary is transient.
