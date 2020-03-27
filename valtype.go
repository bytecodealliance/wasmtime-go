package wasmtime

// #include <wasm.h>
import "C"
import "runtime"

type ValKind int

const (
	KindI32     ValKind = C.WASM_I32
	KindI64     ValKind = C.WASM_I64
	KindF32     ValKind = C.WASM_F32
	KindF64     ValKind = C.WASM_F64
	KindAnyref  ValKind = C.WASM_ANYREF
	KindFuncref ValKind = C.WASM_FUNCREF
)

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
	_ptr *C.wasm_valtype_t
}

func NewValType(kind ValKind) *ValType {
	valtype := &ValType{_ptr: C.wasm_valtype_new(C.wasm_valkind_t(kind))}
	runtime.SetFinalizer(valtype, func(valtype *ValType) {
		C.wasm_valtype_delete(valtype._ptr)
	})
	return valtype
}

func (ty *ValType) Kind() ValKind {
	ret := ValKind(C.wasm_valtype_kind(ty.ptr()))
	runtime.KeepAlive(ty)
	return ret
}

func (ty *ValType) String() string {
	return ty.Kind().String()
}

func (valtype *ValType) ptr() *C.wasm_valtype_t {
	ret := valtype._ptr
	maybeGC()
	return ret
}
