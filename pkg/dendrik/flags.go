package dendrik

import "github.com/spf13/pflag"

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
func Parse(fs *FlagSet, args []string) error {
	return fs.inner.Parse(args)
}

// GetArgs returns positional arguments remaining after parsing.
func (fs *FlagSet) GetArgs() []string {
	return fs.inner.Args()
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
