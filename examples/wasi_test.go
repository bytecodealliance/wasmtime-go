// Example of instantiating a wasm module which uses WASI imports.

// You can execute this example with `go test wasi_test.go`

package wasmtime

import "github.com/bytecodealliance/wasmtime-go"

func ExampleWasi() {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	linker := wasmtime.NewLinker(store)

	// Create an instance of `Wasi` which contains a `WasiConfig`. Note that
	// `WasiConfig` provides a number of ways to configure what the target program
	// will have access to.
	config := wasmtime.NewWasiConfig()
	config.InheritStdout()
	config.InheritArgv()
	wasi, err := wasmtime.NewWasiInstance(store, config, "wasi_snapshot_preview1")
	if err != nil {
		panic(err)
	}
	linker.DefineWasi(wasi)

	// Instantiate our module with the imports we've created, and run it.
	module, err := wasmtime.NewModuleFromFile(store.Engine, "wasi.wat")
	if err != nil {
		panic(err)
	}
	err = linker.DefineModule("", module)
	if err != nil {
		panic(err)
	}
	fn, err := linker.GetDefault("")
	if err != nil {
		panic(err)
	}
	_, err = fn.Call()
	if err != nil {
		panic(err)
	}

	// Output:
}
