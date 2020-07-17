package wasmtime

// #include "shims.h"
import "C"
import (
	"errors"
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

// Func is a function instance, which is the runtime representation of a function.
// It effectively is a closure of the original function over the runtime module instance of its originating module.
// The module instance is used to resolve references to other definitions during execution of the function.
// Read more in [spec](https://webassembly.github.io/spec/core/exec/runtime.html#function-instances)
type Func struct {
	_ptr     *C.wasm_func_t
	_owner   interface{}
	freelist *freeList
}

// TODO
type Caller struct {
	ptr   *C.wasmtime_caller_t
	store *Store
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

var gLock sync.Mutex
var gNewMap = make(map[int]newMapEntry)
var gNewMapSlab slab
var gWrapMap = make(map[int]wrapMapEntry)
var gWrapMapSlab slab
var gCallerPanics = make(map[*freeList]interface{})

// NewFunc creates a new `Func` with the given `ty` which, when called, will call `f`
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
	gLock.Lock()
	idx := gNewMapSlab.allocate()
	gNewMap[idx] = newMapEntry{
		store:    store,
		callback: f,
		nparams:  len(ty.Params()),
		results:  ty.Results(),
	}
	gLock.Unlock()

	ptr := C.c_func_new_with_env(
		store.ptr(),
		ty.ptr(),
		C.size_t(idx),
		0,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)

	return mkFunc(ptr, store.freelist, nil)
}

//export goTrampolineNew
func goTrampolineNew(
	callerPtr *C.wasmtime_caller_t,
	env C.size_t,
	argsPtr *C.wasm_val_t,
	resultsPtr *C.wasm_val_t,
) *C.wasm_trap_t {
	idx := int(env)
	gLock.Lock()
	entry := gNewMap[idx]
	gLock.Unlock()

	caller := &Caller{ptr: callerPtr, store: entry.store}
	defer func() { caller.ptr = nil }()

	params := make([]Val, entry.nparams)
	var val C.wasm_val_t
	base := unsafe.Pointer(argsPtr)
	for i := 0; i < len(params); i++ {
		ptr := (*C.wasm_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		params[i] = mkVal(ptr, entry.store.freelist)
	}

	var results []Val
	var trap *Trap
	var lastPanic interface{}
	func() {
		defer func() { lastPanic = recover() }()
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
	if trap == nil && lastPanic != nil {
		gLock.Lock()
		gCallerPanics[entry.store.freelist] = lastPanic
		gLock.Unlock()
		trap = NewTrap(entry.store, "go panicked")
	}
	if trap != nil {
		runtime.SetFinalizer(trap, nil)
		return trap.ptr()
	}

	base = unsafe.Pointer(resultsPtr)
	for i := 0; i < len(results); i++ {
		ptr := (*C.wasm_val_t)(unsafe.Pointer(uintptr(base) + uintptr(i)*unsafe.Sizeof(val)))
		C.wasm_val_copy(ptr, results[i].ptr())
	}
	runtime.KeepAlive(results)
	return nil
}

//export goFinalizeNew
func goFinalizeNew(env unsafe.Pointer) {
	idx := int(uintptr(env))
	gLock.Lock()
	defer gLock.Unlock()
	delete(gNewMap, idx)
	gNewMapSlab.deallocate(idx)
}

// WrapFunc wraps a native Go function, `f`, as a wasm `Func`.
//
// This function differs from `NewFunc` in that it will determine the type
// signature of the wasm function given the input value `f`. The `f` value
// provided must be a Go function. It may take any number of the following
// types as arguments:
//
// `int32` - a wasm `i32`
//
// `int64` - a wasm `i64`
//
// `float32` - a wasm `f32`
//
// `float64` - a wasm `f32`
//
// `*Caller` - information about the caller's instance
//
// `*Func` - a wasm `funcref`
//
// anything else - a wasm `externref`
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
		paramTy := ty.In(i)
		if paramTy != reflect.TypeOf(caller) {
			params = append(params, typeToValType(paramTy))
		}
	}

	// Then infer the result types, where a final `*Trap` result value is
	// also special.
	results := make([]*ValType, 0, ty.NumOut())
	var trap *Trap
	for i := 0; i < ty.NumOut(); i++ {
		resultTy := ty.Out(i)
		if i == ty.NumOut()-1 && resultTy == reflect.TypeOf(trap) {
			continue
		}
		results = append(results, typeToValType(resultTy))
	}
	wasmTy := NewFuncType(params, results)

	// Store our `f` callback into the slab for wrapped functions, and now
	// we've got everything necessary to make thw asm handle.
	gLock.Lock()
	idx := gWrapMapSlab.allocate()
	gWrapMap[idx] = wrapMapEntry{
		callback: val,
		store:    store,
	}
	gLock.Unlock()

	ptr := C.c_func_new_with_env(
		store.ptr(),
		wasmTy.ptr(),
		C.size_t(idx),
		1, // this is `WrapFunc`, not `NewFunc`
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasmTy)
	return mkFunc(ptr, store.freelist, nil)
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
	var f *Func
	if ty == reflect.TypeOf(f) {
		return NewValType(KindFuncref)
	}
	return NewValType(KindExternref)
}

//export goTrampolineWrap
func goTrampolineWrap(
	callerPtr *C.wasmtime_caller_t,
	env C.size_t,
	argsPtr *C.wasm_val_t,
	resultsPtr *C.wasm_val_t,
) *C.wasm_trap_t {
	// Convert all our parameters to `[]reflect.Value`, taking special care
	// for `*Caller` but otherwise reading everything through `Val`.
	idx := int(env)
	gLock.Lock()
	entry := gWrapMap[idx]
	gLock.Unlock()

	// Wrap our `Caller` argument in case it's needed
	caller := &Caller{ptr: callerPtr, store: entry.store}
	defer func() { caller.ptr = nil }()

	ty := entry.callback.Type()
	params := make([]reflect.Value, ty.NumIn())
	base := unsafe.Pointer(argsPtr)
	var raw C.wasm_val_t
	for i := 0; i < len(params); i++ {
		if ty.In(i) == reflect.TypeOf(caller) {
			params[i] = reflect.ValueOf(caller)
		} else {
			ptr := (*C.wasm_val_t)(base)
			val := mkVal(ptr, entry.store.freelist)
			params[i] = reflect.ValueOf(val.Get())
			base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
		}
	}

	// Invoke the function, catching any panics to propagate later. Panics
	// result in immediately returning a trap.
	var results []reflect.Value
	var lastPanic interface{}
	func() {
		defer func() { lastPanic = recover() }()
		results = entry.callback.Call(params)
	}()
	if lastPanic != nil {
		gLock.Lock()
		gCallerPanics[entry.store.freelist] = lastPanic
		gLock.Unlock()
		trap := NewTrap(entry.store, "go panicked")
		runtime.SetFinalizer(trap, nil)
		return trap.ptr()
	}

	// And now we write all the results into memory depending on the type
	// of value that was returned.
	base = unsafe.Pointer(resultsPtr)
	for _, result := range results {
		ptr := (*C.wasm_val_t)(base)
		switch val := result.Interface().(type) {
		case int32:
			*ptr = *ValI32(val).ptr()
		case int64:
			*ptr = *ValI64(val).ptr()
		case float32:
			*ptr = *ValF32(val).ptr()
		case float64:
			*ptr = *ValF64(val).ptr()
		case *Func:
			raw := ValFuncref(val)
			C.wasm_val_copy(ptr, raw.ptr())
			runtime.KeepAlive(raw)
		case *Trap:
			if val != nil {
				runtime.SetFinalizer(val, nil)
				return val.ptr()
			}
		default:
			raw := ValExternref(val)
			C.wasm_val_copy(ptr, raw.ptr())
			runtime.KeepAlive(raw)
		}
		base = unsafe.Pointer(uintptr(base) + unsafe.Sizeof(raw))
	}
	return nil
}

//export goFinalizeWrap
func goFinalizeWrap(env unsafe.Pointer) {
	idx := int(uintptr(env))
	gLock.Lock()
	defer gLock.Unlock()
	delete(gWrapMap, idx)
	gWrapMapSlab.deallocate(idx)
}

func mkFunc(ptr *C.wasm_func_t, freelist *freeList, owner interface{}) *Func {
	f := &Func{_ptr: ptr, _owner: owner, freelist: freelist}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Func) {
			freelist.lock.Lock()
			defer freelist.lock.Unlock()
			freelist.funcs = append(freelist.funcs, f._ptr)
		})
	}
	return f
}

