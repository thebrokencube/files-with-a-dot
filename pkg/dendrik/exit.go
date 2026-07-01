package dendrik

//dendrik:block exit-code
//dendrik:kind token
//dendrik:layer code
//dendrik:status shipped
//dendrik:definition four exit constants: 0 ok, 1 user-error, 2 external/retry, 3 conflict
//dendrik:intent typed exit signal to route on (fix/retry/resolve), not just pass-fail
//dendrik:conformance exit-constants

const (
	ExitOK          = 0 // Success
	ExitUserError   = 1 // Bad input — agent should fix invocation
	ExitExternalErr = 2 // API down, network failure — agent should retry/report
	ExitConflict    = 3 // Resource conflict — agent should resolve
)
