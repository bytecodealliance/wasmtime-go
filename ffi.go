package wasmtime

// #cgo LDFLAGS:-lwasmtime
// #cgo windows CFLAGS:-DWASM_API_EXTERN=
// #include <wasm.h>
// #include <wasi.h>
// #include <wasmtime.h>
import "C"

func Foo() {
  C.wasm_engine_new()
}
