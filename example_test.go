package wasmtime_test

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v24"
)

// Example of limiting a WebAssembly function's runtime using "fuel consumption".
func ExampleConfig_fuel() {
	config := wasmtime.NewConfig()
	config.SetConsumeFuel(true)
	engine := wasmtime.NewEngineWithConfig(config)
	store := wasmtime.NewStore(engine)
	err := store.SetFuel(10000)
	if err != nil {
		log.Fatal(err)
	}

	// Compile and instantiate a small example with an infinite loop.
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func $fibonacci (param $n i32) (result i32)
	    (if
	      (i32.lt_s (local.get $n) (i32.const 2))
	      (then (return (local.get $n)))
	    )
	    (i32.add
	      (call $fibonacci (i32.sub (local.get $n) (i32.const 1)))
	      (call $fibonacci (i32.sub (local.get $n) (i32.const 2)))
	    )
	  )
	  (export "fibonacci" (func $fibonacci))
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		log.Fatal(err)
	}

	// Invoke `fibonacci` export with higher and higher numbers until we exhaust our fuel.
	fibonacci := instance.GetFunc(store, "fibonacci")
	if fibonacci == nil {
		log.Fatal("Failed to find function export `fibonacci`")
	}
	fuel := uint64(10000)
	for n := 0; ; n++ {
		err := store.SetFuel(fuel)
		if err != nil {
			log.Fatal(err)
		}
		output, err := fibonacci.Call(store, n)
		if err != nil {
			break
		}
		fuelAfter, err := store.GetFuel()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("fib(%d) = %d [consumed %d fuel]\n", n, output, fuel-fuelAfter)
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
}

// Small example of how you can interrupt the execution of a wasm module to
// ensure that it doesn't run for too long.
func ExampleConfig_interrupt() {
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
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		log.Fatal(err)
	}
	run := instance.GetFunc(store, "run")
	if run == nil {
		log.Fatal("Failed to find function export `run`")
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
		log.Fatal("Unexpected error")
	}

	fmt.Println("trap received...")
	if !strings.Contains(trap.Message(), "wasm trap: interrupt") {
		log.Fatalf("Unexpected trap: %s", trap.Message())
	}
	// Output:
	// Entering infinite loop ...
	// Interrupting!
	// trap received...
}

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
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}

	callback := wasmtime.WrapFunc(store, func(a int32, b int64) (int64, int32) {
		return b + 1, a + 1
	})

	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{callback})
	if err != nil {
		log.Fatal(err)
	}

	g := instance.GetFunc(store, "g")

	results, err := g.Call(store, 1, 3)
	if err != nil {
		log.Fatal(err)
	}
	arr := results.([]wasmtime.Val)
	a := arr[0].I64()
	b := arr[1].I32()
	fmt.Printf("> %d %d\n", a, b)

	if a != 4 {
		log.Fatal("unexpected value for a")
	}
	if b != 2 {
		log.Fatal("unexpected value for b")
	}

	roundTripMany := instance.GetFunc(store, "round_trip_many")
	results, err = roundTripMany.Call(store, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	if err != nil {
		log.Fatal(err)
	}
	arr = results.([]wasmtime.Val)

	for i := 0; i < len(arr); i++ {
		fmt.Printf(" %d", arr[i].Get())
		if arr[i].I64() != int64(i) {
			log.Fatal("unexpected value for arr[i]")
		}
	}
	// Output: > 4 2
	//  0 1 2 3 4 5 6 7 8 9
}

