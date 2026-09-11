# dendrik

Keep filesystem access in gatherers and checks pure over `lint.ToolData`. Register every new check in its layer and `pkg/dendrik/conventions/contract.go`; update the matching tests and regenerate building-block outputs.

Keep publishable plugin bundles closed to delivery files. Run `make check && make build` after Go changes; the in-tree binary is transient.
