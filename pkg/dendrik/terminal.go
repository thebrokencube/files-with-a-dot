package dendrik

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal returns true if fd is a terminal.
func IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

// ColorEnabled returns whether color output should be used.
// Priority: --no-color flag > NO_COLOR env > terminal detection.
func ColorEnabled(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return IsTerminal(int(os.Stdout.Fd()))
}

// OutputMode returns output format based on flags and terminal state.
// Non-TTY defaults to JSON so agents get structured output automatically.
func OutputMode(jsonFlag, plainFlag bool) string {
	if jsonFlag {
		return "json"
	}
	if plainFlag {
		return "plain"
	}
	if !IsTerminal(int(os.Stdout.Fd())) {
		return "json"
	}
	return "human"
}
