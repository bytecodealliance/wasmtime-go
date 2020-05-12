package wasmtime

// #include <wasm.h>
import "C"
import "runtime"
import "unsafe"

type Memory struct {
	_ptr     *C.wasm_memory_t
	freelist *freeList
	_owner   interface{}
}

// Creates a new `Memory` in the given `Store` with the specified `ty`.
func NewMemory(store *Store, ty *MemoryType) *Memory {
	ptr := C.wasm_memory_new(store.ptr(), ty.ptr())
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	return mkMemory(ptr, store.freelist, nil)
}

func mkMemory(ptr *C.wasm_memory_t, freelist *freeList, owner interface{}) *Memory {
	f := &Memory{_ptr: ptr, _owner: owner, freelist: freelist}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Memory) {
			f.freelist.lock.Lock()
			defer f.freelist.lock.Unlock()
			f.freelist.memories = append(f.freelist.memories, f._ptr)
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

// Returns the raw memory backed by this `Memory` as a byte slice (`[]byte`).
//
// This is not a safe method to call, hence the "unsafe" in the name. The byte
// slice returned from this function is not managed by the Go garbage collector.
// You need to ensure that `m`, the original `Memory`, lives longer than the
// `[]byte` returned.
//
// Note that you may need to use `runtime.KeepAlive` to keep the original memory
// `m` alive for long enough while you're using the `[]byte` slice. If the
// `[]byte` slice is used after `m` is GC'd then that is undefined behavior.
func (m *Memory) UnsafeData() []byte {
	// see https://github.com/golang/go/wiki/cgo#turning-c-arrays-into-go-slices
	const MAX_LEN = 1 << 32
	length := m.DataSize()
	if length >= MAX_LEN {
		panic("memory is too big")
	}
	return (*[MAX_LEN]byte)(m.Data())[:length:length]
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
	return mkExtern(ptr, m.freelist, m.owner())
}
