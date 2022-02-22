// Example of limiting a WebAssembly function's runtime using "fuel consumption".

package wasmtime_test

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go"
)

func Example_fuel() {
	config := wasmtime.NewConfig()
	config.SetConsumeFuel(true)
	engine := wasmtime.NewEngineWithConfig(config)
	store := wasmtime.NewStore(engine)
	err := store.AddFuel(10000)
	if err != nil {
		panic(err)
	}

	// Compile and instantiate a small example with an infinite loop.
	wasm, err := wasmtime.Wat2Wasm(`
	(module
          (func $fibonacci (param $n i32) (result i32)
            (if
              (i32.lt_s (local.get $n) (i32.const 2))
              (return (local.get $n))
            )
            (i32.add
              (call $fibonacci (i32.sub (local.get $n) (i32.const 1)))
              (call $fibonacci (i32.sub (local.get $n) (i32.const 2)))
            )
          )
          (export "fibonacci" (func $fibonacci))
        )`)
	if err != nil {
		panic(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		panic(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		panic(err)
	}

	// Invoke `fibonacci` export with higher and higher numbers until we exhaust our fuel.
	fibonacci := instance.GetFunc(store, "fibonacci")
	if fibonacci == nil {
		panic("Failed to find function export `fibonacci`")
	}
	for n := 0; ; n++ {
		fuelBefore, _ := store.FuelConsumed()
		output, err := fibonacci.Call(store, n)
		if err != nil {
			fmt.Println(err)
			break
		}
		fuelAfter, _ := store.FuelConsumed()
		fmt.Printf("fib(%d) = %d [consumed %d fuel]\n", n, output, fuelAfter-fuelBefore)
		err = store.AddFuel(fuelAfter - fuelBefore)
		if err != nil {
			panic(err)
		}
	}
	// Output:
	// fib(0) = 0 [consumed 6 fuel]
	// fib(1) = 1 [consumed 6 fuel]
	// fib(2) = 1 [consumed 26 fuel]
	// fib(3) = 2 [consumed 46 fuel]
	// fib(4) = 3 [consumed 86 fuel]
	// fib(5) = 5 [consumed 146 fuel]
	// fib(6) = 8 [consumed 246 fuel]
	// fib(7) = 13 [consumed 406 fuel]
	// fib(8) = 21 [consumed 666 fuel]
	// fib(9) = 34 [consumed 1086 fuel]
	// fib(10) = 55 [consumed 1766 fuel]
	// fib(11) = 89 [consumed 2866 fuel]
	// fib(12) = 144 [consumed 4646 fuel]
	// fib(13) = 233 [consumed 7526 fuel]
	// all fuel consumed by WebAssembly
}
