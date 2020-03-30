package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type ExternType struct {
	_ptr   *C.wasm_externtype_t
	_owner interface{}
}

type AsExtern interface {
	AsExtern() *ExternType
}

func mkExternType(ptr *C.wasm_externtype_t, owner interface{}) *ExternType {
	externtype := &ExternType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(externtype, func(externtype *ExternType) {
			C.wasm_externtype_delete(externtype._ptr)
		})
	}
	return externtype
}

func (ty *ExternType) ptr() *C.wasm_externtype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

func (ty *ExternType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}

// Returns the underlying `FuncType` for this `ExternType` if it's a function
// type. Otherwise returns `nil`.
func (ty *ExternType) FuncType() *FuncType {
	ptr := C.wasm_externtype_as_functype(ty.ptr())
	if ptr == nil {
		return nil
	}
	return mkFuncType(ptr, ty.owner())
}

// Returns the underlying `GlobalType` for this `ExternType` if it's a function
// type. Otherwise returns `nil`.
func (ty *ExternType) GlobalType() *GlobalType {
	ptr := C.wasm_externtype_as_globaltype(ty.ptr())
	if ptr == nil {
		return nil
	}
	return mkGlobalType(ptr, ty.owner())
}

// Returns the underlying `TableType` for this `ExternType` if it's a function
// type. Otherwise returns `nil`.
func (ty *ExternType) TableType() *TableType {
	ptr := C.wasm_externtype_as_tabletype(ty.ptr())
	if ptr == nil {
		return nil
	}
	return mkTableType(ptr, ty.owner())
}

// Returns the underlying `MemoryType` for this `ExternType` if it's a function
// type. Otherwise returns `nil`.
func (ty *ExternType) MemoryType() *MemoryType {
	ptr := C.wasm_externtype_as_memorytype(ty.ptr())
	if ptr == nil {
		return nil
	}
	return mkMemoryType(ptr, ty.owner())
}

// Returns this type itself
func (ty *ExternType) AsExtern() *ExternType {
	return ty
}
