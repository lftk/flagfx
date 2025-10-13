package verfx

import (
	"flag"
	"fmt"
	"os"

	"github.com/lftk/flagfx"
	"go.uber.org/fx"
)

type flagVars struct {
	ShowVersion bool
}

type version string

var Module = fx.Module("verfx",
	// Use flagfx.Provide (instead of fx.Provide) to define flags.
	flagfx.Provide(
		func(fs *flag.FlagSet) *flagVars {
			var v flagVars
			fs.BoolVar(&v.ShowVersion, "version", false, "show version")
			return &v
		},
	),
	// Invoke a function that checks the flag and acts accordingly.
	fx.Invoke(
		func(v *flagVars, ver version) {
			if v.ShowVersion {
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