func ExampleLinker() {
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
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	wasm2, err := wasmtime.Wat2Wasm(`
	(module
	  (func (export "double") (param i32) (result i32)
	    local.get 0
	    i32.const 2
	    i32.mul
	  )
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Next compile both modules
	module1, err := wasmtime.NewModule(store.Engine, wasm1)
	if err != nil {
		log.Fatal(err)
	}
	module2, err := wasmtime.NewModule(store.Engine, wasm2)
	if err != nil {
		log.Fatal(err)
	}

	linker := wasmtime.NewLinker(store.Engine)

	// The second module is instantiated first since it has no imports, and
	// then we insert the instance back into the linker under the name
	// the first module expects.
	instance2, err := linker.Instantiate(store, module2)
	if err != nil {
		log.Fatal(err)
	}
	err = linker.DefineInstance(store, "wasm2", instance2)
	if err != nil {
		log.Fatal(err)
	}

	// And now we can instantiate our first module, executing the result
	// afterwards
	instance1, err := linker.Instantiate(store, module1)
	if err != nil {
		log.Fatal(err)
	}
	doubleAndAdd := instance1.GetFunc(store, "double_and_add")
	result, err := doubleAndAdd.Call(store, 2, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(result.(int32))
	// Output: 7
}

// An example of working with the Memory type to read/write wasm memory.
func ExampleMemory() {
	// Create our `Store` context and then compile a module and create an
	// instance from the compiled module all in one go.
	store := wasmtime.NewStore(wasmtime.NewEngine())
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (memory (export "memory") 2 3)

	  (func (export "size") (result i32) (memory.size))
	  (func (export "load") (param i32) (result i32)
	    (i32.load8_s (local.get 0))
	  )
	  (func (export "store") (param i32 i32)
	    (i32.store8 (local.get 0) (local.get 1))
	  )

	  (data (i32.const 0x1000) "\01\02\03\04")
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		log.Fatal(err)
	}

	// Load up our exports from the instance
	memory := instance.GetExport(store, "memory").Memory()
	sizeFn := instance.GetFunc(store, "size")
	loadFn := instance.GetFunc(store, "load")
	storeFn := instance.GetFunc(store, "store")

	// some helper functions we'll use below
	call32 := func(f *wasmtime.Func, args ...interface{}) int32 {
		ret, err := f.Call(store, args...)
		if err != nil {
			log.Fatal(err)
		}
		return ret.(int32)
	}
	call := func(f *wasmtime.Func, args ...interface{}) {
		_, err := f.Call(store, args...)
		if err != nil {
			log.Fatal(err)
		}
	}
	assertTraps := func(f *wasmtime.Func, args ...interface{}) {
		_, err := f.Call(store, args...)
		_, ok := err.(*wasmtime.Trap)
		if !ok {
			log.Fatal("expected a trap")
		}
	}
	assert := func(b bool) {
		if !b {
			log.Fatal("assertion failed")
		}
	}

	// Check the initial memory sizes/contents
	assert(memory.Size(store) == 2)
	assert(memory.DataSize(store) == 0x20000)
	buf := memory.UnsafeData(store)

	assert(buf[0] == 0)
	assert(buf[0x1000] == 1)
	assert(buf[0x1003] == 4)

	assert(call32(sizeFn) == 2)
	assert(call32(loadFn, 0) == 0)
	assert(call32(loadFn, 0x1000) == 1)
	assert(call32(loadFn, 0x1003) == 4)
	assert(call32(loadFn, 0x1ffff) == 0)
	assertTraps(loadFn, 0x20000)

	// We can mutate memory as well
	buf[0x1003] = 5
	call(storeFn, 0x1002, 6)
	assertTraps(storeFn, 0x20000, 0)

	assert(buf[0x1002] == 6)
	assert(buf[0x1003] == 5)
	assert(call32(loadFn, 0x1002) == 6)
	assert(call32(loadFn, 0x1003) == 5)

	// And like wasm instructions, we can grow memory
	_, err = memory.Grow(store, 1)
	assert(err == nil)
	assert(memory.Size(store) == 3)
	assert(memory.DataSize(store) == 0x30000)

	assert(call32(loadFn, 0x20000) == 0)
	call(storeFn, 0x20000, 0)
	assertTraps(loadFn, 0x30000)
	assertTraps(storeFn, 0x30000, 0)

	// Memory can fail to grow
	_, err = memory.Grow(store, 1)
	assert(err != nil)
	_, err = memory.Grow(store, 0)
	assert(err == nil)

	// Ensure that `memory` lives long enough to cover all our usages of
	// using its internal buffer we read from `UnsafeData()`
	runtime.KeepAlive(memory)

	// Finally we can also create standalone memories to get imported by
	// wasm modules too.
	memorytype := wasmtime.NewMemoryType(5, true, 5)
	memory2, err := wasmtime.NewMemory(store, memorytype)
	assert(err == nil)
	assert(memory2.Size(store) == 5)
	_, err = memory2.Grow(store, 1)
	assert(err != nil)
	_, err = memory2.Grow(store, 0)
	assert(err == nil)
	// Output:
	//
}

// Small example of how to serialize a compiled wasm module, and then
// instantiate it from the compilation artifacts.
func ExampleModule_serialize() {
	// Configure the initial compilation environment.
	engine := wasmtime.NewEngine()

	// Compile the wasm module into an in-memory instance of a `Module`.
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func $hello (import "" "hello"))
	  (func (export "run") (call $hello))
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := module.Serialize()
	if err != nil {
		log.Fatal(err)
	}

	// Configure the initial compilation environment.
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// Deserialize the compiled module.
	module, err = wasmtime.NewModuleDeserialize(store.Engine, bytes)
	if err != nil {
		log.Fatal(err)
	}

	// Here we handle the imports of the module, which in this case is our
	// `helloFunc` callback.
	helloFunc := wasmtime.WrapFunc(store, func() {
		fmt.Println("Calling back...")
		fmt.Println("> Hello World!")
	})

	// Once we've got that all set up we can then move to the instantiation
	// phase, pairing together a compiled module as well as a set of imports.
	// Note that this is where the wasm `start` function, if any, would run.
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{helloFunc})
	if err != nil {
		log.Fatal(err)
	}

	// Next we poke around a bit to extract the `run` function from the module.
	run := instance.GetFunc(store, "run")
	if run == nil {
		log.Fatal("Failed to find function export `run`")
	}

	// And last but not least we can call it!
	fmt.Println("Calling export...")
	_, err = run.Call(store)
	if err != nil {
		log.Fatal(err)
	}
	// Output:
	// Calling export...
	// Calling back...
	// > Hello World!
}

// Small example of how to use `externref`s.
func ExampleVal_Externref() {
	config := wasmtime.NewConfig()
	config.SetWasmReferenceTypes(true)
	store := wasmtime.NewStore(wasmtime.NewEngineWithConfig(config))
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (table $table (export "table") 10 externref)

	  (global $global (export "global") (mut externref) (ref.null extern))

	  (func (export "func") (param externref) (result externref)
	    local.get 0
	  )
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		log.Fatal(err)
	}
	// Create a new `externref` value.
	value := wasmtime.ValExternref("Hello, World!")
	// The `externref`'s wrapped data should be the string "Hello, World!".
	externRef := value.Externref()
	if externRef != "Hello, World!" {
		log.Fatal("unexpected value")
	}
	// Lookup the `table` export.
	table := instance.GetExport(store, "table").Table()
	// Set `table[3]` to our `externref`.
	err = table.Set(store, 3, value)
	if err != nil {
		log.Fatal(err)
	}
	// `table[3]` should now be our `externref`.
	tableValue, err := table.Get(store, 3)
	if err != nil {
		log.Fatal(err)
	}
	if tableValue.Externref() != externRef {
		log.Fatal("unexpected value in table")
	}
	// Lookup the `global` export.
	global := instance.GetExport(store, "global").Global()
	// Set the global to our `externref`.
	err = global.Set(store, value)
	if err != nil {
		log.Fatal(err)
	}
	// Get the global, and it should return our `externref` again.
	globalValue := global.Get(store)
	if globalValue.Externref() != externRef {
		log.Fatal("unexpected value in global")
	}
	// Lookup the `func` export.
	fn := instance.GetFunc(store, "func")
	// And call it!
	result, err := fn.Call(store, value)
	if err != nil {
		log.Fatal(err)
	}
	// `func` returns the same reference we gave it, so `results` should be
	// our `externref`.
	if result != externRef {
		log.Fatal("unexpected value from func")
	}
	// Output:
	//
}

// An example of linking WASI to the runtime in order to interact with the system.
// It uses the WAT code from https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-tutorial.md#web-assembly-text-example
func ExampleWasiConfig() {
	dir, err := os.MkdirTemp("", "out")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	stdoutPath := filepath.Join(dir, "stdout")

	engine := wasmtime.NewEngine()

	// Create our module
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  ;; Import the required fd_write WASI function which will write the given io vectors to stdout
	  ;; The function signature for fd_write is:
	  ;; (File Descriptor, *iovs, iovs_len, nwritten) -> Returns number of bytes written
	  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))

	  (memory 1)
	  (export "memory" (memory 0))

	  ;; Write 'hello world\n' to memory at an offset of 8 bytes
	  ;; Note the trailing newline which is required for the text to appear
	  (data (i32.const 8) "hello world\n")

	  (func $main (export "_start")
	    ;; Creating a new io vector within linear memory
	    (i32.store (i32.const 0) (i32.const 8))  ;; iov.iov_base - This is a pointer to the start of the 'hello world\n' string
	    (i32.store (i32.const 4) (i32.const 12))  ;; iov.iov_len - The length of the 'hello world\n' string

	    (call $fd_write
	      (i32.const 1) ;; file_descriptor - 1 for stdout
	      (i32.const 0) ;; *iovs - The pointer to the iov array, which is stored at memory location 0
	      (i32.const 1) ;; iovs_len - We're printing 1 string stored in an iov - so one.
	      (i32.const 20) ;; nwritten - A place in memory to store the number of bytes written
	    )
	    drop ;; Discard the number of bytes written from the top of the stack
	  )
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(engine, wasm)
	if err != nil {
		log.Fatal(err)
	}

	// Create a linker with WASI functions defined within it
	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		log.Fatal(err)
	}

	// Configure WASI imports to write stdout into a file, and then create
	// a `Store` using this wasi configuration.
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutPath)
	store := wasmtime.NewStore(engine)
	store.SetWasi(wasiConfig)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		log.Fatal(err)
	}

	// Run the function
	nom := instance.GetFunc(store, "_start")
	_, err = nom.Call(store)
	if err != nil {
		log.Fatal(err)
	}

	// Print WASM stdout
	out, err := os.ReadFile(stdoutPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(out))
	// Output: hello world
}

// An example of instantiating a small wasm module which imports functionality
// from the host, then calling into wasm which calls back into the host.
func Example_hello() {
	// Almost all operations in wasmtime require a contextual `store`
	// argument to share, so create that first
	store := wasmtime.NewStore(wasmtime.NewEngine())

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (import "" "hello" (func $hello))
	  (func (export "run")
	    (call $hello)
	  )
	)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}

	// Our `hello.wat` file imports one item, so we create that function
	// here.
	item := wasmtime.WrapFunc(store, func() {
		fmt.Println("Hello from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{item})
	if err != nil {
		log.Fatal(err)
	}

	// After we've instantiated we can lookup our `run` function and call
	// it.
	run := instance.GetFunc(store, "run")
	_, err = run.Call(store)
	if err != nil {
		log.Fatal(err)
	}
	// Output: Hello from Go!
}

// An example of a wasm module which calculates the GCD of two numbers
func Example_gcd() {
	store := wasmtime.NewStore(wasmtime.NewEngine())
	wasm, err := wasmtime.Wat2Wasm(`
	(module
	  (func $gcd (param i32 i32) (result i32)
	    (local i32)
	    block  ;; label = @1
	      block  ;; label = @2
	        local.get 0
	        br_if 0 (;@2;)
	        local.get 1
	        local.set 2
	        br 1 (;@1;)
	      end
	      loop  ;; label = @2
	        local.get 1
	        local.get 0
	        local.tee 2
	        i32.rem_u
	        local.set 0
	        local.get 2
	        local.set 1
	        local.get 0
	        br_if 0 (;@2;)
	      end
	    end
	    local.get 2
	  )
	  (export "gcd" (func $gcd))
	)
	`)
	if err != nil {
		log.Fatal(err)
	}
	module, err := wasmtime.NewModule(store.Engine, wasm)
	if err != nil {
		log.Fatal(err)
	}
	instance, err := wasmtime.NewInstance(store, module, []wasmtime.AsExtern{})
	if err != nil {
		log.Fatal(err)
	}
	run := instance.GetFunc(store, "gcd")
	result, err := run.Call(store, 6, 27)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("gcd(6, 27) = %d\n", result.(int32))
	// Output: gcd(6, 27) = 3
}
