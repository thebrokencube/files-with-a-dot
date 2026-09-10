# jf

Implement commands as `cmd_<name>.go` with `run<Name>(args []string) int`; register them by prerequisite tier and parse flags with `parseFlags`. Preserve exit-code and structured-output contracts.

Keep external runners injectable in tests. Use dry runs for external changes and require explicit noninteractive confirmation. Run `make check && make build` after Go changes; the in-tree binary is transient.
