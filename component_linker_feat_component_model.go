package wasmtime

// #include <wasmtime.h>
// #include "component_shims_feat_component_model.h"
import "C"

import (
	"fmt"
	"runtime"
)

// ComponentLinker is used to satisfy the imports of a [Component] and
// instantiate it. Use [NewComponentLinker] to create one.
type ComponentLinker struct {
	_ptr   *C.wasmtime_component_linker_t
	locked bool
}

// NewComponentLinker creates a new [ComponentLinker] for the given engine.
func NewComponentLinker(engine *Engine) *ComponentLinker {
	ptr := C.wasmtime_component_linker_new(engine.ptr())
	runtime.KeepAlive(engine)
	return mkComponentLinker(ptr)
}

func mkComponentLinker(ptr *C.wasmtime_component_linker_t) *ComponentLinker {
	l := &ComponentLinker{_ptr: ptr}
	runtime.SetFinalizer(l, func(l *ComponentLinker) {
		l.Close()
	})
	return l
}

func (l *ComponentLinker) ptr() *C.wasmtime_component_linker_t {
	ret := l._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	if l.locked {
		panic("component linker is exclusively borrowed by a linker instance")
	}
	maybeGC()
	return ret
}

// Root exclusively borrows the linker's root namespace for definitions. Close
// the returned instance before using the linker again.
func (l *ComponentLinker) Root() *ComponentLinkerInstance {
	ptr := C.wasmtime_component_linker_root(l.ptr())
	l.locked = true
	runtime.KeepAlive(l)
	return mkComponentLinkerInstance(ptr, l)
}

// ComponentLinkerInstance is an exclusive definition namespace borrowed from
// a [ComponentLinker]. Close it before using the parent linker again.
type ComponentLinkerInstance struct {
	_ptr   *C.wasmtime_component_linker_instance_t
	parent *ComponentLinker
}

func mkComponentLinkerInstance(ptr *C.wasmtime_component_linker_instance_t, parent *ComponentLinker) *ComponentLinkerInstance {
	instance := &ComponentLinkerInstance{_ptr: ptr, parent: parent}
	runtime.SetFinalizer(instance, func(instance *ComponentLinkerInstance) { instance.Close() })
	return instance
}

func (i *ComponentLinkerInstance) ptr() *C.wasmtime_component_linker_instance_t {
	if i._ptr == nil {
		panic("component linker instance has been closed")
	}
	return i._ptr
}

// AddFunc defines a host component function in this namespace.
func (i *ComponentLinkerInstance) AddFunc(name string, callback ComponentHostFunc) error {
	if callback == nil {
		return fmt.Errorf("component host function callback is required")
	}
	index := insertComponentFunc(callback)
	err := C.go_component_linker_instance_add_func(
		i.ptr(), C._GoStringPtr(name), C._GoStringLen(name), C.size_t(index),
	)
	runtime.KeepAlive(i)
	runtime.KeepAlive(name)
	runtime.KeepAlive(callback)
	if err != nil {
		removeComponentFunc(index)
		return mkError(err)
	}
	return nil
}

// AddResource defines a host resource type and its destructor in this namespace.
func (i *ComponentLinkerInstance) AddResource(name string, resourceType *ComponentResourceType, destructor ComponentResourceDestructor) error {
	if destructor == nil {
		return fmt.Errorf("component resource destructor is required")
	}
	index := insertComponentResourceDestructor(destructor)
	err := C.go_component_linker_instance_add_resource(
		i.ptr(), C._GoStringPtr(name), C._GoStringLen(name), resourceType.ptr(), C.size_t(index),
	)
	runtime.KeepAlive(i)
	runtime.KeepAlive(name)
	runtime.KeepAlive(resourceType)
	runtime.KeepAlive(destructor)
	if err != nil {
		removeComponentResourceDestructor(index)
		return mkError(err)
	}
	return nil
}

// Close releases the exclusive namespace borrow.
func (i *ComponentLinkerInstance) Close() {
	if i == nil || i._ptr == nil {
		return
	}
	runtime.SetFinalizer(i, nil)
	C.wasmtime_component_linker_instance_delete(i._ptr)
	i._ptr = nil
	if i.parent != nil {
		i.parent.locked = false
		i.parent = nil
	}
}

// Instantiate creates a new [ComponentInstance] of `component` using the
// imports defined in this linker.
func (l *ComponentLinker) Instantiate(store Storelike, component *Component) (*ComponentInstance, error) {
	var val C.wasmtime_component_instance_t
	err := C.wasmtime_component_linker_instantiate(
		l.ptr(),
		store.Context(),
		component.ptr(),
		&val,
	)
	runtime.KeepAlive(l)
	runtime.KeepAlive(store)
	runtime.KeepAlive(component)
	if err != nil {
		return nil, mkError(err)
	}
	return mkComponentInstance(val), nil
}

// DefineUnknownImportsAsTraps defines every import of `component` that is not
// already satisfied by this linker as a function that traps when called.
//
// This is useful for instantiating components whose imports won't be invoked
// at runtime, or for diagnosing missing-import errors lazily.
func (l *ComponentLinker) DefineUnknownImportsAsTraps(component *Component) error {
	err := C.wasmtime_component_linker_define_unknown_imports_as_traps(
		l.ptr(), component.ptr(),
	)
	runtime.KeepAlive(l)
	runtime.KeepAlive(component)
	if err != nil {
		return mkError(err)
	}
	return nil
}

// TODO: WASIp2 / wasi:http integration via `wasmtime_component_linker_add_*`.

// Close deallocates this linker's state explicitly.
//
// For more information see the documentation for engine.Close().
func (l *ComponentLinker) Close() {
	if l._ptr == nil {
		return
	}
	if l.locked {
		panic("component linker is exclusively borrowed by a linker instance")
	}
	runtime.SetFinalizer(l, nil)
	C.wasmtime_component_linker_delete(l._ptr)
	l._ptr = nil
}
