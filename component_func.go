package wasmtime

// #include <wasmtime.h>
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// ComponentFunc is a handle to a function exported by a [ComponentInstance].
//
// Like [ComponentInstance], this is a value-type handle whose lifetime is
// tied to the store that produced it. Passing it to a different store is
// undefined behavior.
type ComponentFunc struct {
	val C.wasmtime_component_func_t
}

func mkComponentFunc(val C.wasmtime_component_func_t) *ComponentFunc {
	return &ComponentFunc{val: val}
}

// Call invokes this component function with the given Go values as arguments
// and returns its result, if any.
//
// Argument types must match the component function's parameter types
// according to the following mapping:
//
//   - WIT `bool`            <-> Go `bool`
//   - WIT `s8`/`s16`/`s32`/`s64` <-> Go `int8`/`int16`/`int32`/`int64`
//   - WIT `u8`/`u16`/`u32`/`u64` <-> Go `uint8`/`uint16`/`uint32`/`uint64`
//   - WIT `f32`/`f64`        <-> Go `float32`/`float64`
//   - WIT `char`             <-> Go `rune` (or `int32`)
//   - WIT `string`           <-> Go `string`
//
// Functions whose signatures use composite types (list, record, tuple,
// variant, enum, option, result, flags, resource) are not yet supported and
// will return an error indicating an unsupported type kind.
//
// The result is `nil` for void functions and a single Go value otherwise. WIT
// only supports zero or one result, so multi-value returns are represented as
// a tuple at the WIT level (and are unsupported here for now).
func (f *ComponentFunc) Call(store Storelike, args ...interface{}) (interface{}, error) {
	// The function's type is consulted internally to know the expected kinds
	// of each parameter (so that primitives like int32 vs rune can be routed
	// to s32 vs char correctly). A public type-reflection API is intentionally
	// not exposed yet; that will come together with composite-type support.
	// TODO: expose a public Type API (ComponentType / ComponentValType / ...)
	// once the composite-type representation is settled.
	fty := C.wasmtime_component_func_type(&f.val, store.Context())
	runtime.KeepAlive(store)
	if fty == nil {
		return nil, fmt.Errorf("could not retrieve component function type")
	}
	defer C.wasmtime_component_func_type_delete(fty)

	paramCount := int(C.wasmtime_component_func_type_param_count(fty))
	if len(args) != paramCount {
		return nil, fmt.Errorf("wrong number of arguments: got %d, expected %d", len(args), paramCount)
	}

	cArgs := make([]C.wasmtime_component_val_t, paramCount)
	defer func() {
		for i := range cArgs {
			C.wasmtime_component_val_delete(&cArgs[i])
		}
	}()
	for i := 0; i < paramCount; i++ {
		var nameP *C.char
		var nameLen C.size_t
		var paramTy C.wasmtime_component_valtype_t
		if !bool(C.wasmtime_component_func_type_param_nth(fty, C.size_t(i), &nameP, &nameLen, &paramTy)) {
			return nil, fmt.Errorf("could not retrieve parameter %d", i)
		}
		err := componentMarshalArg(args[i], paramTy.kind, &cArgs[i])
		C.wasmtime_component_valtype_delete(&paramTy)
		if err != nil {
			return nil, fmt.Errorf("argument %d: %w", i, err)
		}
	}

	var resultTy C.wasmtime_component_valtype_t
	hasResult := bool(C.wasmtime_component_func_type_result(fty, &resultTy))
	var resultArr [1]C.wasmtime_component_val_t
	var resultsPtr *C.wasmtime_component_val_t
	var resultsLen C.size_t
	if hasResult {
		defer C.wasmtime_component_valtype_delete(&resultTy)
		defer C.wasmtime_component_val_delete(&resultArr[0])
		resultsPtr = &resultArr[0]
		resultsLen = 1
	}

	var argsPtr *C.wasmtime_component_val_t
	if len(cArgs) > 0 {
		argsPtr = (*C.wasmtime_component_val_t)(unsafe.Pointer(&cArgs[0]))
	}
	cerr := C.wasmtime_component_func_call(
		&f.val,
		store.Context(),
		argsPtr,
		C.size_t(len(cArgs)),
		resultsPtr,
		resultsLen,
	)
	runtime.KeepAlive(f)
	runtime.KeepAlive(store)
	if cerr != nil {
		return nil, mkError(cerr)
	}

	if !hasResult {
		return nil, nil
	}
	return componentUnmarshalVal(&resultArr[0])
}
