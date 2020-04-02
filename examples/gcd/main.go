package main

// An example of instantiating a small wasm modules which calculates the
// greatest common divisor of two inputs.

import (
	"fmt"
	"github.com/alexcrichton/wasmtime-go/pkg/wasmtime"
)

func main() {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	module, err := wasmtime.NewModuleFromFile(store, "./examples/gcd/gcd.wat")
	check(err)
	instance, err := wasmtime.NewInstance(module, []*wasmtime.Extern{})
	check(err)
	run := instance.GetExport("gcd").Func()
	result, err := run.Call(6, 27)
	check(err)
	fmt.Printf("gcd(6, 27) = %d\n", result.(int32))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
