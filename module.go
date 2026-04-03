package wasmtime

// #include "shims.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

// Module is a module which collects definitions for types, functions, tables, memories, and globals.
// In addition, it can declare imports and exports and provide initialization logic in the form of data and element segments or a start function.
// Modules organized WebAssembly programs as the unit of deployment, loading, and compilation.
type Module struct {
	_ptr *C.wasmtime_module_t
}

func mkModule(ptr *C.wasmtime_module_t) *Module {
	module := &Module{_ptr: ptr}
	runtime.SetFinalizer(module, func(module *Module) {
		module.Close()
	})
	return module
}

func (m *Module) ptr() *C.wasmtime_module_t {
	ret := m._ptr
	if ret == nil {
		panic("object has been closed already")
	}
	maybeGC()
	return ret
}

// Close will deallocate this module's state explicitly.
//
// For more information see the documentation for engine.Close()
func (m *Module) Close() {
	if m._ptr == nil {
		return
	}
	runtime.SetFinalizer(m, nil)
	C.wasmtime_module_delete(m._ptr)
	m._ptr = nil

}

// Imports returns a list of `ImportType` which are the items imported by
// this module and are required for instantiation
func (m *Module) Imports() []*ImportType {
	imports := &importTypeList{}
	C.wasmtime_module_imports(m.ptr(), &imports.vec)
	runtime.KeepAlive(m)
	return imports.mkGoList()
}

// Exports returns a list of `ExportType` which are the items that will be
// exported by this module after instantiation.
func (m *Module) Exports() []*ExportType {
	exports := &exportTypeList{}
	C.wasmtime_module_exports(m.ptr(), &exports.vec)
	runtime.KeepAlive(m)
	return exports.mkGoList()
}

type importTypeList struct {
	vec C.wasm_importtype_vec_t
}

func (list *importTypeList) mkGoList() []*ImportType {
	runtime.SetFinalizer(list, func(imports *importTypeList) {
		C.wasm_importtype_vec_delete(&imports.vec)
	})

	ret := make([]*ImportType, int(list.vec.size))
	base := unsafe.Pointer(list.vec.data)
	var ptr *C.wasm_importtype_t
	for i := 0; i < int(list.vec.size); i++ {
		ptr := *(**C.wasm_importtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkImportType(ptr, list)
		ret[i] = ty
	}
	return ret
}

type exportTypeList struct {
	vec C.wasm_exporttype_vec_t
}

func (list *exportTypeList) mkGoList() []*ExportType {
	runtime.SetFinalizer(list, func(exports *exportTypeList) {
		C.wasm_exporttype_vec_delete(&exports.vec)
	})

	ret := make([]*ExportType, int(list.vec.size))
	base := unsafe.Pointer(list.vec.data)
	var ptr *C.wasm_exporttype_t
	for i := 0; i < int(list.vec.size); i++ {
		ptr := *(**C.wasm_exporttype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkExportType(ptr, list)
		ret[i] = ty
	}
	return ret
}

// NewModuleDeserialize decodes and deserializes in-memory bytes previously
// produced by `module.Serialize()`.
//
// This function does not take a WebAssembly binary as input. It takes
// as input the results of a previous call to `Serialize()`, and only takes
// that as input.
//
// If deserialization is successful then a compiled module is returned,
// otherwise nil and an error are returned.
//
// Note that to deserialize successfully the bytes provided must have been
// produced with an `Engine` that has the same compilation options as the
// provided engine, and from the same version of this library.
func NewModuleDeserialize(engine *Engine, encoded []byte) (*Module, error) {
	var encodedPtr *C.uint8_t
	var ptr *C.wasmtime_module_t
	if len(encoded) > 0 {
		encodedPtr = (*C.uint8_t)(unsafe.Pointer(&encoded[0]))
	}
	err := C.wasmtime_module_deserialize(
		engine.ptr(),
		encodedPtr,
		C.size_t(len(encoded)),
		&ptr,
	)
	runtime.KeepAlive(engine)
	runtime.KeepAlive(encoded)

	if err != nil {
		return nil, mkError(err)
	}

	return mkModule(ptr), nil
}

// NewModuleDeserializeFile is the same as `NewModuleDeserialize` except that
// the bytes are read from a file instead of provided as an argument.
func NewModuleDeserializeFile(engine *Engine, path string) (*Module, error) {
	cs := C.CString(path)
	var ptr *C.wasmtime_module_t
	err := C.wasmtime_module_deserialize_file(engine.ptr(), cs, &ptr)
	runtime.KeepAlive(engine)
	C.free(unsafe.Pointer(cs))

	if err != nil {
		return nil, mkError(err)
	}

	return mkModule(ptr), nil
}
