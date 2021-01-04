package wasmtime

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
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
	check(err)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := NewModule(store.Engine, wasm)
	check(err)

	// Our `hello.wat` file imports one item, so we create that function
	// here.
	item := WrapFunc(store, func() {
		fmt.Println("Hello from Go!")
	})

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err := NewInstance(store, module, []*Extern{item.AsExtern()})
	check(err)

	// After we've instantiated we can lookup our `run` function and call
	// it.
	run := instance.GetExport("run").Func()
	_, err = run.Call()
	check(err)

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
	check(err)
	module, err := NewModule(store.Engine, wasm)
	check(err)
	instance, err := NewInstance(store, module, []*Extern{})
	check(err)
	run := instance.GetExport("gcd").Func()
	result, err := run.Call(6, 27)
	check(err)
	fmt.Printf("gcd(6, 27) = %d\n", result.(int32))
	// Output: gcd(6, 27) = 3
}

// An example of working with the Memory type to read/write wasm memory.
func Example_memory() {
	// Create our `Store` context and then compile a module and create an
	// instance from the compiled module all in one go.
	wasmtimeStore := NewStore(NewEngine())
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
	check(err)
	module, err := NewModule(wasmtimeStore.Engine, wasm)
	check(err)
	instance, err := NewInstance(wasmtimeStore, module, []*Extern{})
	check(err)

	// Load up our exports from the instance
	memory := instance.GetExport("memory").Memory()
	size := instance.GetExport("size").Func()
	load := instance.GetExport("load").Func()
	store := instance.GetExport("store").Func()

	// some helper functions we'll use below
	call32 := func(f *Func, args ...interface{}) int32 {
		ret, err := f.Call(args...)
		check(err)
		return ret.(int32)
	}
	call := func(f *Func, args ...interface{}) {
		_, err := f.Call(args...)
		check(err)
	}
	assertTraps := func(f *Func, args ...interface{}) {
		_, err := f.Call(args...)
		_, ok := err.(*Trap)
		if !ok {
			panic("expected a trap")
		}
	}

	// Check the initial memory sizes/contents
	assert(memory.Size() == 2)
	assert(memory.DataSize() == 0x20000)
	buf := memory.UnsafeData()

	assert(buf[0] == 0)
	assert(buf[0x1000] == 1)
	assert(buf[0x1003] == 4)

	assert(call32(size) == 2)
	assert(call32(load, 0) == 0)
	assert(call32(load, 0x1000) == 1)
	assert(call32(load, 0x1003) == 4)
	assert(call32(load, 0x1ffff) == 0)
	assertTraps(load, 0x20000)

	// We can mutate memory as well
	buf[0x1003] = 5
	call(store, 0x1002, 6)
	assertTraps(store, 0x20000, 0)

	assert(buf[0x1002] == 6)
	assert(buf[0x1003] == 5)
	assert(call32(load, 0x1002) == 6)
	assert(call32(load, 0x1003) == 5)

	// And like wasm instructions, we can grow memory
	assert(memory.Grow(1))
	assert(memory.Size() == 3)
	assert(memory.DataSize() == 0x30000)

	assert(call32(load, 0x20000) == 0)
	call(store, 0x20000, 0)
	assertTraps(load, 0x30000)
	assertTraps(store, 0x30000, 0)

	// Memory can fail to grow
	assert(!memory.Grow(1))
	assert(memory.Grow(0))

	// Ensure that `memory` lives long enough to cover all our usages of
	// using its internal buffer we read from `UnsafeData()`
	runtime.KeepAlive(memory)

	// Finally we can also create standalone memories to get imported by
	// wasm modules too.
	memorytype := NewMemoryType(Limits{Min: 5, Max: 5})
	memory2 := NewMemory(wasmtimeStore, memorytype)
	assert(memory2.Size() == 5)
	assert(!memory2.Grow(1))
	assert(memory2.Grow(0))

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
	check(err)
	module, err := NewModule(store.Engine, wasm)
	check(err)

	callback := WrapFunc(store, func(a int32, b int64) (int64, int32) {
		return b + 1, a + 1
	})

	instance, err := NewInstance(store, module, []*Extern{callback.AsExtern()})
	check(err)

	g := instance.GetExport("g").Func()

	results, err := g.Call(1, 3)
	check(err)
	arr := results.([]Val)
	a := arr[0].I64()
	b := arr[1].I32()
	fmt.Printf("> %d %d\n", a, b)

	assert(a == 4)
	assert(b == 2)

	roundTripMany := instance.GetExport("round_trip_many").Func()
	results, err = roundTripMany.Call(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	check(err)
	arr = results.([]Val)

	for i := 0; i < len(arr); i++ {
		fmt.Printf(" %d", arr[i].Get())
		assert(arr[i].I64() == int64(i))
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
	check(err)
	defer os.RemoveAll(dir)
	stdoutPath := filepath.Join(dir, "stdout")

	engine := NewEngine()
	store := NewStore(engine)

	linker := NewLinker(store)

	// Configure WASI imports to write stdout into a file.
	wasiConfig := NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutPath)

	// Set the version to the same as in the WAT.
	wasi, err := NewWasiInstance(store, wasiConfig, "wasi_snapshot_preview1")
	check(err)

	// Link WASI
	err = linker.DefineWasi(wasi)
	check(err)

	// Create our module
	wasm, err := Wat2Wasm(TextWat)
	check(err)
	module, err := NewModule(store.Engine, wasm)
	check(err)
	instance, err := linker.Instantiate(module)
	check(err)

	// Run the function
	nom := instance.GetExport("_start").Func()
	_, err = nom.Call()
	check(err)

	// Print WASM stdout
	out, err := ioutil.ReadFile(stdoutPath)
	check(err)
	fmt.Print(string(out))

	// Output: hello world
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
