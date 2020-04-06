package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

// Enumeration of different kinds of value types
type ValKind C.wasm_valkind_t

const (
	KindI32     ValKind = C.WASM_I32
	KindI64     ValKind = C.WASM_I64
	KindF32     ValKind = C.WASM_F32
	KindF64     ValKind = C.WASM_F64
	KindAnyref  ValKind = C.WASM_ANYREF
	KindFuncref ValKind = C.WASM_FUNCREF
)

// Renders this kind as a string, similar to the `*.wat` format
func (ty ValKind) String() string {
	switch ty {
	case KindI32:
		return "i32"
	case KindI64:
		return "i64"
	case KindF32:
		return "f32"
	case KindF64:
		return "f64"
	case KindAnyref:
		return "anyref"
	case KindFuncref:
		return "funcref"
	}
	panic("unknown kind")
}

type ValType struct {
	_ptr   *C.wasm_valtype_t
	_owner interface{}
}

// Creates a new `ValType` with the `kind` provided
func NewValType(kind ValKind) *ValType {
	ptr := C.wasm_valtype_new(C.wasm_valkind_t(kind))
	return mkValType(ptr, nil)
}

func mkValType(ptr *C.wasm_valtype_t, owner interface{}) *ValType {
	valtype := &ValType{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(valtype, func(valtype *ValType) {
			C.wasm_valtype_delete(valtype._ptr)
		})
	}
	return valtype
}

// Returns the corresponding `ValKind` for this `ValType`
func (ty *ValType) Kind() ValKind {
	ret := ValKind(C.wasm_valtype_kind(ty.ptr()))
	runtime.KeepAlive(ty)
	return ret
}

// Converts this `ValType` into a string according to the string representation
// of `ValKind`.
func (ty *ValType) String() string {
	return ty.Kind().String()
}

func (valtype *ValType) ptr() *C.wasm_valtype_t {
	ret := valtype._ptr
	maybeGC()
	return ret
}

func (ty *ValType) owner() interface{} {
	if ty._owner != nil {
		return ty._owner
	}
	return ty
}
