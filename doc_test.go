package wasmtime

import (
	"fmt"
)

// An example of instantiating a small wasm module which imports functionality
// from the host, then calling into wasm which calls back into the host.
func Example() {
	// Almost all operations in wasmtime require a contextual `store`
	// argument to share, so create that first
	store := NewStore(NewEngine())

	// Compiling modules requires WebAssembly binary input, but the wasmtime
	// package also supports converting the WebAssembly text format to the
	// binary format.
	wasm, err := Wat2Wasm(`
	  (module
	  (import "" "hello" (func $hello))
	  (func (export "run")
	    (call $hello))
	  )
      `)
	exampleCheckErr(err)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := NewModule(store.Engine, wasm)
	exampleCheckErr(err)

	// Our `hello.wat` file imports one item, so we create that function
	// here.
	item := WrapFunc(store, func() {
		fmt.Println("Hello from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := NewInstance(store, module, []AsExtern{item})
	exampleCheckErr(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	run := instance.GetFunc(store, "run")
	_, err = run.Call(store)
	exampleCheckErr(err)

	// Output: Hello from Go!
}

const GcdWat = `
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
`

// An example of a wasm module which calculates the GCD of two numbers
func Example_gcd() {
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(GcdWat)
	exampleCheckErr(err)
	module, err := NewModule(store.Engine, wasm)
	exampleCheckErr(err)
	instance, err := NewInstance(store, module, []AsExtern{})
	exampleCheckErr(err)
	run := instance.GetFunc(store, "gcd")
	result, err := run.Call(store, 6, 27)
	exampleCheckErr(err)
	fmt.Printf("gcd(6, 27) = %d\n", result.(int32))
	// Output: gcd(6, 27) = 3
}

<<<<<<< HEAD
// An example of working with the Memory type to read/write wasm memory.
func ExampleMemory() {
	// Create our `Store` context and then compile a module and create an
	// instance from the compiled module all in one go.
	store := NewStore(NewEngine())
	wasm, err := Wat2Wasm(`
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
	exampleCheckErr(err)
	module, err := NewModule(store.Engine, wasm)
	exampleCheckErr(err)
	instance, err := NewInstance(store, module, []AsExtern{})
	exampleCheckErr(err)

	// Load up our exports from the instance
	memory := instance.GetExport(store, "memory").Memory()
	sizeFn := instance.GetFunc(store, "size")
	loadFn := instance.GetFunc(store, "load")
	storeFn := instance.GetFunc(store, "store")

	// some helper functions we'll use below
	call32 := func(f *Func, args ...interface{}) int32 {
		ret, err := f.Call(store, args...)
		exampleCheckErr(err)
		return ret.(int32)
	}
	call := func(f *Func, args ...interface{}) {
		_, err := f.Call(store, args...)
		exampleCheckErr(err)
	}
	assertTraps := func(f *Func, args ...interface{}) {
		_, err := f.Call(store, args...)
		_, ok := err.(*Trap)
		if !ok {
			panic("expected a trap")
		}
	}

	// Check the initial memory sizes/contents
	exampleAssert(memory.Size(store) == 2)
	exampleAssert(memory.DataSize(store) == 0x20000)
	buf := memory.UnsafeData(store)

	exampleAssert(buf[0] == 0)
	exampleAssert(buf[0x1000] == 1)
	exampleAssert(buf[0x1003] == 4)

	exampleAssert(call32(sizeFn) == 2)
	exampleAssert(call32(loadFn, 0) == 0)
	exampleAssert(call32(loadFn, 0x1000) == 1)
	exampleAssert(call32(loadFn, 0x1003) == 4)
	exampleAssert(call32(loadFn, 0x1ffff) == 0)
	assertTraps(loadFn, 0x20000)

	// We can mutate memory as well
	buf[0x1003] = 5
	call(storeFn, 0x1002, 6)
	assertTraps(storeFn, 0x20000, 0)

	exampleAssert(buf[0x1002] == 6)
	exampleAssert(buf[0x1003] == 5)
	exampleAssert(call32(loadFn, 0x1002) == 6)
	exampleAssert(call32(loadFn, 0x1003) == 5)

	// And like wasm instructions, we can grow memory
	_, err = memory.Grow(store, 1)
	exampleAssert(err == nil)
	exampleAssert(memory.Size(store) == 3)
	exampleAssert(memory.DataSize(store) == 0x30000)

	exampleAssert(call32(loadFn, 0x20000) == 0)
	call(storeFn, 0x20000, 0)
	assertTraps(loadFn, 0x30000)
	assertTraps(storeFn, 0x30000, 0)

	// Memory can fail to grow
	_, err = memory.Grow(store, 1)
	exampleAssert(err != nil)
	_, err = memory.Grow(store, 0)
	exampleAssert(err == nil)

	// Ensure that `memory` lives long enough to cover all our usages of
	// using its internal buffer we read from `UnsafeData()`
	runtime.KeepAlive(memory)

	// Finally we can also create standalone memories to get imported by
	// wasm modules too.
	memorytype := NewMemoryType(5, true, 5)
	memory2, err := NewMemory(store, memorytype)
	exampleAssert(err == nil)
	exampleAssert(memory2.Size(store) == 5)
	_, err = memory2.Grow(store, 1)
	exampleAssert(err != nil)
	_, err = memory2.Grow(store, 0)
	exampleAssert(err == nil)

	// Output:
}

// An example of enabling the multi-value feature of WebAssembly and
// interacting with multi-value functions.
func Example_multi() {
	// Configure our `Store`, but be sure to use a `Config` that enables the
	// wasm multi-value feature since it's not stable yet.
	config := NewConfig()
	config.SetWasmMultiValue(true)
	store := NewStore(NewEngineWithConfig(config))

	wasm, err := Wat2Wasm(`
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
	      local.get 9)
	  )
        `)
	exampleCheckErr(err)
	module, err := NewModule(store.Engine, wasm)
	exampleCheckErr(err)

	callback := WrapFunc(store, func(a int32, b int64) (int64, int32) {
		return b + 1, a + 1
	})

	instance, err := NewInstance(store, module, []AsExtern{callback})
	exampleCheckErr(err)

	g := instance.GetFunc(store, "g")

	results, err := g.Call(store, 1, 3)
	exampleCheckErr(err)
	arr := results.([]Val)
	a := arr[0].I64()
	b := arr[1].I32()
	fmt.Printf("> %d %d\n", a, b)

	exampleAssert(a == 4)
	exampleAssert(b == 2)

	roundTripMany := instance.GetFunc(store, "round_trip_many")
	results, err = roundTripMany.Call(store, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	exampleCheckErr(err)
	arr = results.([]Val)

	for i := 0; i < len(arr); i++ {
		fmt.Printf(" %d", arr[i].Get())
		exampleAssert(arr[i].I64() == int64(i))
	}

	// Output: > 4 2
	//  0 1 2 3 4 5 6 7 8 9
}

const TextWat = `
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
`

// An example of linking WASI to the runtime in order to interact with the system.
// It uses the WAT code from https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-tutorial.md#web-assembly-text-example
func Example_wasi() {
	dir, err := ioutil.TempDir("", "out")
	exampleCheckErr(err)
	defer os.RemoveAll(dir)
	stdoutPath := filepath.Join(dir, "stdout")

	engine := NewEngine()

	// Create our module
	wasm, err := Wat2Wasm(TextWat)
	exampleCheckErr(err)
	module, err := NewModule(engine, wasm)
	exampleCheckErr(err)

	// Create a linker with WASI functions defined within it
	linker := NewLinker(engine)
	err = linker.DefineWasi()
	exampleCheckErr(err)

	// Configure WASI imports to write stdout into a file, and then create
	// a `Store` using this wasi configuration.
	wasiConfig := NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutPath)
	store := NewStore(engine)
	store.SetWasi(wasiConfig)
	instance, err := linker.Instantiate(store, module)
	exampleCheckErr(err)

	// Run the function
	nom := instance.GetFunc(store, "_start")
	_, err = nom.Call(store)
	exampleCheckErr(err)

	// Print WASM stdout
	out, err := ioutil.ReadFile(stdoutPath)
	exampleCheckErr(err)
	fmt.Print(string(out))

	// Output: hello world
}

func exampleCheckErr(e error) {
	if e != nil {
		panic(e)
	}
}

func exampleAssert(b bool) {
	if !b {
		panic("assertion failed")
	}
}
