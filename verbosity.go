package cli

// Verbosity is an optional embeddable flag group giving an author the
// conventional -v/-vv/-vvv switches without the framework imposing any
// verbosity policy of its own. Embed it in a command's input struct to get the
// flags and the level query for free:
//
//	type MyInputs struct {
//		cli.Verbosity
//		Name string `arg:"name"`
//	}
//
// The three flags are independent switches matched as their own short tokens
// (-v, -vv, -vvv), not a true repeat-counter. Use Level to read the highest one
// the user passed rather than touching the bool fields directly.
type Verbosity struct {
	V1 bool `arg:"v" short:"v" help:"Verbose output (-v)"`
	V2 bool `arg:"vv" short:"vv" help:"More verbose output (-vv)"`
	V3 bool `arg:"vvv" short:"vvv" help:"Most verbose output (-vvv)"`
}

// Level returns the verbosity the user selected: 0 when no flag was passed, up
// to 3 for -vvv. The highest flag wins, so -vvv reports 3 even if -v is also set.
func (v Verbosity) Level() int {
	switch {
	case v.V3:
		return 3
	case v.V2:
		return 2
	case v.V1:
		return 1
	default:
		return 0
	}
}

// Verbose reports whether any verbosity flag was passed (Level > 0).
func (v Verbosity) Verbose() bool {
	return v.Level() > 0
}

// AtLeast reports whether the selected verbosity is at or above level.
func (v Verbosity) AtLeast(level int) bool {
	return v.Level() >= level
}
