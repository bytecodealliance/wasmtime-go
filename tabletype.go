package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type TableType struct {
	_ptr  *C.wasm_tabletype_t
	owner interface{}
}

// Creates a new `TableType` with the `kind` provided and whether it's
// `mutable` or not
func NewTableType(element *ValType, limits Limits) *TableType {
	valptr := C.wasm_valtype_new(C.wasm_valtype_kind(element.ptr()))
	runtime.KeepAlive(element)
	limits_ffi := limits.ffi()
	ptr := C.wasm_tabletype_new(valptr, &limits_ffi)

	return mkTableType(ptr, nil)
}

func mkTableType(ptr *C.wasm_tabletype_t, owner interface{}) *TableType {
	tabletype := &TableType{_ptr: ptr, owner: owner}
	if owner == nil {
		runtime.SetFinalizer(tabletype, func(tabletype *TableType) {
			C.wasm_tabletype_delete(tabletype._ptr)
		})
	}
	return tabletype
}

func (ty *TableType) ptr() *C.wasm_tabletype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

// Returns the type of value stored in this table
func (ty *TableType) Element() *ValType {
	ptr := C.wasm_tabletype_element(ty.ptr())
	return mkValType(ptr, ty)
}

// Returns limits on the size of this table type
func (ty *TableType) Limits() Limits {
	ptr := C.wasm_tabletype_limits(ty.ptr())
	return mkLimits(ptr, ty)
}
