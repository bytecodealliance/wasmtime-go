package wasmtime_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bytecodealliance/wasmtime-go"
)

// An example of linking WASI to the runtime in order to interact with the system.
// It uses the WAT code from https://github.com/bytecodealliance/wasmtime/blob/main/docs/WASI-tutorial.md#web-assembly-text-example
func ExampleWasiConfig() {
	dir, err := ioutil.TempDir("", "out")
	if err != nil {
		panic(err)
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
		panic(err)
	}
	module, err := wasmtime.NewModule(engine, wasm)
	if err != nil {
		panic(err)
	}

	// Create a linker with WASI functions defined within it
	linker := wasmtime.NewLinker(engine)
	err = linker.DefineWasi()
	if err != nil {
		panic(err)
	}

	// Configure WASI imports to write stdout into a file, and then create
	// a `Store` using this wasi configuration.
	wasiConfig := wasmtime.NewWasiConfig()
	wasiConfig.SetStdoutFile(stdoutPath)
	store := wasmtime.NewStore(engine)
	store.SetWasi(wasiConfig)
	instance, err := linker.Instantiate(store, module)
	if err != nil {
		panic(err)
	}

	// Run the function
	nom := instance.GetFunc(store, "_start")
	_, err = nom.Call(store)
	if err != nil {
		panic(err)
	}

	// Print WASM stdout
	out, err := ioutil.ReadFile(stdoutPath)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(out))
	// Output: hello world
}
