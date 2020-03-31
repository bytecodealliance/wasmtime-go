package wasmtime

// #include "func.h"
// #include <wasmtime.h>
import "C"
import "unsafe"
import "runtime"

type Func struct {
	_ptr   *C.wasm_func_t
	_owner interface{}
}

type Caller struct {
	ptr *C.wasmtime_caller_t
}

type mapEntry struct {
	store    *Store
	callback func(*Caller, []Val) ([]Val, *Trap)
	nparams  int
	results  []*ValType
}

var NEW_MAP = make(map[int]mapEntry)
var SLAB slab
var LAST_PANIC interface{}

// Creates a new `Func` with the given `ty` which, when called, will call `f`
//
// The `ty` given is the wasm type signature of the `Func` to create. When called
// the `f` callback receives two arguments. The first is a `Caller` to learn
// information about the calling context and the second is a list of arguments
// represented as a `Val`. The parameters are guaranteed to match the parameters
// types specified in `ty`.
//
// The `f` callback is expected to produce one of two values. Results can be
// returned as an array of `[]Val`. The number and types of these results much
// match the `ty` given, otherwise the program will panic. The `f` callback can
// also produce a trap which will trigger trap unwinding in wasm, and the trap
// will be returned to the original caller.
//
// If the `f` callback panics then the panic will be propagated to the caller
// as well.
func NewFunc(
	store *Store,
	ty *FuncType,
	f func(*Caller, []Val) ([]Val, *Trap),
) *Func {
	idx := SLAB.allocate()
	NEW_MAP[idx] = mapEntry{
		store:    store,
		callback: f,
		nparams:  len(ty.Params()),
		results:  ty.Results(),
	}

	ptr := C.c_func_new_with_env(
		store.ptr(),
		ty.ptr(),
		unsafe.Pointer(uintptr(idx)),
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)

	return mkFunc(ptr, nil)
}

func mkFunc(ptr *C.wasm_func_t, owner interface{}) *Func {
	f := &Func{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Func) {
			C.wasm_func_delete(f._ptr)
		})
	}
	return f
}

//export goTrampoline
func goTrampoline(
	caller_ptr *C.wasmtime_caller_t,
	env unsafe.Pointer,
	args_ptr *C.wasm_val_t,
	results_ptr *C.wasm_val_t,
) *C.wasm_trap_t {
	idx := int(uintptr(env))
	caller := &Caller{ptr: caller_ptr}
	defer func() { caller.ptr = nil }()

	entry := NEW_MAP[idx]
	params := make([]Val, entry.nparams)
	var val C.wasm_val_t
	base := unsafe.Pointer(args_ptr)
	for i := 0; i < len(params); i++ {
		ptr := (*C.wasm_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		params[i] = Val{raw: *ptr}
	}

	var results []Val
	var trap *Trap
	func() {
		defer func() { LAST_PANIC = recover() }()
		results, trap = entry.callback(caller, params)
		if trap != nil {
			return
		}
		if len(results) != len(entry.results) {
			panic("callback didn't produce the correct number of results")
		}
		for i, ty := range entry.results {
			if results[i].Kind() != ty.Kind() {
				panic("callback produced wrong type of result")
			}
		}
	}()
	if trap == nil && LAST_PANIC != nil {
		trap = NewTrap(entry.store, "go panicked")
	}
	if trap != nil {
		runtime.SetFinalizer(trap, nil)
		return trap.ptr()
	}

	base = unsafe.Pointer(results_ptr)
	for i := 0; i < len(results); i++ {
		ptr := (*C.wasm_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		*ptr = results[i].raw
	}
	return nil
}

//export goFinalize
func goFinalize(env unsafe.Pointer) {
	idx := int(uintptr(env))
	delete(NEW_MAP, idx)
	SLAB.deallocate(idx)
}

func (f *Func) ptr() *C.wasm_func_t {
	ret := f._ptr
	maybeGC()
	return ret
}

func (f *Func) owner() interface{} {
	if f._owner != nil {
		return f._owner
	}
	return f
}

// Returns the type of this func
func (f *Func) Type() *FuncType {
	ptr := C.wasm_func_type(f.ptr())
	runtime.KeepAlive(f)
	return mkFuncType(ptr, nil)
}

// Returns the numer of parameters this function expects
func (f *Func) ParamArity() int {
	ret := C.wasm_func_param_arity(f.ptr())
	runtime.KeepAlive(f)
	return int(ret)
}

// Returns the numer of results this function produces
func (f *Func) ResultArity() int {
	ret := C.wasm_func_result_arity(f.ptr())
	runtime.KeepAlive(f)
	return int(ret)
}

// Invokes this function with the provided `args`.
//
// This variadic function must be invoked with the correct number and type of
// `args` as specified by the type of this function. This property is checked
// at runtime. Each `args` may have one of the following types:
//
// * `int32` - a wasm `i32`
// * `int64` - a wasm `i64`
// * `float32` - a wasm `f32`
// * `float64` - a wasm `f64`
// * `Val` - correspond to a wasm value
//
// Any other types of `args` will cause this function to panic.
//
// This function will have one of three results:
//
// 1. If the function returns successfully, then the `interface{}` return
//    argument will be the result of the function. If there were 0 results then
//    this value is `nil`. If there was one result then this is that result.
//    Otherwise if there were multiple results then `[]Val` is returned.
//
// 2. If this function invocation traps, then the returned `interface{}` value
//    will be `nil` and a non-`nil` `*Trap` will be returned with information
//    about the trap that happened.
//
// 3. If a panic in Go ends up happening somewhere, then this function will
//    panic.
func (f *Func) Call(args ...interface{}) (interface{}, *Trap) {
	params := f.Type().Params()
	if len(args) != len(params) {
		panic("wrong number of arguments provided")
	}
	params_raw := make([]C.wasm_val_t, len(params))
	for i, param := range args {
		if val, ok := param.(int32); ok {
			params_raw[i] = ValI32(val).raw
			continue
		}
		if val, ok := param.(int64); ok {
			params_raw[i] = ValI64(val).raw
			continue
		}
		if val, ok := param.(float32); ok {
			params_raw[i] = ValF32(val).raw
			continue
		}
		if val, ok := param.(float64); ok {
			params_raw[i] = ValF64(val).raw
			continue
		}
		if val, ok := param.(Val); ok {
			params_raw[i] = val.raw
			continue
		}
		panic("couldn't convert provided argument to wasm type")
	}

	results_raw := make([]C.wasm_val_t, f.ResultArity())

	var params_ptr, results_ptr *C.wasm_val_t
	if len(params_raw) > 0 {
		params_ptr = &params_raw[0]
	}
	if len(results_raw) > 0 {
		results_ptr = &results_raw[0]
	}

	trap := C.wasm_func_call(f.ptr(), params_ptr, results_ptr)
	runtime.KeepAlive(f)
	runtime.KeepAlive(params_raw)

	if trap != nil {
		trap := mkTrap(trap)
		last_panic := LAST_PANIC
		LAST_PANIC = nil
		if last_panic != nil {
			panic(last_panic)
		}
		return nil, trap
	}

	if len(results_raw) == 0 {
		return nil, nil
	} else if len(results_raw) == 1 {
		val := Val{raw: results_raw[0]}
		return val.Get(), nil
	} else {
		results := make([]Val, len(results_raw))
		for i, raw := range results_raw {
			results[i] = Val{raw}
		}
		return results, nil
	}

}
