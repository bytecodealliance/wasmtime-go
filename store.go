package wasmtime

// #include <wasm.h>
// #include <wasmtime.h>
import "C"
import (
	"errors"
	"runtime"
)

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

func (store *Store) InterruptHandle() (*InterruptHandle, error) {
	ptr := C.wasmtime_interrupt_handle_new(store.ptr())
	runtime.KeepAlive(store)
	if ptr == nil {
		return nil, errors.New("interrupts not enabled in `Config`")
	}

	handle := &InterruptHandle{_ptr: ptr}
	runtime.SetFinalizer(handle, func(handle *InterruptHandle) {
		C.wasmtime_interrupt_handle_delete(handle._ptr)
	})
	return handle, nil
}

func (store *Store) ptr() *C.wasm_store_t {
	ret := store._ptr
	maybeGC()
	return ret
}

// An `InterruptHandle` is used to interrupt the execution of currently running
// wasm code.
//
// For more information see
// https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Store.html#method.interrupt_handle
type InterruptHandle struct {
	_ptr *C.wasmtime_interrupt_handle_t
}

// Interrupts currently executing WebAssembly code, if it's currently running,
// or interrupts wasm the next time it starts running.
//
// For more information see
// https://bytecodealliance.github.io/wasmtime/api/wasmtime/struct.Store.html#method.interrupt_handle
func (i *InterruptHandle) Interrupt() {
	C.wasmtime_interrupt_handle_interrupt(i.ptr())
	runtime.KeepAlive(i)
}

func (i *InterruptHandle) ptr() *C.wasmtime_interrupt_handle_t {
	ret := i._ptr
	maybeGC()
	return ret
}
