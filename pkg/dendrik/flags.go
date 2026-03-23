package dendrik

import "github.com/peterbourgon/ff/v4"

// NewFlagSet creates an ff.FlagSet with standard dendrik conventions.
func NewFlagSet(name string) *ff.FlagSet {
	return ff.NewFlagSet(name)
}

// Parse parses args with the full config hierarchy: flags > env > config file.
// Returns error on failure — dendrik never calls os.Exit.
func Parse(fs *ff.FlagSet, args []string, opts ...ff.Option) error {
	return ff.Parse(fs, args, opts...)
}
