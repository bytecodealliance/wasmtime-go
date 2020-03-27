package wasmtime

// #cgo CFLAGS:-Iwasmtime/include
// #cgo LDFLAGS:-lwasmtime -Lwasmtime/lib -lm -ldl
// #include <wasm.h>
// #include <wasi.h>
// #include <wasmtime.h>
import "C"

func Foo() {
  C.wasm_engine_new()
}
