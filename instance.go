package wasmtime

// #include <wasmtime.h>
import "C"
import "runtime"
import "unsafe"

type Instance struct {
	_ptr     *C.wasm_instance_t
	exports  map[string]*Extern
	freelist *freeList
}

// Instantiates a WebAssembly `module` with the `imports` provided.
//
// This function will attempt to create a new wasm instance given the provided
// imports. This can fail if the wrong number of imports are specified, the
// imports aren't of the right type, or for other resource-related issues.
//
// This will also run the `start` function of the instance, returning an error
// if it traps.
func NewInstance(store *Store, module *Module, imports []*Extern) (*Instance, error) {
	imports_raw := make([]*C.wasm_extern_t, len(imports))
	for i, imp := range imports {
		imports_raw[i] = imp.ptr()
	}
	var imports_raw_ptr **C.wasm_extern_t
	if len(imports) > 0 {
		imports_raw_ptr = &imports_raw[0]
	}
	var trap *C.wasm_trap_t
	var ptr *C.wasm_instance_t
	err := C.wasmtime_instance_new(
		store.ptr(),
		module.ptr(),
		imports_raw_ptr,
		C.size_t(len(imports)),
		&ptr,
		&trap,
	)
	runtime.KeepAlive(store)
	runtime.KeepAlive(module)
	runtime.KeepAlive(imports)
	runtime.KeepAlive(imports_raw)
	if err != nil {
		return nil, mkError(err)
	}
	if trap != nil {
		return nil, mkTrap(trap)
	}
	return mkInstance(ptr, module), nil
}

func mkInstance(ptr *C.wasm_instance_t, module *Module) *Instance {
	instance := &Instance{
		_ptr:     ptr,
		exports:  make(map[string]*Extern),
		freelist: module.Store.freelist,
	}
	runtime.SetFinalizer(instance, func(instance *Instance) {
		freelist := instance.freelist
		freelist.lock.Lock()
		defer freelist.lock.Unlock()
		freelist.instances = append(freelist.instances, instance._ptr)
	})
	exports := instance.Exports()
	for i, ty := range module.Exports() {
		instance.exports[ty.Name()] = exports[i]
	}
	return instance
}

func (m *Instance) ptr() *C.wasm_instance_t {
	ret := m._ptr
	maybeGC()
	return ret
}

type externList struct {
	vec C.wasm_extern_vec_t
}

// Returns a list of exports from this instance.
//
// Each export is returned as a `*Extern` and lines up with the exports list of
// the associated `Module`.
func (i *Instance) Exports() []*Extern {
	externs := &externList{}
	C.wasm_instance_exports(i.ptr(), &externs.vec)
	runtime.KeepAlive(i)
	freelist := i.freelist
	runtime.SetFinalizer(externs, func(externs *externList) {
		freelist.lock.Lock()
		defer freelist.lock.Unlock()
		freelist.extern_vecs = append(freelist.extern_vecs, &externs.vec)
	})

	ret := make([]*Extern, int(externs.vec.size))
	base := unsafe.Pointer(externs.vec.data)
	var ptr *C.wasm_extern_t
	for i := 0; i < int(externs.vec.size); i++ {
		ptr := *(**C.wasm_extern_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkExtern(ptr, freelist, externs)
		ret[i] = ty
	}
	return ret
}

// Attempts to find an export on this instance by `name`
//
// May return `nil` if this instance has no export named `name`
func (i *Instance) GetExport(name string) *Extern {
	return i.exports[name]
}
