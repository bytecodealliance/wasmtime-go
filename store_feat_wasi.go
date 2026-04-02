package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"

// SetWasi will configure the WASI state to use for instances within this
// `Store`.
//
// The `wasi` argument cannot be reused for another `Store`, it's consumed by
// this function.
func (store *Store) SetWasi(wasi *WasiConfig) {
	runtime.SetFinalizer(wasi, nil)
	ptr := wasi.ptr()
	wasi._ptr = nil
	C.wasmtime_context_set_wasi(store.Context(), ptr)
	runtime.KeepAlive(store)
}
