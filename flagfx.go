// Package flagfx provides integration between Go's standard `flag` package and the `fx` dependency injection framework.
// It allows flag definitions to be co-located with the components that use them,
// ensuring that flags are parsed before any component that depends on their values is instantiated.
package flagfx

import (
	"flag"
	"os"

	"go.uber.org/fx"

	"github.com/lftk/fxbarrier"
)

// Module is the core `fx.Module` for the flagfx system.
var Module = fx.Module("flagfx",
	// Provide the default dependencies for the parse action.
	fx.Provide(defaultFlagSet, defaultArgs),
	// The barrier ensures that flags are parsed before any constructors provided
	// via this module's Provide function are invoked.
	fxbarrier.Barrier("flagfx",
		func(fs *flag.FlagSet, args Arguments) error {
			return fs.Parse(args)
		},
	),
)

// defaultFlagSet provides the default flag set, which is the global flag.CommandLine.
// This can be replaced using the FlagSet option.
func defaultFlagSet() *flag.FlagSet {
	return flag.CommandLine
}

// defaultArgs provides the default command-line arguments, which are os.Args[1:].
// This can be replaced using the Args option.
func defaultArgs() Arguments {
	return os.Args[1:]
}

// FlagSet allows replacing the default `*flag.FlagSet` (which is flag.CommandLine)
// with a custom one.
func FlagSet(fs *flag.FlagSet) fx.Option {
	return fx.Replace(fs)
}

// Arguments represents the command-line Arguments to be parsed.
type Arguments []string

// Args allows replacing the default command-line arguments (os.Args[1:])
// with a custom slice of strings.
func Args(args []string) fx.Option {
	return fx.Replace(Arguments(args))
}

// Provide is a wrapper around fxbarrier.Provide for use with command-line flags.
// It uses the "flagfx" barrier to ensure flags are parsed before dependents are instantiated.
func Provide(constructors ...any) fx.Option {
	return fxbarrier.Provide("flagfx", constructors...)
}
