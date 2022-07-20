package wasmtime_test

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go"
)

func Example_linker() {
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// Compile two wasm modules where the first references the second
	wasm1, err := wasmtime.Wat2Wasm(`
	(module
	  (import "wasm2" "double" (func $double (param i32) (result i32)))
	  (func (export "double_and_add") (param i32 i32) (result i32)
	    local.get 0
	    call $double
	    local.get 1
	    i32.add
	  )
	)`)
	if err != nil {
		panic(err)
	}

	wasm2, err := wasmtime.Wat2Wasm(`
	(module
	  (func (export "double") (param i32) (result i32)
	    local.get 0
	    i32.const 2
	    i32.mul
	  )
	)`)
	if err != nil {
		panic(err)
	}

	// Next compile both modules
	module1, err := wasmtime.NewModule(store.Engine, wasm1)
	if err != nil {
		panic(err)
	}
	module2, err := wasmtime.NewModule(store.Engine, wasm2)
	if err != nil {
		panic(err)
	}

	linker := wasmtime.NewLinker(store.Engine)

	// The second module is instantiated first since it has no imports, and
	// then we insert the instance back into the linker under the name
	// the first module expects.
	instance2, err := linker.Instantiate(store, module2)
	if err != nil {
		panic(err)
	}
	err = linker.DefineInstance(store, "wasm2", instance2)
	if err != nil {
		panic(err)
	}

	// And now we can instantiate our first module, executing the result
	// afterwards
	instance1, err := linker.Instantiate(store, module1)
	if err != nil {
		panic(err)
	}
	doubleAndAdd := instance1.GetFunc(store, "double_and_add")
	result, err := doubleAndAdd.Call(store, 2, 3)
	if err != nil {
		panic(err)
	}
	fmt.Print(result.(int32))
	// Output: 7
}
