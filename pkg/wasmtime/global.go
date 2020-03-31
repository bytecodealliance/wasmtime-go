package wasmtime

// #include <wasmtime.h>
import "C"
import "errors"
import "runtime"

type Global struct {
	_ptr   *C.wasm_global_t
	_owner interface{}
}

// Creates a new `Global` in the given `Store` with the specified `ty` and
// initial value `val`.
func NewGlobal(
	store *Store,
	ty *GlobalType,
	val Val,
) (*Global, error) {
	ptr := C.wasm_global_new(
		store.ptr(),
		ty.ptr(),
		&val.raw,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	if ptr == nil {
		return nil, errors.New("wrong type of `Val` pass for global type")
	}

	return mkGlobal(ptr, nil), nil
}

func mkGlobal(ptr *C.wasm_global_t, owner interface{}) *Global {
	f := &Global{_ptr: ptr, _owner: owner}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Global) {
			C.wasm_global_delete(f._ptr)
		})
	}
	return f
}

func (f *Global) ptr() *C.wasm_global_t {
	ret := f._ptr
	maybeGC()
	return ret
}

func (f *Global) owner() interface{} {
	if f._owner != nil {
		return f._owner
	}
	return f
}

// Returns the type of this global
func (g *Global) Type() *GlobalType {
	ptr := C.wasm_global_type(g.ptr())
	runtime.KeepAlive(g)
	return mkGlobalType(ptr, nil)
}

// Gets the value of this global
func (g *Global) Get() Val {
	ret := Val{}
	C.wasm_global_get(g.ptr(), &ret.raw)
	runtime.KeepAlive(g)
	return ret
}

// Sets the value of this global
func (g *Global) Set(val Val) {
	C.wasm_global_set(g.ptr(), &val.raw)
	runtime.KeepAlive(g)
}
