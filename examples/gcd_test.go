// Example of instantiating of the WebAssembly module and invoking its exported
// function.

// You can execute this example with `go test gcd_test.go`

package wasmtime

import (
	"fmt"

	"github.com/bytecodealliance/wasmtime-go"
)

func ExampleGcd() {
	// Load our WebAssembly (parsed WAT in our case) into a `Module`
	// which is attached to a `Store` cache. After we've got that we
	// can instantiate it.
	store := wasmtime.NewStore(wasmtime.NewEngine())
	module, err := wasmtime.NewModuleFromFile(store.Engine, "gcd.wat")
	if err != nil {
		panic(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []*wasmtime.Extern{})
	if err != nil {
		panic(err)
	}

	// Invoke `gcd` export.
	gcd := instance.GetFunc("gcd")
	if gcd == nil {
		panic("Failed to find function export `gcd`")
	}
	result, err := gcd.Call(6, 27)
	if err != nil {
		panic(err)
	}

	fmt.Printf("gcd(6, 27) = %d\n", result)

	// Output:
	// gcd(6, 27) = 3
}
