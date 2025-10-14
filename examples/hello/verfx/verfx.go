package verfx

import (
	"flag"
	"fmt"
	"os"

	"github.com/lftk/flagfx"
	"go.uber.org/fx"
)

type flags struct {
	ShowVersion bool
}

type version string

var Module = fx.Module("verfx",
	// Use flagfx.Provide (instead of fx.Provide) to define flags.
	flagfx.Provide(
		func(fs *flag.FlagSet) *flags {
			var f flags
			fs.BoolVar(&f.ShowVersion, "version", false, "show version")
			return &f
		},
	),
	// Invoke a function that checks the flag and acts accordingly.
	fx.Invoke(
		func(f *flags, ver version) {
			if f.ShowVersion {
				fmt.Println("Version:", ver)
				os.Exit(0)
			}
		},
	),
	// Supply a default version string, which can be overridden.
	fx.Supply(version("unknown")),
)

// Version returns an fx.Option that replaces the default version string.
func Version(ver string) fx.Option {
	return fx.Replace(version(ver))
}
