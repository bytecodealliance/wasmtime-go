package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"

type Global struct {
	_ptr     *C.wasm_global_t
	_owner   interface{}
	freelist *freeList
}

// Creates a new `Global` in the given `Store` with the specified `ty` and
// initial value `val`.
func NewGlobal(
	store *Store,
	ty *GlobalType,
	val Val,
) (*Global, error) {
	var ptr *C.wasm_global_t
	err := C.wasmtime_global_new(
		store.ptr(),
		ty.ptr(),
		&val.raw,
		&ptr,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(ty)
	if err != nil {
		return nil, mkError(err)
	}

	return mkGlobal(ptr, store.freelist, nil), nil
}

func mkGlobal(ptr *C.wasm_global_t, freelist *freeList, owner interface{}) *Global {
	f := &Global{_ptr: ptr, _owner: owner, freelist: freelist}
	if owner == nil {
		runtime.SetFinalizer(f, func(f *Global) {
			f.freelist.lock.Lock()
			defer f.freelist.lock.Unlock()
			f.freelist.globals = append(f.freelist.globals, f._ptr)
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
func (g *Global) Set(val Val) error {
	err := C.wasmtime_global_set(g.ptr(), &val.raw)
	runtime.KeepAlive(g)
	if err == nil {
		return nil
	} else {
		return mkError(err)
	}
}

func (g *Global) AsExtern() *Extern {
	ptr := C.wasm_global_as_extern(g.ptr())
	return mkExtern(ptr, g.freelist, g.owner())
}
