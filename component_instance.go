package wasmtime

// #include <wasmtime.h>
import "C"

import "runtime"

// ComponentInstance is a value-type handle to an instantiated component
// living inside a [Store]. It is created by
// [ComponentLinker.Instantiate].
//
// Instances do not have a destructor; their lifetime is tied to the store
// that owns them. Passing a [ComponentInstance] to a different store is
// undefined behavior and will likely abort the process.
type ComponentInstance struct {
	val C.wasmtime_component_instance_t
}

func mkComponentInstance(val C.wasmtime_component_instance_t) *ComponentInstance {
	return &ComponentInstance{val: val}
}

// GetExportIndex looks up the export named `name` in this instance,
// optionally nested within `parent`. Pass `nil` for `parent` to search the
// instance's root namespace. Returns `nil` if no matching export is found.
func (i *ComponentInstance) GetExportIndex(store Storelike, parent *ComponentExportIndex, name string) *ComponentExportIndex {
	var parentPtr *C.wasmtime_component_export_index_t
	if parent != nil {
		parentPtr = parent.ptr()
	}
	idxPtr := C.wasmtime_component_instance_get_export_index(
		&i.val,
		store.Context(),
		parentPtr,
		C._GoStringPtr(name),
		C._GoStringLen(name),
	)
	runtime.KeepAlive(i)
	runtime.KeepAlive(store)
	runtime.KeepAlive(parent)
	runtime.KeepAlive(name)
	if idxPtr == nil {
		return nil
	}
	return mkComponentExportIndex(idxPtr)
}

// GetFunc returns the [ComponentFunc] exported under `name` from this
// instance, or `nil` if no such function exists.
func (i *ComponentInstance) GetFunc(store Storelike, name string) *ComponentFunc {
	idx := i.GetExportIndex(store, nil, name)
	if idx == nil {
		return nil
	}
	defer idx.Close()
	return i.GetFuncByIndex(store, idx)
}

// GetFuncByIndex returns the [ComponentFunc] referenced by the given export
// index, or `nil` if the index does not correspond to a function export.
func (i *ComponentInstance) GetFuncByIndex(store Storelike, idx *ComponentExportIndex) *ComponentFunc {
	var f C.wasmtime_component_func_t
	found := bool(C.wasmtime_component_instance_get_func(
		&i.val,
		store.Context(),
		idx.ptr(),
		&f,
	))
	runtime.KeepAlive(i)
	runtime.KeepAlive(store)
	runtime.KeepAlive(idx)
	if !found {
		return nil
	}
	return mkComponentFunc(f)
}
