package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

// A `Store` is a general group of wasm instances, and many objects
// must all be created with and reference the same `Store`
type Store struct {
	_ptr *C.wasm_store_t
}

// Creates a new `Store` from the configuration provided in `engine`
func NewStore(engine *Engine) *Store {
	store := &Store{_ptr: C.wasm_store_new(engine.ptr())}
	runtime.KeepAlive(engine)
	runtime.SetFinalizer(store, func(store *Store) {
		C.wasm_store_delete(store._ptr)
	})
	return store
}

func (store *Store) ptr() *C.wasm_store_t {
	ret := store._ptr
	maybeGC()
	return ret
}
