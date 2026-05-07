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

// GetExportIndex looks up the export named `name` in this instance and
// returns a reusable [ComponentExportIndex] handle pointing at it. The
// `parent` argument behaves the same way as in [Component.GetExportIndex]
// (`nil` for the root namespace; an instance index for nested traversal).
// Returns `nil` if no matching export is found.
//
// Use this when the export you want lives inside an instantiated
// [ComponentInstance] rather than statically inside the [Component].
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
//
// This is a convenience wrapper that combines [GetExportIndex] +
// [GetFuncByIndex]. For repeated lookups of the same export — for example,
// when calling the function many times in a hot loop — cache a
// [ComponentExportIndex] up front and pass it to [GetFuncByIndex] directly
// to avoid the per-call name lookup.
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
//
// Prefer this over [GetFunc] when:
//
//   - The same export is looked up many times and the [ComponentExportIndex]
//     has been cached up front (avoids the per-call name lookup).
//   - The function lives inside a nested instance export, so the index
//     was produced by passing a non-nil `parent` to [Component.GetExportIndex]
//     or [ComponentInstance.GetExportIndex].
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
