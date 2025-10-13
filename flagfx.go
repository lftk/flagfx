// Package flagfx provides integration between Go's standard `flag` package and the `fx` dependency injection framework.
// It allows flag definitions to be co-located with the components that use them,
// ensuring that flags are parsed before any component that depends on their values is instantiated.
package flagfx

import (
	"errors"
	"flag"
	"os"
	"reflect"

	"go.uber.org/fx"
)

// Module is the core `fx.Module` for the flagfx system.
// It provides the necessary components to parse command-line flags
// and make their values available for injection into other fx components.
var Module = fx.Module("flagfx", fx.Provide(defaultFlagSet, defaultArguments, parse))

// defaultFlagSet provides the default flag set, which is the global flag.CommandLine.
// This can be replaced using the FlagSet option.
func defaultFlagSet() *flag.FlagSet {
	return flag.CommandLine
}

// defaultArguments provides the default command-line arguments, which are os.Args[1:].
// This can be replaced using the Args option.
func defaultArguments() Arguments {
	return os.Args[1:]
}

// parse is the central function that orchestrates flag registration and parsing.
// It is invoked by fx after all flag-defining entries have been collected.
func parse(fs *flag.FlagSet, args Arguments, p params) (*parsed, error) {
	parsed := &parsed{
		entries: make(map[*entry][]reflect.Value),
	}
	// Iterate through all registered entries. Each entry corresponds to a constructor
	// passed to flagfx.Provide.
	for _, e := range p.Entries {
		// Call the original constructor (e.g., `func(fs *flag.FlagSet) *MyFlags`).
		// This executes the code that registers flags (e.g., `fs.BoolVar(...)`).
		// The return values (e.g., the `*MyFlags` struct) are stored in the entries map.
		parsed.entries[e] = e.fn.Call(e.args)
	}
	// After all flags have been registered, parse the command-line arguments.
	if !fs.Parsed() {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
	}
	// The *parsed object, now containing the populated flag structs, is returned
	// and made available to the fx dependency graph.
	return parsed, nil
}

// entry represents a constructor provided to flagfx.Provide.
// It holds the constructor function itself and its captured dependencies.
type entry struct {
	fn   reflect.Value   // The original constructor function.
	args []reflect.Value // The dependencies required by the constructor.
}

// parsed is a container that holds the results of all flag-defining constructors
// after they have been executed and the flags have been parsed.
type parsed struct {
	entries map[*entry][]reflect.Value
}

// params is an fx.In struct that collects all registered *entry objects
// from the fx graph using a group tag.
type params struct {
	fx.In
	Entries []*entry `group:"flagfx_entries"`
}

// result is an fx.Out struct used to provide an *entry into the fx graph,
// tagging it to be collected by the `params` struct.
type result struct {
	fx.Out
	Entry *entry `group:"flagfx_entries"`
}

var (
	parsedPtrType  = reflect.TypeFor[*parsed]()
	parsedPtrTypes = []reflect.Type{parsedPtrType}
	resultType     = reflect.TypeFor[result]()
	resultTypes    = []reflect.Type{resultType}
)

// Provide wraps fx.Provide for use with command-line flags.
//
// Constructors passed to Provide can accept dependencies just like normal
// fx.Provide constructors. They are expected to register one or more command-line
// flags using the standard "flag" package within the constructor body.
//
// The values produced by the constructor (which typically hold the parsed flag values)
// will be made available for injection into other fx components. Fx will ensure
// that the flags are parsed before these values are provided to other components.
func Provide(constructors ...any) fx.Option {
	// This function transforms a single constructor into two separate providers
	// that are orchestrated by fx to achieve delayed parsing of flags.
	//
	// The transformation can be visualized as follows:
	//
	// Given a constructor:
	//   constructor := func(deps...) (results...)
	//
	// It is transformed into two new providers:
	//
	//   // fn1 (Wrapper Provider): Captures dependencies and registers the entry.
	//   fn1 := func(deps...) *entry { ... }
	//
	//   // fn2 (Result Provider): Returns results after flags are parsed.
	//   fn2 := func(*parsed) (results...) { ... }
	//
	// The dependency flow is: fx.New() -> fn1 -> parse() -> fn2 -> final components.

	var opts []fx.Option

	for _, constructor := range constructors {
		fn := reflect.ValueOf(constructor)
		if fn.Kind() != reflect.Func {
			opts = append(opts, fx.Error(errors.New("flagfx: Provide must be used with functions")))
			break
		}

		var (
			ft  = fn.Type()
			in  []reflect.Type
			out []reflect.Type
		)

		for i := range ft.NumIn() {
			in = append(in, ft.In(i))
		}
		for i := range ft.NumOut() {
			out = append(out, ft.Out(i))
		}

		// e is the entry that will be shared between the two generated functions.
		e := &entry{fn: fn}

		// fn1 is the Stage 1 provider. It captures the dependencies of the original constructor.
		fn1 := reflect.MakeFunc(
			reflect.FuncOf(in, resultTypes, false),
			func(args []reflect.Value) []reflect.Value {
				// Capture the arguments provided by fx.
				e.args = args
				// Return the entry in a `result` struct to be collected by the `params` group.
				r := result{Entry: e}
				return []reflect.Value{reflect.ValueOf(r)}
			},
		)

		// fn2 is the Stage 2 provider. It provides the actual results of the constructor
		// after flags have been parsed.
		fn2 := reflect.MakeFunc(
			reflect.FuncOf(parsedPtrTypes, out, false),
			func(args []reflect.Value) []reflect.Value {
				// This function should only be called by fx with a single `*parsed` argument.
				// If this invariant is broken, it's an unrecoverable internal error.
				if len(args) != 1 {
					panic("flagfx: unexpected number of arguments")
				}
				parsed, ok := typeAssert[*parsed](args[0])
				if !ok {
					panic("flagfx: unexpected argument type")
				}
				// Look up the result of the original constructor call from the parsed object.
				return parsed.entries[e]
			},
		)

		opts = append(opts, fx.Provide(fn1.Interface(), fn2.Interface()))
	}

	return fx.Options(opts...)
}

// FlagSet allows replacing the default `*flag.FlagSet` (which is flag.CommandLine)
// with a custom one.
func FlagSet(fs *flag.FlagSet) fx.Option {
	return fx.Replace(fs)
}

// Arguments represents the command-line arguments to be parsed.
type Arguments []string

// Args allows replacing the default command-line arguments (os.Args[1:])
// with a custom slice of strings.
func Args(args []string) fx.Option {
	return fx.Replace(Arguments(args))
}

// typeAssert is a compatibility wrapper for reflect.Value.Interface().(T).
// It can be replaced with reflect.TypeAssert[T](v) in Go 1.25 and later.
func typeAssert[T any](v reflect.Value) (T, bool) {
	v2, ok := v.Interface().(T)
	return v2, ok
}
