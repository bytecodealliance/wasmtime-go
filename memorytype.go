package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type MemoryType struct {
	_ptr  *C.wasm_memorytype_t
	owner interface{}
}

// Creates a new `MemoryType` with the `kind` provided and whether it's
// `mumemory` or not
func NewMemoryType(limits Limits) *MemoryType {
	limits_ffi := limits.ffi()
	ptr := C.wasm_memorytype_new(&limits_ffi)

	return mkMemoryType(ptr, nil)
}

func mkMemoryType(ptr *C.wasm_memorytype_t, owner interface{}) *MemoryType {
	memorytype := &MemoryType{_ptr: ptr, owner: owner}
	if owner == nil {
		runtime.SetFinalizer(memorytype, func(memorytype *MemoryType) {
			C.wasm_memorytype_delete(memorytype._ptr)
		})
	}
	return memorytype
}

func (ty *MemoryType) ptr() *C.wasm_memorytype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

// Returns the limits on the size of this memory type
func (ty *MemoryType) Limits() Limits {
	ptr := C.wasm_memorytype_limits(ty.ptr())
	return mkLimits(ptr, ty)
}
