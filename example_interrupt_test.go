// Small example of how you can interrupt the execution of a wasm module to
// ensure that it doesn't run for too long.

package wasmtime_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bytecodealliance/wasmtime-go"
)

func Example_interrupt() {
	// Enable interruptable code via `Config` and then create an interrupt
	// handle which we'll use later to interrupt running code.
	config := wasmtime.NewConfig()
	config.SetEpochInterruption(true)
	engine := wasmtime.NewEngineWithConfig(config)
	store := wasmtime.NewStore(engine)
	store.SetEpochDeadline(1)

	// Compile and instantiate a small example with an infinite loop.
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func (export "run")
	    (loop
	      br 0)
	  )
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
	run := instance.GetFunc(store, "run")
	if run == nil {
		panic("Failed to find function export `run`")
	}

	// Spin up a goroutine to send us an interrupt in a second
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Interrupting!")
		engine.IncrementEpoch()
	}()

	fmt.Println("Entering infinite loop ...")
	_, err = run.Call(store)
	var trap *wasmtime.Trap
	if !errors.As(err, &trap) {
		panic("Unexpected error")
	}

	fmt.Println("trap received...")
	if !strings.Contains(trap.Message(), "wasm trap: interrupt") {
		panic("Unexpected trap: " + trap.Message())
	}

	// Output:
	// Entering infinite loop ...
	// Interrupting!
	// trap received...
}
