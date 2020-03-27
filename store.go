package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type Store struct {
	_ptr *C.wasm_store_t
}

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
