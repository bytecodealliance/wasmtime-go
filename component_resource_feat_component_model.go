//go:build wasmtime_component_model

package wasmtime

/*
#include <wasmtime.h>

static inline void go_component_val_resource(wasmtime_component_val_t *v, wasmtime_component_resource_any_t *resource) { v->kind = WASMTIME_COMPONENT_RESOURCE; v->of.resource = resource; }
static inline wasmtime_component_resource_any_t *go_component_val_get_resource(const wasmtime_component_val_t *v) { return v->of.resource; }
*/
import "C"

import (
	"fmt"
	"runtime"
)

// ComponentResourceType identifies a host-defined component resource class.
type ComponentResourceType struct {
	_ptr *C.wasmtime_component_resource_type_t
}

func NewComponentResourceType(typeID uint32) *ComponentResourceType {
	resourceType := &ComponentResourceType{_ptr: C.wasmtime_component_resource_type_new_host(C.uint32_t(typeID))}
	runtime.SetFinalizer(resourceType, func(resourceType *ComponentResourceType) { resourceType.Close() })
	return resourceType
}

func (t *ComponentResourceType) ptr() *C.wasmtime_component_resource_type_t {
	if t == nil || t._ptr == nil {
		panic("component resource type has been closed")
	}
	return t._ptr
}

func (t *ComponentResourceType) Clone() *ComponentResourceType {
	clone := &ComponentResourceType{_ptr: C.wasmtime_component_resource_type_clone(t.ptr())}
	runtime.SetFinalizer(clone, func(clone *ComponentResourceType) { clone.Close() })
	runtime.KeepAlive(t)
	return clone
}

func (t *ComponentResourceType) Equal(other *ComponentResourceType) bool {
	equal := bool(C.wasmtime_component_resource_type_equal(t.ptr(), other.ptr()))
	runtime.KeepAlive(t)
	runtime.KeepAlive(other)
	return equal
}

func (t *ComponentResourceType) Close() {
	if t == nil || t._ptr == nil {
		return
	}
	runtime.SetFinalizer(t, nil)
	C.wasmtime_component_resource_type_delete(t._ptr)
	t._ptr = nil
}

// ComponentHostResource describes a host resource value recovered from a
// ComponentVal.
type ComponentHostResource struct {
	Owned  bool
	Rep    uint32
	TypeID uint32
}

// NewComponentHostResource creates an own or borrow resource runtime value.
func NewComponentHostResource(store Storelike, resource ComponentHostResource) (*ComponentVal, error) {
	host := C.wasmtime_component_resource_host_new(C.bool(resource.Owned), C.uint32_t(resource.Rep), C.uint32_t(resource.TypeID))
	defer C.wasmtime_component_resource_host_delete(host)
	var anyResource *C.wasmtime_component_resource_any_t
	err := C.wasmtime_component_resource_host_to_any(store.Context(), host, &anyResource)
	runtime.KeepAlive(store)
	if err != nil {
		return nil, mkError(err)
	}
	return newComponentVal(func(value *C.wasmtime_component_val_t) { C.go_component_val_resource(value, anyResource) }), nil
}

// TakeHostResource consumes the store-tracked resource handle and returns its
// host representation. Call Close afterward, but do not call DropResource on
// the consumed value.
func (v *ComponentVal) TakeHostResource(store Storelike) (ComponentHostResource, error) {
	if v.Kind() != ComponentValKindResource {
		return ComponentHostResource{}, fmt.Errorf("component value is not a resource")
	}
	var host *C.wasmtime_component_resource_host_t
	err := C.wasmtime_component_resource_any_to_host(store.Context(), C.go_component_val_get_resource(v.ptr()), &host)
	runtime.KeepAlive(v)
	runtime.KeepAlive(store)
	if err != nil {
		return ComponentHostResource{}, mkError(err)
	}
	defer C.wasmtime_component_resource_host_delete(host)
	return ComponentHostResource{
		Owned:  bool(C.wasmtime_component_resource_host_owned(host)),
		Rep:    uint32(C.wasmtime_component_resource_host_rep(host)),
		TypeID: uint32(C.wasmtime_component_resource_host_type(host)),
	}, nil
}

// DropResource releases the component-model ownership tracked by the store.
// Close must still be called afterward to release host-side memory.
func (v *ComponentVal) DropResource(store Storelike) error {
	if v.Kind() != ComponentValKindResource {
		return fmt.Errorf("component value is not a resource")
	}
	err := C.wasmtime_component_resource_any_drop(store.Context(), C.go_component_val_get_resource(v.ptr()))
	runtime.KeepAlive(v)
	runtime.KeepAlive(store)
	if err != nil {
		return mkError(err)
	}
	return nil
}
