//go:build wasmtime_component_model

package wasmtime

/*
#include <wasmtime.h>
#include "component_shims_feat_component_model.h"
*/
import "C"

import (
	"fmt"
	"runtime"
)

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