func (f *Func) ptr() *C.wasm_func_t {
	f.freelist.clear()
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

// Type returns the type of this func
func (f *Func) Type() *FuncType {
	ptr := C.wasm_func_type(f.ptr())
	runtime.KeepAlive(f)
	return mkFuncType(ptr, nil)
}

// ParamArity returns the numer of parameters this function expects
func (f *Func) ParamArity() int {
	ret := C.wasm_func_param_arity(f.ptr())
	runtime.KeepAlive(f)
	return int(ret)
}

// ResultArity returns the numer of results this function produces
func (f *Func) ResultArity() int {
	ret := C.wasm_func_result_arity(f.ptr())
	runtime.KeepAlive(f)
	return int(ret)
}

// Call invokes this function with the provided `args`.
//
// This variadic function must be invoked with the correct number and type of
// `args` as specified by the type of this function. This property is checked
// at runtime. Each `args` may have one of the following types:
//
// `int32` - a wasm `i32`
//
// `int64` - a wasm `i64`
//
// `float32` - a wasm `f32`
//
// `float64` - a wasm `f64`
//
// `Val` - correspond to a wasm value
//
// `*Func` - a wasm `funcref`
//
// anything else - a wasm `externref`
//
// This function will have one of three results:
//
// 1. If the function returns successfully, then the `interface{}` return
// argument will be the result of the function. If there were 0 results then
// this value is `nil`. If there was one result then this is that result.
// Otherwise if there were multiple results then `[]Val` is returned.
//
// 2. If this function invocation traps, then the returned `interface{}` value
// will be `nil` and a non-`nil` `*Trap` will be returned with information
// about the trap that happened.
//
// 3. If a panic in Go ends up happening somewhere, then this function will
// panic.
func (f *Func) Call(args ...interface{}) (interface{}, error) {
	params := f.Type().Params()
	if len(args) > len(params) {
		return nil, errors.New("too many arguments provided")
	}
	paramsRaw := make([]C.wasm_val_t, len(args))
	synthesizedParams := make([]Val, 0)
	for i, param := range args {
		switch val := param.(type) {
		case int:
			switch params[i].Kind() {
			case KindI32:
				paramsRaw[i] = *ValI32(int32(val)).ptr()
			case KindI64:
				paramsRaw[i] = *ValI64(int64(val)).ptr()
			default:
				return nil, errors.New("integer provided for non-integer argument")
			}
		case int32:
			paramsRaw[i] = *ValI32(val).ptr()
		case int64:
			paramsRaw[i] = *ValI64(val).ptr()
		case float32:
			paramsRaw[i] = *ValF32(val).ptr()
		case float64:
			paramsRaw[i] = *ValF64(val).ptr()
		case *Func:
			ffi := ValFuncref(val)
			paramsRaw[i] = *ffi.ptr()
			synthesizedParams = append(synthesizedParams, ffi)
		case Val:
			paramsRaw[i] = *val.ptr()

		default:
			ffi := ValExternref(val)
			paramsRaw[i] = *ffi.ptr()
			synthesizedParams = append(synthesizedParams, ffi)
		}
	}

	resultsRaw := make([]C.wasm_val_t, f.ResultArity())

	var paramsPtr, resultsPtr *C.wasm_val_t
	var trap *C.wasm_trap_t
	if len(paramsRaw) > 0 {
		paramsPtr = &paramsRaw[0]
	}
	if len(resultsRaw) > 0 {
		resultsPtr = &resultsRaw[0]
	}

	err := C.wasmtime_func_call(
		f.ptr(),
		paramsPtr,
		C.size_t(len(paramsRaw)),
		resultsPtr,
		C.size_t(len(resultsRaw)),
		&trap,
	)
	runtime.KeepAlive(f)
	runtime.KeepAlive(paramsRaw)
	runtime.KeepAlive(args)
	runtime.KeepAlive(synthesizedParams)

	if err != nil {
		return nil, mkError(err)
	}

	if trap != nil {
		trap := mkTrap(trap)
		gLock.Lock()
		defer gLock.Unlock()
		lastPanic := gCallerPanics[f.freelist]
		delete(gCallerPanics, f.freelist)
		if lastPanic != nil {
			panic(lastPanic)
		}
		return nil, trap
	}

	if len(resultsRaw) == 0 {
		return nil, nil
	} else if len(resultsRaw) == 1 {
		return takeVal(&resultsRaw[0], f.freelist).Get(), nil
	} else {
		results := make([]Val, len(resultsRaw))
		for i := 0; i < len(resultsRaw); i++ {
			results[i] = takeVal(&resultsRaw[i], f.freelist)
		}
		return results, nil
	}

}

func (f *Func) AsExtern() *Extern {
	ptr := C.wasm_func_as_extern(f.ptr())
	return mkExtern(ptr, f.freelist, f.owner())
}

// GetExport gets an exported item from the caller's module.
//
// May return `nil` if the export doesn't, if it's not a memory, if there isn't
// a caller, etc.
func (c *Caller) GetExport(name string) *Extern {
	if c.ptr == nil {
		return nil
	}
	ptr := C.go_caller_export_get(
		c.ptr,
		C._GoStringPtr(name),
		C._GoStringLen(name),
	)
	runtime.KeepAlive(name)
	runtime.KeepAlive(c)
	if ptr == nil {
		return nil
	}

	return mkExtern(ptr, c.store.freelist, nil)
}
