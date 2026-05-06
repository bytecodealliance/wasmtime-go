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
