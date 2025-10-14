package logfx

import (
	"flag"
	"strings"

	"github.com/lftk/flagfx"
	"go.uber.org/fx"
)

type LogLevel string

type flags struct {
	Level string
}

var Module = fx.Module("logfx",
	// Use flagfx.Provide (instead of fx.Provide) to define flags.
	flagfx.Provide(
		func(fs *flag.FlagSet) *flags {
			var f flags
			fs.StringVar(&f.Level, "log-level", "info", "log level (e.g., debug, info, warn)")
			return &f
		},
	),
	// Provide a clean LogLevel type to the container, derived from the raw flag value.
	fx.Provide(
		func(f *flags) LogLevel {
			return LogLevel(strings.ToLower(f.Level))
		},
	),
)
