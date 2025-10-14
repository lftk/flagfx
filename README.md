# flagfx

`flagfx` provides seamless integration between Go's standard `flag` package and the `uber-go/fx` dependency injection framework. It allows you to co-locate flag definitions with the components that use them, while ensuring that flags are parsed before any component that depends on their values is instantiated.

## Installation

```sh
go get github.com/lftk/flagfx
```

## Usage

Here is a simple, complete example of how to use `flagfx`.

```go
package main

import (
	"flag"
	"fmt"

	"github.com/lftk/flagfx"
	"go.uber.org/fx"
)

// 1. Define a struct to hold your flag values.
type flags struct {
	name string
}

func main() {
	app := fx.New(
		// 2. Add the core flagfx.Module.
		flagfx.Module,

		// 3. Use flagfx.Provide to define flags.
		// This constructor will be called before flags are parsed.
		flagfx.Provide(func(fs *flag.FlagSet) *flags {
			var f flags
			fs.StringVar(&f.name, "name", "World", "name to greet")
			return &f
		}),

		// 4. Use the parsed flag values in your application.
		// This fx.Invoke function will run after flags are parsed.
		fx.Invoke(func(f *flags) {
			fmt.Printf("Hello, %s!\n", f.name)
		}),

		// Suppress fx's logger for a clean output.
		fx.NopLogger,
	)
	app.Run()
}
```

### Running the example

```sh
# Run with the default flag value
$ go run .
Hello, World!

# Run with a custom flag value
$ go run . -name=flagfx
Hello, flagfx!
```

## Advanced Examples

For more advanced, modular examples, please see the [`examples`](./examples) directory.

## License

This project is licensed under the MIT License.
