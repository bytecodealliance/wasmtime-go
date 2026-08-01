//go:build wasmtime_component_model

package wasmtime

/*
#include <wasmtime.h>
#include <stdlib.h>

static inline wasmtime_component_val_t *go_component_val_at(void *p, size_t i) {
  return &((wasmtime_component_val_t *)p)[i];
}
static inline void go_component_val_shallow_copy(void *p, size_t i, const wasmtime_component_val_t *v) {
  ((wasmtime_component_val_t *)p)[i] = *v;
}
*/
import "C"

import (
	"fmt"
	"runtime"
)

// ComponentFunc is a component-model function tied to one Store.
type ComponentFunc struct{ val C.wasmtime_component_func_t }

// ComponentFuncType describes a component function's parameters and result.
type ComponentFuncType struct {
	_ptr *C.wasmtime_component_func_type_t
}

func (f *ComponentFunc) Type(store Storelike) *ComponentFuncType {
	ptr := C.wasmtime_component_func_type(&f.val, store.Context())
	runtime.KeepAlive(f)
	runtime.KeepAlive(store)
	return &ComponentFuncType{_ptr: ptr}
}

func (t *ComponentFuncType) ParamCount() int {
	return int(C.wasmtime_component_func_type_param_count(t._ptr))
}

func (t *ComponentFuncType) HasResult() bool {
	var result C.wasmtime_component_valtype_t
	found := bool(C.wasmtime_component_func_type_result(t._ptr, &result))
	if found {
		C.wasmtime_component_valtype_delete(&result)
	}
	return found
}

func (t *ComponentFuncType) Close() {
	if t != nil && t._ptr != nil {
		C.wasmtime_component_func_type_delete(t._ptr)
		t._ptr = nil
	}
}

// Call invokes the component function synchronously. Returned values are owned
// and must be closed by the caller.
func (f *ComponentFunc) Call(store Storelike, args []*ComponentVal) ([]*ComponentVal, error) {
	typeInfo := f.Type(store)
	if typeInfo == nil || typeInfo._ptr == nil {
		return nil, fmt.Errorf("component function type unavailable")
	}
	defer typeInfo.Close()
	resultCount := 0
	if typeInfo.HasResult() {
		resultCount = 1
	}
	argsMem := componentValArray(len(args))
	if argsMem != nil {
		defer C.free(argsMem)
	}
	for i, arg := range args {
		C.go_component_val_shallow_copy(argsMem, C.size_t(i), arg.ptr())
	}
	resultsMem := componentValArray(resultCount)
	if resultsMem != nil {
		defer C.free(resultsMem)
	}
	err := enterWasm(store, func(_ **C.wasm_trap_t) *C.wasmtime_error_t {
		return C.wasmtime_component_func_call(
			&f.val, store.Context(),
			(*C.wasmtime_component_val_t)(argsMem), C.size_t(len(args)),
			(*C.wasmtime_component_val_t)(resultsMem), C.size_t(resultCount),
		)
	})
	runtime.KeepAlive(f)
	runtime.KeepAlive(store)
	runtime.KeepAlive(args)
	if err != nil {
		return nil, err
	}
	results := make([]*ComponentVal, resultCount)
	for i := range results {
		results[i] = ownComponentVal(*C.go_component_val_at(resultsMem, C.size_t(i)))
	}
	return results, nil
}

// GetFuncByIndex resolves a function from a reusable component export index.
func (i *ComponentInstance) GetFuncByIndex(store Storelike, index *ComponentExportIndex) *ComponentFunc {
	var value C.wasmtime_component_func_t
	found := C.wasmtime_component_instance_get_func(&i.val, store.Context(), index.ptr(), &value)
	runtime.KeepAlive(i)
	runtime.KeepAlive(store)
	runtime.KeepAlive(index)
	if !bool(found) {
		return nil
	}
	return &ComponentFunc{val: value}
}

// GetFunc resolves a root or nested exported component function by name.
func (i *ComponentInstance) GetFunc(store Storelike, parent *ComponentExportIndex, name string) *ComponentFunc {
	index := i.GetExportIndex(store, parent, name)
	if index == nil {
		return nil
	}
	defer index.Close()
	return i.GetFuncByIndex(store, index)
}
