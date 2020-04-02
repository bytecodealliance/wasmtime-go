package wasmtime

import "fmt"

func Example() {
	// Almost all operations in wasmtime require a contextual `store`
	// argument to share, so create that first
	store := NewStore(NewEngine())

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	wasm, err := Wat2Wasm(`
	    (module
		(import "" "hello" (func $hello))
		(func (export "run")
		  (call $hello))
	    )
	`)
	check(err)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := NewModule(store, wasm)
	check(err)

	// Our `hello.wat` file imports one item, so we create that function
	// here.
	item := WrapFunc(store, func() {
		fmt.Println("Hello from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one improt so we pass that in here.
	instance, err := NewInstance(module, []*Extern{item.AsExtern()})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	run := instance.GetExport("run").Func()
	_, err = run.Call()
	check(err)

	// Output: Hello from Go!
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
