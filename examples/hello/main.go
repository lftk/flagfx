package main

import (
	"fmt"

	"github.com/lftk/flagfx"
	"github.com/lftk/flagfx/examples/hello/logfx"
	"github.com/lftk/flagfx/examples/hello/verfx"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		// Disable fx's default logger for a clean output in this example.
		fx.NopLogger,
		// Add the core flagfx.Module to enable flag parsing.
		flagfx.Module,
		// Add other modules that use flagfx to define their own flags.
		verfx.Module, logfx.Module,
		// Override the default version string in the verfx module.
		verfx.Version("v0.1.0"),
		// Invoke a function that uses the LogLevel provided by the logfx module.
		fx.Invoke(func(level logfx.LogLevel) {
			fmt.Println("Log level set to:", level)
		}),
	)
	app.Run()
}
