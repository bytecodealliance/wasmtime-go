package wasmtime

// #cgo LDFLAGS:-lwasmtime
// #include <wasm.h>
// #include <wasi.h>
// #include <wasmtime.h>
import "C"

func Foo() {
  C.wasm_engine_new()
}
