package wasmtime

// #include "func.h"
// #include <wasmtime.h>
import "C"
import "unsafe"
import "reflect"
import "runtime"

type Func struct {
	_ptr   *C.wasm_func_t
	_owner interface{}
}

type Caller struct {
	ptr *C.wasmtime_caller_t
}

type newMapEntry struct {
	store    *Store
	callback func(*Caller, []Val) ([]Val, *Trap)
	nparams  int
	results  []*ValType
}

type wrapMapEntry struct {
	store    *Store
	callback reflect.Value
}

var NEW_MAP = make(map[int]newMapEntry)
var NEW_MAP_SLAB slab
var WRAP_MAP = make(map[int]wrapMapEntry)
var WRAP_MAP_SLAB slab
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
	idx := NEW_MAP_SLAB.allocate()
	NEW_MAP[idx] = newMapEntry{
		store:    store,
		callback: f,
		nparams:  len(ty.Params()),
		results:  ty.Results(),
	}

	ptr := C.c_func_new_with_env(
		store.ptr(),
		ty.ptr(),
		C.size_t(idx),
		0,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)

	return mkFunc(ptr, nil)
}

//export goTrampolineNew
func goTrampolineNew(
	caller_ptr *C.wasmtime_caller_t,
	env C.size_t,
	args_ptr *C.wasm_val_t,
	results_ptr *C.wasm_val_t,
) *C.wasm_trap_t {
	idx := int(env)
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

//export goFinalizeNew
func goFinalizeNew(env unsafe.Pointer) {
	idx := int(uintptr(env))
	delete(NEW_MAP, idx)
	NEW_MAP_SLAB.deallocate(idx)
}

// Wraps a native Go function, `f`, as a wasm `Func`.
//
// This function differs from `NewFunc` in that it will determine the type
// signature of the wasm function given the input value `f`. The `f` value
// provided must be a Go function. It may take any number of the following
// types as arguments:
//
// * `int32` - a wasm `i32`
// * `int64` - a wasm `i64`
// * `float32` - a wasm `f32`
// * `float64` - a wasm `f32`
// * `*Caller` - information about the caller's instance
//
// The Go function may return any number of values. It can return any number of
// primitive wasm values (integers/floats), and the last return value may
// optionally be `*Trap`. If a `*Trap` returned is `nil` then the other values
// are returned from the wasm function. Otherwise the `*Trap` is returned and
// it's considered as if the host function trapped.
//
// If the function `f` panics then the panic will be propagated to the caller.
func WrapFunc(
	store *Store,
	f interface{},
) *Func {
	// Make sure the `interface{}` passed in was indeed a function
	val := reflect.ValueOf(f)
	ty := val.Type()
	if ty.Kind() != reflect.Func {
		panic("callback provided must be a `func`")
	}

	// infer the parameter types, and `*Caller` type is special in the
	// parameters so be sure to case on that as well.
	params := make([]*ValType, 0, ty.NumIn())
	var caller *Caller
	for i := 0; i < ty.NumIn(); i++ {
		param_ty := ty.In(i)
		if param_ty != reflect.TypeOf(caller) {
			params = append(params, typeToValType(param_ty))
		}
	}

	// Then infer the result types, where a final `*Trap` result value is
	// also special.
	results := make([]*ValType, 0, ty.NumOut())
	var trap *Trap
	for i := 0; i < ty.NumOut(); i++ {
		result_ty := ty.Out(i)
		if i == ty.NumOut()-1 && result_ty == reflect.TypeOf(trap) {
			continue
		}
		results = append(results, typeToValType(result_ty))
	}
	wasm_ty := NewFuncType(params, results)

	// Store our `f` callback into the slab for wrapped functions, and now
	// we've got everything necessary to make thw asm handle.
	idx := WRAP_MAP_SLAB.allocate()
	WRAP_MAP[idx] = wrapMapEntry{
		callback: val,
		store:    store,
	}
	ptr := C.c_func_new_with_env(
		store.ptr(),
		wasm_ty.ptr(),
		C.size_t(idx),
		1, // this is `WrapFunc`, not `NewFunc`
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasm_ty)
	return mkFunc(ptr, nil)
}

func typeToValType(ty reflect.Type) *ValType {
	switch ty.Kind() {
	case reflect.Int32:
		return NewValType(KindI32)
	case reflect.Int64:
		return NewValType(KindI64)
	case reflect.Float32:
		return NewValType(KindF32)
	case reflect.Float64:
		return NewValType(KindF64)
	}
	panic("invalid type in callback that couldn't be converted to wasm type")
}

//export goTrampolineWrap
func goTrampolineWrap(
	caller_ptr *C.wasmtime_caller_t,
	env C.size_t,
	args_ptr *C.wasm_val_t,
	results_ptr *C.wasm_val_t,
) *C.wasm_trap_t {
	// Wrap our `Caller` argument in case it's needed
	caller := &Caller{ptr: caller_ptr}
	defer func() { caller.ptr = nil }()

	// Convert all our parameters to `[]reflect.Value`, taking special care
	// for `*Caller` but otherwise reading everything through `Val`.
	idx := int(env)
	entry := WRAP_MAP[idx]
	ty := entry.callback.Type()
	params := make([]reflect.Value, ty.NumIn())
	base := unsafe.Pointer(args_ptr)
	var raw C.wasm_val_t
	for i := 0; i < len(params); i++ {
		if ty.In(i) == reflect.TypeOf(caller) {
			params[i] = reflect.ValueOf(caller)
		} else {
			ptr := (*C.wasm_val_t)(base)
			val := Val{raw: *ptr}
			params[i] = reflect.ValueOf(val.Get())
			base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
		}
	}

	// Invoke the function, catching any panics to propagate later. Panics
	// result in immediately returning a trap.
	var results []reflect.Value
	func() {
		defer func() { LAST_PANIC = recover() }()
		results = entry.callback.Call(params)
	}()
	if LAST_PANIC != nil {
		trap := NewTrap(entry.store, "go panicked")
		runtime.SetFinalizer(trap, nil)
		return trap.ptr()
	}

	// And now we write all the results into memory depending on the type
	// of value that was returned.
	base = unsafe.Pointer(results_ptr)
	for _, result := range results {
		ptr := (*C.wasm_val_t)(base)
		switch val := result.Interface().(type) {
		case int32:
			*ptr = ValI32(val).raw
		case int64:
			*ptr = ValI64(val).raw
		case float32:
			*ptr = ValF32(val).raw
		case float64:
			*ptr = ValF64(val).raw
		case *Trap:
			if val != nil {
				runtime.SetFinalizer(val, nil)
				return val.ptr()
			}
		}
		base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
	}
	return nil
}

//export goFinalizeWrap
func goFinalizeWrap(env unsafe.Pointer) {
	idx := int(uintptr(env))
	delete(WRAP_MAP, idx)
	WRAP_MAP_SLAB.deallocate(idx)
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
		switch val := param.(type) {
		case int:
			switch params[i].Kind() {
			case KindI32:
				params_raw[i] = ValI32(int32(val)).raw
			case KindI64:
				params_raw[i] = ValI64(int64(val)).raw
			default:
				panic("integer provided for non-integer argument")
			}
		case int32:
			if params[i].Kind() != KindI32 {
				panic("unexpected i32 argument")
			}
			params_raw[i] = ValI32(val).raw
		case int64:
			if params[i].Kind() != KindI64 {
				panic("unexpected i64 argument")
			}
			params_raw[i] = ValI64(val).raw
		case float32:
			if params[i].Kind() != KindF32 {
				panic("unexpected f32 argument")
			}
			params_raw[i] = ValF32(val).raw
		case float64:
			if params[i].Kind() != KindF64 {
				panic("unexpected f64 argument")
			}
			params_raw[i] = ValF64(val).raw
		case Val:
			if params[i].Kind() != val.Kind() {
				panic("unexpected type in `Val`argument")
			}
			params_raw[i] = val.raw

		default:
			panic("couldn't convert provided argument to wasm type")
		}
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

func (f *Func) AsExtern() *Extern {
	ptr := C.wasm_func_as_extern(f.ptr())
	return mkExtern(ptr, f.owner())
}

// Gets an exported item from the caller's module
//
// May return `nil` if the export doesn't, if it's not a memory, if there isn't
// a caller, etc.
func (c *Caller) GetExport(name string) *Extern {
	if c.ptr == nil {
		return nil
	}
	name_vec := stringToBorrowedByteVec(name)
	ptr := C.wasmtime_caller_export_get(c.ptr, &name_vec)
	runtime.KeepAlive(name)
	if ptr == nil {
		return nil
	} else {
		return mkExtern(ptr, nil)
	}
}
