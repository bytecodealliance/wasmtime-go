// Small example of how you can interrupt the execution of a wasm module to
// ensure that it doesn't run for too long.

// You can execute this example with `go test interrupt_test.go`

package wasmtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/bytecodealliance/wasmtime-go"
)

func ExampleInterrupt() {
	// Enable interruptable code via `Config` and then create an interrupt
	// handle which we'll use later to interrupt running code.
	config := wasmtime.NewConfig()
	config.SetInterruptable(true)
	engine := wasmtime.NewEngineWithConfig(config)
	store := wasmtime.NewStore(engine)
	interrupt_handle, err := store.InterruptHandle()
	if err != nil {
		panic(err)
	}

	// Compile and instantiate a small example with an infinite loop.
	module, err := wasmtime.NewModuleFromFile(store.Engine, "interrupt.wat")
	if err != nil {
		panic(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []*wasmtime.Extern{})
	if err != nil {
		panic(err)
	}
	run := instance.GetFunc("run")
	if run == nil {
		panic("Failed to find function export `run`")
	}

	// Spin up a goroutine to send us an interrupt in a second
	go func() {
		time.Sleep(1)
		fmt.Println("Interrupting!")
		interrupt_handle.Interrupt()
	}()

	fmt.Println("Entering infinite loop ...")
	_, err = run.Call()
	trap, ok := err.(*wasmtime.Trap)
	if !ok {
		panic("Unexpected error")
	}

	fmt.Println("trap received...")
	if !strings.Contains(trap.Message(), "wasm trap: interrupt") {
		panic("Unexpected trap")
	}

	// Output:
	// Entering infinite loop ...
	// Interrupting!
	// trap received...
}
