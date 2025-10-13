package logfx

import (
	"flag"
	"strings"

	"github.com/lftk/flagfx"
	"go.uber.org/fx"
)

type LogLevel string

type flagVars struct {
	Level string
}

var Module = fx.Module("logfx",
	// Use flagfx.Provide (instead of fx.Provide) to define flags.
	flagfx.Provide(
		func(fs *flag.FlagSet) *flagVars {
			var v flagVars
			fs.StringVar(&v.Level, "log-level", "info", "log level (e.g., debug, info, warn)")
			return &v
		},
	),
	// Provide a clean LogLevel type to the container, derived from the raw flag value.
	fx.Provide(func(v *flagVars) LogLevel {
		return LogLevel(strings.ToLower(v.Level))
	}),
)
