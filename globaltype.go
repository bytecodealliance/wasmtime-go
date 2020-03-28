package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type GlobalType struct {
	_ptr  *C.wasm_globaltype_t
	owner interface{}
}

// Creates a new `GlobalType` with the `kind` provided and whether it's
// `mutable` or not
func NewGlobalType(content *ValType, mutable bool) *GlobalType {
	mutability := C.WASM_CONST
	if mutable {
		mutability = C.WASM_VAR
	}
        content_ptr := C.wasm_valtype_new(C.wasm_valtype_kind(content.ptr()))
	runtime.KeepAlive(content)
	ptr := C.wasm_globaltype_new(content_ptr, C.wasm_mutability_t(mutability))

	return mkGlobalType(ptr, nil)
}

func mkGlobalType(ptr *C.wasm_globaltype_t, owner interface{}) *GlobalType {
	globaltype := &GlobalType{_ptr: ptr, owner: owner}
	if owner == nil {
		runtime.SetFinalizer(globaltype, func(globaltype *GlobalType) {
			C.wasm_globaltype_delete(globaltype._ptr)
		})
	}
	return globaltype
}

func (ty *GlobalType) ptr() *C.wasm_globaltype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

// Returns the type of value stored in this global
func (ty *GlobalType) Content() *ValType {
	ptr := C.wasm_globaltype_content(ty.ptr())
	return mkValType(ptr, ty)
}

// Returns whether this global type is mutable or not
func (ty *GlobalType) Mutable() bool {
	ret := C.wasm_globaltype_mutability(ty.ptr()) == C.WASM_VAR
	runtime.KeepAlive(ty)
	return ret
}
