package main

// An example of instantiating two modules which link to each other.

import "github.com/alexcrichton/wasmtime-go/pkg/wasmtime"

func main() {
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// In this example we'll be using a `*Linker` to link modules together.
	// The first thing we put in the linker is an instance of WASI which the
	// first module we're instantiating uses.
	linker := wasmtime.NewLinker(store)
	wasi_cfg := wasmtime.NewWasiConfig()
	wasi_cfg.InheritStdout()
	wasi, err := wasmtime.NewWasiInstance(store, wasi_cfg, "wasi_snapshot_preview1")
	check(err)
	err = linker.DefineWasi(wasi)
	check(err)

	// Next we'll load up and compile our two modules
	linking1_mod, err := wasmtime.NewModuleFromFile(store, "examples/linking/linking1.wat")
	check(err)
	linking2_mod, err := wasmtime.NewModuleFromFile(store, "examples/linking/linking2.wat")
	check(err)

	// We can now instantiate the first module, and then we'll register it
	// back with the linker
	linking2, err := linker.Instantiate(linking2_mod)
	check(err)
	linker.DefineInstance("linking2", linking2)

	// And finally we can perform the final link and execute the module
	// we're interested in
	linking1, err := linker.Instantiate(linking1_mod)
	check(err)
	run := linking1.GetExport("run").Func()
	_, err = run.Call()
	check(err)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
