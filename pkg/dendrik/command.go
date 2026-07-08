package dendrik

import (
	"fmt"
	"io"
	"os"
)

// ArgSpec is a leaf command's positional-argument arity. Max < 0 == unbounded.
type ArgSpec struct{ Min, Max int }

// ArgsNone rejects any positional argument.
var ArgsNone = ArgSpec{0, 0}

// ArgsExactly requires exactly n positionals.
func ArgsExactly(n int) ArgSpec { return ArgSpec{n, n} }

// ArgsBetween requires min..max positionals (max < 0 == unbounded).
func ArgsBetween(min, max int) ArgSpec { return ArgSpec{min, max} }

// Command is one node in a CLI's subcommand tree. Exactly one of Sub, Run, or
// RunRaw must be set (enforced at Execute time). A Group node (Sub non-empty) is
// pure dispatch; a leaf runs work. Building the tree through this type is what
// makes help, unknown-subcommand, flag parsing, and arity uniform and
// unskippable across every consumer.
type Command struct {
	Name    string
	Short   string                              // one-line summary; feeds auto usage
	Version string                              // root only: injected per tool, enables version/-V
	Sub     []Command                           // non-empty => group (pure dispatch)
	Flags   func(fs *FlagSet)                   // leaf: register flags (nil ok)
	Args    ArgSpec                             // leaf: positional arity
	Run     func(fs *FlagSet, pos []string) int // leaf handler
	RunRaw  func(args []string) int             // lint-visible escape hatch (self-parses)
}

// Execute runs the router over args (typically os.Args[1:]).
func (c Command) Execute(args []string) int { return c.execute(c.Name, args) }

func (c Command) execute(path string, args []string) int {
	c.mustBeWellFormed()
	switch {
	case len(c.Sub) > 0:
		return c.executeGroup(path, args)
	case c.RunRaw != nil:
		return c.RunRaw(args)
	default:
		return c.executeLeaf(path, args)
	}
}

// mustBeWellFormed panics if the node does not set exactly one of Sub/Run/RunRaw.
// This is a dev-time construction invariant exercised by tests, not a user error.
func (c Command) mustBeWellFormed() {
	n := 0
	if len(c.Sub) > 0 {
		n++
	}
	if c.Run != nil {
		n++
	}
	if c.RunRaw != nil {
		n++
	}
	if n != 1 {
		panic(fmt.Sprintf("dendrik.Command %q: exactly one of Sub/Run/RunRaw must be set (got %d)", c.Name, n))
	}
}

func (c Command) executeGroup(path string, args []string) int {
	if len(args) == 0 {
		c.printGroupUsage(os.Stderr, path)
		return ExitUserError
	}
	head := args[0]
	if IsHelpArg(head) {
		c.printGroupUsage(os.Stderr, path)
		return ExitOK
	}
	if c.Version != "" && isVersionArg(head) {
		fmt.Printf("%s %s\n", c.Name, c.Version)
		return ExitOK
	}
	for _, sub := range c.Sub {
		if sub.Name == head {
			return sub.execute(path+" "+sub.Name, args[1:])
		}
	}
	fmt.Fprintf(os.Stderr, "Unknown %s command: %s\n", c.Name, head)
	c.printGroupUsage(os.Stderr, path)
	return ExitUserError
}

func (c Command) executeLeaf(path string, args []string) int {
	fs := NewFlagSet(path)
	if c.Flags != nil {
		c.Flags(fs)
	}
	if done, code := ParseCheck(fs, args); done {
		return code
	}
	pos := fs.GetArgs()
	if len(pos) < c.Args.Min {
		fmt.Fprintf(os.Stderr, "missing argument: %s expects at least %d, got %d\n", path, c.Args.Min, len(pos))
		return ExitUserError
	}
	if c.Args.Max >= 0 && len(pos) > c.Args.Max {
		fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", pos[c.Args.Max])
		return ExitUserError
	}
	return c.Run(fs, pos)
}

func (c Command) printGroupUsage(w io.Writer, path string) {
	fmt.Fprintf(w, "Usage: %s <command> [flags]\n\nCommands:\n", path)
	width := 0
	for _, s := range c.Sub {
		if len(s.Name) > width {
			width = len(s.Name)
		}
	}
	for _, s := range c.Sub {
		fmt.Fprintf(w, "  %-*s  %s\n", width, s.Name, s.Short)
	}
}

func isVersionArg(s string) bool {
	return s == "version" || s == "--version" || s == "-V"
}
