package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type MemoryType struct {
	_ptr   *C.wasm_memorytype_t
	_owner interface{}
}

// Creates a new `MemoryType` with the `limits` on size provided
func NewMemoryType(limits Limits) *MemoryType {
	limits_ffi := limits.ffi()
	ptr := C.wasm_memorytype_new(&limits_ffi)
	return mkMemoryType(ptr, nil)
}

func mkMemoryType(ptr *C.wasm_memorytype_t, owner interface{}) *MemoryType {
	memorytype := &MemoryType{_ptr: ptr, _owner: owner}
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

func (ty *MemoryType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Returns the limits on the size of this memory type
func (ty *MemoryType) Limits() Limits {
	ptr := C.wasm_memorytype_limits(ty.ptr())
	return mkLimits(ptr, ty.owner())
}

// Converts this type to an instance of `ExternType`
func (ty *MemoryType) AsExternType() *ExternType {
	ptr := C.wasm_memorytype_as_externtype_const(ty.ptr())
	return mkExternType(ptr, ty.owner())
}
