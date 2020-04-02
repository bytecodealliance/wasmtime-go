package main

// This is an example of working with mulit-value modules and dealing with
// multi-value functions.

import (
	"fmt"
	"github.com/alexcrichton/wasmtime-go/pkg/wasmtime"
)

func main() {
	// Configure our `Store`, but be sure to use a `Config` that enables the
	// wasm multi-value feature since it's not stable yet.
	println("Initializing...")
	config := wasmtime.NewConfig()
	config.SetWasmMultiValue(true)
	store := wasmtime.NewStore(wasmtime.NewEngineWithConfig(config))

	println("Compiling module...")
	module, err := wasmtime.NewModuleFromFile(store, "examples/multi/multi.wat")
	check(err)

	println("Creating callback...")
	callback := wasmtime.WrapFunc(store, func(a int32, b int64) (int64, int32) {
		return b + 1, a + 1
	})

	println("Instantiating module...")
	instance, err := wasmtime.NewInstance(module, []*wasmtime.Extern{callback.AsExtern()})
	check(err)

	println("Extracting export...")
	g := instance.GetExport("g").Func()

	println("Calling export \"g\"...")
	results, err := g.Call(1, 3)
	check(err)
	arr := results.([]wasmtime.Val)
	a := arr[0].I64()
	b := arr[1].I32()
	fmt.Printf("> %d %d\n", a, b)

	assert(a == 4)
	assert(b == 2)

	println("Calling export \"round_trip_many\"...")
	round_trip_many := instance.GetExport("round_trip_many").Func()
	results, err = round_trip_many.Call(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	check(err)
	arr = results.([]wasmtime.Val)

	println("Printing result...")
	print(">")
	for i := 0; i < len(arr); i++ {
		fmt.Printf(" %d", arr[i].Get())
		assert(arr[i].I64() == int64(i))
	}
	println()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func assert(b bool) {
	if !b {
		panic("assertion failed")
	}
}
