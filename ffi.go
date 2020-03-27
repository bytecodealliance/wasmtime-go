package wasmtime

// #cgo LDFLAGS:-lwasmtime
// #cgo windows CFLAGS:-DWASM_EXTERN_API
// #include <wasm.h>
// #include <wasi.h>
// #include <wasmtime.h>
import "C"

func Foo() {
  C.wasm_engine_new()
}
