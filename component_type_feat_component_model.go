package wasmtime

// #include <wasmtime.h>
//
// static inline uint8_t go_component_item_kind(const wasmtime_component_item_t *it) {
//   return it->kind;
// }
//
// static inline void go_component_item_get_type(
//     const wasmtime_component_item_t *it,
//     wasmtime_component_valtype_t *out) {
//   wasmtime_component_valtype_clone(&it->of.type, out);
// }
import "C"

import "runtime"

// ComponentType describes the type of a [Component] — its lists of imports
// and exports. Walk the lists with [ComponentType.ImportNth] and
// [ComponentType.ExportNth]; each entry is a [ComponentItem] whose
// discriminator identifies the kind of item it carries.
//
// A ComponentType is independently owned — closing the parent [Component]
// does not invalidate it. Close it explicitly (or leave it to the
// finalizer) when finished.
type ComponentType struct {
	_ptr *C.wasmtime_component_type_t

	// engine is the [Engine] the parent [Component] was compiled with. The
	// type uses the engine for the import / export queries below, so the
	// engine must outlive the type. Held privately because exposing it does
	// not enable any user-facing capability that the [Component] does not
	// already provide.
	engine *Engine
}

func mkComponentType(ptr *C.wasmtime_component_type_t, engine *Engine) *ComponentType {
	ct := &ComponentType{_ptr: ptr, engine: engine}
	runtime.SetFinalizer(ct, func(ct *ComponentType) {
		ct.Close()
	})
	return ct
}

func (ct *ComponentType) ptr() *C.wasmtime_component_type_t {
	ret := ct._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return ret
}

// ImportCount returns the number of imports declared by this component.
func (ct *ComponentType) ImportCount() int {
	n := C.wasmtime_component_type_import_count(ct.ptr(), ct.engine.ptr())
	runtime.KeepAlive(ct)
	runtime.KeepAlive(ct.engine)
	return int(n)
}

// ExportCount returns the number of exports declared by this component.
func (ct *ComponentType) ExportCount() int {
	n := C.wasmtime_component_type_export_count(ct.ptr(), ct.engine.ptr())
	runtime.KeepAlive(ct)
	runtime.KeepAlive(ct.engine)
	return int(n)
}

// ImportNth returns the name and the [ComponentItem] for the `i`-th import.
// Returns `("", nil)` if `i` is out of range.
func (ct *ComponentType) ImportNth(i int) (string, *ComponentItem) {
	return ct.itemNth(i, true)
}

// ExportNth returns the name and the [ComponentItem] for the `i`-th export.
// Returns `("", nil)` if `i` is out of range.
func (ct *ComponentType) ExportNth(i int) (string, *ComponentItem) {
	return ct.itemNth(i, false)
}

func (ct *ComponentType) itemNth(i int, isImport bool) (string, *ComponentItem) {
	var nameP *C.char
	var nameLen C.size_t
	var item C.wasmtime_component_item_t
	var found C.bool
	if isImport {
		found = C.wasmtime_component_type_import_nth(
			ct.ptr(), ct.engine.ptr(), C.size_t(i),
			&nameP, &nameLen, &item)
	} else {
		found = C.wasmtime_component_type_export_nth(
			ct.ptr(), ct.engine.ptr(), C.size_t(i),
			&nameP, &nameLen, &item)
	}
	runtime.KeepAlive(ct)
	runtime.KeepAlive(ct.engine)
	if !bool(found) {
		return "", nil
	}
	name := C.GoStringN(nameP, C.int(nameLen))
	return name, mkComponentItem(item)
}

// Close deallocates this component type explicitly.
func (ct *ComponentType) Close() {
	if ct._ptr == nil {
		return
	}
	runtime.SetFinalizer(ct, nil)
	C.wasmtime_component_type_delete(ct._ptr)
	ct._ptr = nil
}

// ComponentItemKind discriminates which sub-type a [ComponentItem] carries.
type ComponentItemKind uint8

const (
	ComponentItemKindComponent         ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_COMPONENT
	ComponentItemKindComponentInstance ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_COMPONENT_INSTANCE
	ComponentItemKindModule            ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_MODULE
	ComponentItemKindComponentFunc     ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_COMPONENT_FUNC
	ComponentItemKindResource          ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_RESOURCE
	ComponentItemKindCoreFunc          ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_CORE_FUNC
	ComponentItemKindType              ComponentItemKind = C.WASMTIME_COMPONENT_ITEM_TYPE
)

// ComponentItem is one entry in a component's import or export list.
//
// The struct is a discriminated union: a `kind` tag plus a payload that
// depends on the tag. This release exposes a payload accessor for type-
// alias entries via [ComponentItem.TypeAlias]. Accessors for the remaining
// kinds (component, component-instance, module, component-func, resource,
// core-func) will be wired up in follow-up work.
type ComponentItem struct {
	val    C.wasmtime_component_item_t
	closed bool
}

func mkComponentItem(val C.wasmtime_component_item_t) *ComponentItem {
	it := &ComponentItem{val: val}
	runtime.SetFinalizer(it, func(it *ComponentItem) {
		it.Close()
	})
	return it
}

// Kind returns the discriminator of this item.
func (it *ComponentItem) Kind() ComponentItemKind {
	if it.closed {
		panic("object has been closed already")
	}
	maybeGC()
	return ComponentItemKind(C.go_component_item_kind(&it.val))
}

// TypeAlias returns the [ComponentValType] embedded in this item when its
// kind is [ComponentItemKindType] (a WIT `type X = ...` alias). Returns
// `nil` for any other kind.
//
// The returned [ComponentValType] is independently owned and must be closed
// (or left to the finalizer).
func (it *ComponentItem) TypeAlias() *ComponentValType {
	if it.Kind() != ComponentItemKindType {
		return nil
	}
	var cloned C.wasmtime_component_valtype_t
	C.go_component_item_get_type(&it.val, &cloned)
	runtime.KeepAlive(it)
	return mkComponentValType(cloned)
}

// Close deallocates this item explicitly.
func (it *ComponentItem) Close() {
	if it.closed {
		return
	}
	runtime.SetFinalizer(it, nil)
	C.wasmtime_component_item_delete(&it.val)
	it.closed = true
}
