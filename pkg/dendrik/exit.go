package dendrik

const (
	ExitOK          = 0 // Success
	ExitUserError   = 1 // Bad input — agent should fix invocation
	ExitExternalErr = 2 // API down, network failure — agent should retry/report
	ExitConflict    = 3 // Resource conflict — agent should resolve
)
