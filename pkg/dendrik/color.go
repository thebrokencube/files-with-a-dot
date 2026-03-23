package dendrik

import "fmt"

// Palette holds ANSI escape sequences, or empty strings when color is disabled.
type Palette struct {
	Red    string
	Green  string
	Yellow string
	Bold   string
	Dim    string
	Reset  string
}

// NewPalette returns a Palette with ANSI codes if color is true, empty strings otherwise.
func NewPalette(color bool) Palette {
	if !color {
		return Palette{}
	}
	return Palette{
		Red:    "\033[0;31m",
		Green:  "\033[0;32m",
		Yellow: "\033[0;33m",
		Bold:   "\033[1m",
		Dim:    "\033[2m",
		Reset:  "\033[0m",
	}
}

// Errf returns a formatted error string with red "Error:" prefix.
func (p Palette) Errf(format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%sError:%s %s", p.Red, p.Reset, msg)
}

// Successf returns a formatted success string with green checkmark prefix.
func (p Palette) Successf(format string, args ...any) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s✓%s %s", p.Green, p.Reset, msg)
}
