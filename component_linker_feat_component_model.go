package wasmtime

// #include <wasmtime.h>
import "C"

import "runtime"

// ComponentLinker is used to satisfy the imports of a [Component] and
// instantiate it. Use [NewComponentLinker] to create one.
type ComponentLinker struct {
	_ptr *C.wasmtime_component_linker_t
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
	maybeGC()
	return ret
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

// TODO: expose ComponentLinker.Root() and the LinkerInstance type so host
// functions, modules, and resources can be defined. The C API has an
// "exclusive access" requirement on the parent linker while a
// LinkerInstance is alive; mirror that with a locking flag (see
// wasmtime-py's `Linker.locked` for the reference pattern). The
// `wasmtime_component_linker_allow_shadowing` knob is meaningful only once
// definitions exist, so it will be wired up alongside the host-side API.
// TODO: WASIp2 / wasi:http integration via `wasmtime_component_linker_add_*`.

// Close deallocates this linker's state explicitly.
//
// For more information see the documentation for engine.Close().
func (l *ComponentLinker) Close() {
	if l._ptr == nil {
		return
	}
	runtime.SetFinalizer(l, nil)
	C.wasmtime_component_linker_delete(l._ptr)
	l._ptr = nil
}
