package wasmtime

// #include <wasmtime.h>
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

// TODO: expose Component.Type() (wasmtime_component_type) and the surrounding
// type-reflection API once the composite WIT type representation is decided.

// Component is a compiled WebAssembly component, the binary representation of
// a component-model artifact. Components are instantiated through a
// [ComponentLinker].
type Component struct {
	_ptr *C.wasmtime_component_t
}

// NewComponent compiles a component-model binary into a [Component].
//
// This function expects a component binary, such as what is produced by Rust's
// `cargo component` tooling or `wasm-tools component new`.
func NewComponent(engine *Engine, wasm []byte) (*Component, error) {
	var wasmPtr *C.uint8_t
	if len(wasm) > 0 {
		wasmPtr = (*C.uint8_t)(unsafe.Pointer(&wasm[0]))
	}
	var ptr *C.wasmtime_component_t
	err := C.wasmtime_component_new(engine.ptr(), wasmPtr, C.size_t(len(wasm)), &ptr)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(wasm)
	if err != nil {
		return nil, mkError(err)
	}
	return mkComponent(ptr), nil
}

func mkComponent(ptr *C.wasmtime_component_t) *Component {
	c := &Component{_ptr: ptr}
	runtime.SetFinalizer(c, func(c *Component) {
		c.Close()
	})
	return c
}

func (c *Component) ptr() *C.wasmtime_component_t {
	ret := c._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return ret
}

// Serialize encodes this component as a sequence of bytes which can later be
// passed to [NewComponentDeserialize] to recover the same component without
// recompiling.
func (c *Component) Serialize() ([]byte, error) {
	retVec := C.wasm_byte_vec_t{}
	err := C.wasmtime_component_serialize(c.ptr(), &retVec)
	runtime.KeepAlive(c)
	if err != nil {
		return nil, mkError(err)
	}
	ret := C.GoBytes(unsafe.Pointer(retVec.data), C.int(retVec.size))
	C.wasm_byte_vec_delete(&retVec)
	return ret, nil
}

// NewComponentDeserialize decodes serialized bytes previously produced by
// [Component.Serialize] back into a [Component].
//
// This function does not take a component-model binary as input. It only
// accepts the result of a prior call to [Component.Serialize].
func NewComponentDeserialize(engine *Engine, encoded []byte) (*Component, error) {
	var encodedPtr *C.uint8_t
	if len(encoded) > 0 {
		encodedPtr = (*C.uint8_t)(unsafe.Pointer(&encoded[0]))
	}
	var ptr *C.wasmtime_component_t
	err := C.wasmtime_component_deserialize(
		engine.ptr(), encodedPtr, C.size_t(len(encoded)), &ptr,
	)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(encoded)
	if err != nil {
		return nil, mkError(err)
	}
	return mkComponent(ptr), nil
}

// NewComponentDeserializeFile is the same as [NewComponentDeserialize] except
// the bytes are read from a file at `path`.
func NewComponentDeserializeFile(engine *Engine, path string) (*Component, error) {
	cs := C.CString(path)
	defer C.free(unsafe.Pointer(cs))
	var ptr *C.wasmtime_component_t
	err := C.wasmtime_component_deserialize_file(engine.ptr(), cs, &ptr)
	runtime.KeepAlive(engine)
	if err != nil {
		return nil, mkError(err)
	}
	return mkComponent(ptr), nil
}

// Clone creates a shallow clone of this component, incrementing its internal
// reference count. The returned component is independently owned and must be
// explicitly closed (or left to the finalizer).
func (c *Component) Clone() *Component {
	ptr := C.wasmtime_component_clone(c.ptr())
	runtime.KeepAlive(c)
	return mkComponent(ptr)
}

// GetExportIndex looks up the export named `name` in this component and
// returns a reusable [ComponentExportIndex] handle pointing at it.
//
// Pass `nil` for `parent` to search the component's root namespace. Pass an
// index returned by an earlier call (whose result identified an exported
// instance) to traverse one level deeper into nested instance exports —
// for example, to walk from a `wasi:http/incoming-handler` instance export
// down to its `handle` function.
//
// Returns `nil` if no matching export is found.
func (c *Component) GetExportIndex(parent *ComponentExportIndex, name string) *ComponentExportIndex {
	var parentPtr *C.wasmtime_component_export_index_t
	if parent != nil {
		parentPtr = parent.ptr()
	}
	idxPtr := C.wasmtime_component_get_export_index(
		c.ptr(),
		parentPtr,
		C._GoStringPtr(name),
		C._GoStringLen(name),
	)
	runtime.KeepAlive(c)
	runtime.KeepAlive(parent)
	runtime.KeepAlive(name)
	if idxPtr == nil {
		return nil
	}
	return mkComponentExportIndex(idxPtr)
}

// Close deallocates this component's state explicitly.
//
// For more information see the documentation for engine.Close().
func (c *Component) Close() {
	if c._ptr == nil {
		return
	}
	runtime.SetFinalizer(c, nil)
	C.wasmtime_component_delete(c._ptr)
	c._ptr = nil
}

// ComponentExportIndex is an opaque, precomputed handle pointing at a known
// export of a component. It has two main uses:
//
//   - Caching: pass an index to [ComponentInstance.GetFuncByIndex] instead
//     of looking up an export by name on every call, useful in hot paths.
//   - Nested traversal: pass an index identifying an exported instance as
//     the `parent` argument to [Component.GetExportIndex] (or
//     [ComponentInstance.GetExportIndex]) to look up an export within that
//     nested instance.
//
// An index owns its own underlying handle and must be closed (or left to
// the finalizer) independently of the component or instance that produced
// it.
type ComponentExportIndex struct {
	_ptr *C.wasmtime_component_export_index_t
}

func mkComponentExportIndex(ptr *C.wasmtime_component_export_index_t) *ComponentExportIndex {
	idx := &ComponentExportIndex{_ptr: ptr}
	runtime.SetFinalizer(idx, func(idx *ComponentExportIndex) {
		idx.Close()
	})
	return idx
}

func (idx *ComponentExportIndex) ptr() *C.wasmtime_component_export_index_t {
	ret := idx._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return ret
}

// Clone creates a separately-owned copy of this index.
func (idx *ComponentExportIndex) Clone() *ComponentExportIndex {
	ptr := C.wasmtime_component_export_index_clone(idx.ptr())
	runtime.KeepAlive(idx)
	return mkComponentExportIndex(ptr)
}

// Close deallocates this index explicitly.
func (idx *ComponentExportIndex) Close() {
	if idx._ptr == nil {
		return
	}
	runtime.SetFinalizer(idx, nil)
	C.wasmtime_component_export_index_delete(idx._ptr)
	idx._ptr = nil
}
