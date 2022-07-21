package wasmtime_test

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go"
)

// An example of enabling the multi-value feature of WebAssembly and
// interacting with multi-value functions.
func ExampleConfig_multi() {
	// Configure our `Store`, but be sure to use a `Config` that enables the
	// wasm multi-value feature since it's not stable yet.
	config := wasmtime.NewConfig()
	config.SetWasmMultiValue(true)
	store := wasmtime.NewStore(wasmtime.NewEngineWithConfig(config))

	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func $f (import "" "f") (param i32 i64) (result i64 i32))

	  (func $g (export "g") (param i32 i64) (result i64 i32)
	    (call $f (local.get 0) (local.get 1))
	  )

	  (func $round_trip_many
	    (export "round_trip_many")
	    (param i64 i64 i64 i64 i64 i64 i64 i64 i64 i64)
	    (result i64 i64 i64 i64 i64 i64 i64 i64 i64 i64)

	    local.get 0
	    local.get 1
	    local.get 2
	    local.get 3
	    local.get 4
	    local.get 5
	    local.get 6
	    local.get 7
	    local.get 8
	    local.get 9
	  )
	)
	`)
	if err != nil {
		panic(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}

	callback := wasmtime.WrapFunc(store, func(a int32, b int64) (int64, int32) {
		return b + 1, a + 1
	})

	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{callback})
	if err != nil {
		panic(err)
	}

	g := instance.GetFunc(store, "g")

	results, err := g.Call(store, 1, 3)
	if err != nil {
		panic(err)
	}
	arr := results.([]wasmtime.Val)
	a := arr[0].I64()
	b := arr[1].I32()
	fmt.Printf("> %d %d\n", a, b)

	if a != 4 {
		panic("unexpected value for a")
	}
	if b != 2 {
		panic("unexpected value for b")
	}

	roundTripMany := instance.GetFunc(store, "round_trip_many")
	results, err = roundTripMany.Call(store, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	if err != nil {
		panic(err)
	}
	arr = results.([]wasmtime.Val)

	for i := 0; i < len(arr); i++ {
		fmt.Printf(" %d", arr[i].Get())
		if arr[i].I64() != int64(i) {
			panic("unexpected value for arr[i]")
		}
	}
	// Output: > 4 2
	//  0 1 2 3 4 5 6 7 8 9
}
