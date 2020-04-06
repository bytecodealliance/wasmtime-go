package wasmtime

// #include <wasm.h>
import "C"
import "runtime"
import "unsafe"

type Memory struct {
	_ptr   *C.wasm_memory_t
	_owner interface{}
}

// Creates a new `Memory` in the given `Store` with the specified `ty`.
func NewMemory(store *Store, ty *MemoryType) *Memory {
	ptr := C.wasm_memory_new(store.ptr(), ty.ptr())
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	return mkMemory(ptr, nil)
}

func mkMemory(ptr *C.wasm_memory_t, owner interface{}) *Memory {
	f := &Memory{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Memory) {
			C.wasm_memory_delete(f._ptr)
		})
	}
	return f
}

func (f *Memory) ptr() *C.wasm_memory_t {
	ret := f._ptr
	maybeGC()
	return ret
}

func (f *Memory) owner() interface{} {
	if f._owner != nil {
		return f._owner
	}
	return f
}

// Returns the type of this memory
func (m *Memory) Type() *MemoryType {
	ptr := C.wasm_memory_type(m.ptr())
	runtime.KeepAlive(m)
	return mkMemoryType(ptr, nil)
}

// Returns the raw pointer in memory of where this memory starts
func (m *Memory) Data() unsafe.Pointer {
	ret := unsafe.Pointer(C.wasm_memory_data(m.ptr()))
	runtime.KeepAlive(m)
	return ret
}

// Returns the size, in bytes, that `Data()` is valid for
func (m *Memory) DataSize() uintptr {
	ret := uintptr(C.wasm_memory_data_size(m.ptr()))
	runtime.KeepAlive(m)
	return ret
}

// Returns the size, in wasm pages, of this memory
func (m *Memory) Size() uint32 {
	ret := uint32(C.wasm_memory_size(m.ptr()))
	runtime.KeepAlive(m)
	return ret
}

// Grows this memory by `delta` pages
func (m *Memory) Grow(delta uint) bool {
	ret := C.wasm_memory_grow(m.ptr(), C.wasm_memory_pages_t(delta))
	runtime.KeepAlive(m)
	return bool(ret)
}

func (m *Memory) AsExtern() *Extern {
	ptr := C.wasm_memory_as_extern(m.ptr())
	return mkExtern(ptr, m.owner())
}
