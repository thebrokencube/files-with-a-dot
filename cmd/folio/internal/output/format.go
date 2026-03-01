package output

import "fmt"

// Exported ANSI constants for command handlers that manage their own output.
const (
	Red    = ansiRed
	Green  = ansiGreen
	Yellow = ansiYellow
	Bold   = ansiBold
	Dim    = ansiDim
	Reset  = ansiReset
)

// Errf returns a formatted error string with red "Error:" prefix.
func Errf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%sError:%s %s", Red, Reset, msg)
}

// Successf returns a formatted success string with green checkmark prefix.
func Successf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s✓%s %s", Green, Reset, msg)
}
