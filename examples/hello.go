package main

// An example of instantiating a small wasm module which imports some
// functionality from the host.

import "github.com/alexcrichton/wasmtime-go/pkg/wasmtime"

func main() {
	// Almost all operations in wasmtime require a contextual `store`
	// argument to share, so create that first
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := wasmtime.NewModuleFromFile(store, "./examples/hello.wat")
	check(err)

	// Our `hello.wat` file imports one item, so we create that function
	// here.
	item := wasmtime.WrapFunc(store, func() {
		println("Hello from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one improt so we pass that in here.
	instance, err := wasmtime.NewInstance(module, []*wasmtime.Extern{item.AsExtern()})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	run := instance.GetExport("run").Func()
	_, err = run.Call()
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
