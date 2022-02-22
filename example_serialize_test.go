// Small example of how to serialize a compiled wasm module, and then
// instantiate it from the compilation artifacts.

package wasmtime_test

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go"
)

func serialize() []byte {
	// Configure the initial compilation environment.
	fmt.Println("Initializing...")
	engine := wasmtime.NewEngine()

	// Compile the wasm module into an in-memory instance of a `Module`.
	fmt.Println("Compiling module...")
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func $hello (import "" "hello"))
	  (func (export "run") (call $hello))
	)`)
	if err != nil {
		panic(err)
	}
	module, err := wasmtime.NewModule(engine, wasm)
	if err != nil {
		panic(err)
	}
	serialized, err := module.Serialize()
	if err != nil {
		panic(err)
	}

	fmt.Println("Serialized.")
	return serialized
}

func deserialize(encoded []byte) {
	// Configure the initial compilation environment.
	fmt.Println("Initializing...")
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// Deserialize the compiled module.
	fmt.Println("Deserialize module...")
	module, err := wasmtime.NewModuleDeserialize(store.Engine, encoded)
	if err != nil {
		panic(err)
	}

	// Here we handle the imports of the module, which in this case is our
	// `helloFunc` callback.
	fmt.Println("Creating callback...")
	helloFunc := wasmtime.WrapFunc(store, func() {
		fmt.Println("Calling back...")
		fmt.Println("> Hello World!")
	})

	// Once we've got that all set up we can then move to the instantiation
	// phase, pairing together a compiled module as well as a set of imports.
	// Note that this is where the wasm `start` function, if any, would run.
	fmt.Println("Instantiating module...")
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{helloFunc})
	if err != nil {
		panic(err)
	}

	// Next we poke around a bit to extract the `run` function from the module.
	fmt.Println("Extracting export...")
	run := instance.GetFunc(store, "run")
	if run == nil {
		panic("Failed to find function export `run`")
	}

	// And last but not least we can call it!
	fmt.Println("Calling export...")
	_, err = run.Call(store)
	if err != nil {
		panic(err)
	}

	fmt.Println("Done.")
}

func Example_serialize() {
	bytes := serialize()
	deserialize(bytes)

	// Output:
	// Initializing...
	// Compiling module...
	// Serialized.
	// Initializing...
	// Deserialize module...
	// Creating callback...
	// Instantiating module...
	// Extracting export...
	// Calling export...
	// Calling back...
	// > Hello World!
	// Done.
}
