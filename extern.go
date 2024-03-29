package wasmtime

// #include "shims.h"
import "C"
import "runtime"

// Extern is an external value, which is the runtime representation of an entity that can be imported or exported.
// It is an address denoting either a function instance, table instance, memory instance, or global instances in the shared store.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#external-values)
type Extern struct {
	_ptr *C.wasmtime_extern_t
}

// AsExtern is an interface for all types which can be imported or exported as an Extern
type AsExtern interface {
	AsExtern() C.wasmtime_extern_t
}

func mkExtern(ptr *C.wasmtime_extern_t) *Extern {
	f := &Extern{_ptr: ptr}
	runtime.SetFinalizer(f, func(e *Extern) {
		e.Close()
	})
	return f
}

func (e *Extern) ptr() *C.wasmtime_extern_t {
	ret := e._ptr
	if ret == nil {
		panic("object already closed")
	}
	maybeGC()
	return ret
}

// Close will deallocate this extern's state explicitly.
//
// For more information see the documentation for engine.Close()
func (e *Extern) Close() {
	if e._ptr == nil {
		return
	}
	runtime.SetFinalizer(e, nil)
	C.wasmtime_extern_delete(e._ptr)
	e._ptr = nil

}

// Type returns the type of this export
func (e *Extern) Type(store Storelike) *ExternType {
	ptr := C.wasmtime_extern_type(store.Context(), e.ptr())
	runtime.KeepAlive(e)
	runtime.KeepAlive(store)
	return mkExternType(ptr, nil)
}

// Func returns a Func if this export is a function or nil otherwise
func (e *Extern) Func() *Func {
	ptr := e.ptr()
	if ptr.kind != C.WASMTIME_EXTERN_FUNC {
		return nil
	}
	ret := mkFunc(C.go_wasmtime_extern_func_get(ptr))
	runtime.KeepAlive(e)
	return ret
}

// Global returns a Global if this export is a global or nil otherwise
func (e *Extern) Global() *Global {
	ptr := e.ptr()
	if ptr.kind != C.WASMTIME_EXTERN_GLOBAL {
		return nil
	}
	ret := mkGlobal(C.go_wasmtime_extern_global_get(ptr))
	runtime.KeepAlive(e)
	return ret
}

// Memory returns a Memory if this export is a memory or nil otherwise
func (e *Extern) Memory() *Memory {
	ptr := e.ptr()
	if ptr.kind != C.WASMTIME_EXTERN_MEMORY {
		return nil
	}
	ret := mkMemory(C.go_wasmtime_extern_memory_get(ptr))
	runtime.KeepAlive(e)
	return ret
}

// Table returns a Table if this export is a table or nil otherwise
func (e *Extern) Table() *Table {
	ptr := e.ptr()
	if ptr.kind != C.WASMTIME_EXTERN_TABLE {
		return nil
	}
	ret := mkTable(C.go_wasmtime_extern_table_get(ptr))
	runtime.KeepAlive(e)
	return ret
}

func (e *Extern) AsExtern() C.wasmtime_extern_t {
	return *e.ptr()
}
