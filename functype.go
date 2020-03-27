package wasmtime

// #include <wasm.h>
import "C"
import "runtime"
import "unsafe"

type FuncType struct {
	_ptr  *C.wasm_functype_t
	owner interface{}
}

// Creates a new `FuncType` with the `kind` provided
func NewFuncType(params, results []*ValType) *FuncType {
	param_vec := mkValTypeList(params)
	result_vec := mkValTypeList(results)

	ptr := C.wasm_functype_new(&param_vec, &result_vec)

	return mkFuncType(ptr, nil)
}

func mkValTypeList(tys []*ValType) C.wasm_valtype_vec_t {
	vec := C.wasm_valtype_vec_t{}
	C.wasm_valtype_vec_new_uninitialized(&vec, C.size_t(len(tys)))
	base := unsafe.Pointer(vec.data)
	for i, ty := range tys {
		ptr := C.wasm_valtype_new(C.wasm_valtype_kind(ty.ptr()))
		*(**C.wasm_valtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i))) = ptr
	}
	runtime.KeepAlive(tys)
	return vec
}

func mkFuncType(ptr *C.wasm_functype_t, owner interface{}) *FuncType {
	functype := &FuncType{_ptr: ptr, owner: owner}
	if owner != nil {
		runtime.SetFinalizer(functype, func(functype *FuncType) {
			C.wasm_functype_delete(functype._ptr)
		})
	}
	return functype
}

func (ty *FuncType) ptr() *C.wasm_functype_t {
	ret := ty._ptr
	maybeGC()
	return ret
}

// Returns the parameter types of this function type
func (ty *FuncType) Params() []*ValType {
	ptr := C.wasm_functype_params(ty.ptr())
	return ty.convertTypeList(ptr)
}

// Returns the result types of this function type
func (ty *FuncType) Results() []*ValType {
	ptr := C.wasm_functype_results(ty.ptr())
	return ty.convertTypeList(ptr)
}

func (ty *FuncType) convertTypeList(list *C.wasm_valtype_vec_t) []*ValType {
	ret := make([]*ValType, list.size)

	base := unsafe.Pointer(list.data)
	var ptr *C.wasm_valtype_t
	for i := 0; i < int(list.size); i++ {
		ptr := *(**C.wasm_valtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkValType(ptr, ty)
		ret[i] = ty
	}
	return ret
}
