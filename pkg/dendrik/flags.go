package dendrik

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

// ErrHelp is returned by Parse when --help or -h is passed.
// Callers should exit with ExitOK rather than treating it as an error.
var ErrHelp = errors.New("help requested")

// FlagSet wraps pflag.FlagSet with dendrik conventions.
// Short names use rune (matching the old ff v4 API) and are converted
// to pflag's string shorthand internally.
type FlagSet struct {
	inner *pflag.FlagSet
}

// NewFlagSet creates a FlagSet with ContinueOnError handling.
func NewFlagSet(name string) *FlagSet {
	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	return &FlagSet{inner: fs}
}

// Parse parses args against the flag set. pflag supports interspersed
// flags natively, so flags can appear before or after positional args.
// Returns ErrHelp when --help/-h is passed (pflag prints usage to stderr).
func Parse(fs *FlagSet, args []string) error {
	err := fs.inner.Parse(args)
	if errors.Is(err, pflag.ErrHelp) {
		return ErrHelp
	}
	return err
}

// ParseCheck handles the common parse-then-check pattern. Returns (true, code)
// when the caller should return immediately (help requested or parse error),
// or (false, 0) when parsing succeeded and the caller should continue.
// On parse error, the error is printed to stderr.
func ParseCheck(fs *FlagSet, args []string) (bool, int) {
	err := fs.inner.Parse(args)
	if errors.Is(err, pflag.ErrHelp) {
		return true, ExitOK
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return true, ExitUserError
	}
	return false, 0
}

// GetArgs returns positional arguments remaining after parsing.
func (fs *FlagSet) GetArgs() []string {
	return fs.inner.Args()
}

// IsHelpArg reports whether s is a help request (--help, -h, or help).
// Use it in manual subcommand dispatch so help is recognized uniformly,
// the same way FlagSet-based commands get it for free via pflag.
func IsHelpArg(s string) bool {
	return s == "--help" || s == "-h" || s == "help"
}

// NoExtraArgs reports a user error if any positional args remain after
// parsing. Flag-less subcommands use it so stray arguments are rejected
// instead of silently ignored. Returns (true, code) when the caller should
// return immediately.
func NoExtraArgs(fs *FlagSet) (bool, int) {
	rest := fs.GetArgs()
	if len(rest) > 0 {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", rest[0])
		return true, ExitUserError
	}
	return false, 0
}

// Bool defines a boolean flag with short and long names.
func (fs *FlagSet) Bool(short rune, long string, usage string) *bool {
	return fs.inner.BoolP(long, string(short), false, usage)
}

// BoolLong defines a boolean flag with only a long name.
func (fs *FlagSet) BoolLong(long string, usage string) *bool {
	return fs.inner.Bool(long, false, usage)
}

// String defines a string flag with short and long names.
func (fs *FlagSet) String(short rune, long string, def string, usage string) *string {
	return fs.inner.StringP(long, string(short), def, usage)
}

// StringLong defines a string flag with only a long name.
func (fs *FlagSet) StringLong(long string, def string, usage string) *string {
	return fs.inner.String(long, def, usage)
}

// Int defines an int flag with short and long names.
func (fs *FlagSet) Int(short rune, long string, def int, usage string) *int {
	return fs.inner.IntP(long, string(short), def, usage)
}

// IntLong defines an int flag with only a long name.
func (fs *FlagSet) IntLong(long string, def int, usage string) *int {
	return fs.inner.Int(long, def, usage)
}
