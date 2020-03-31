package wasmtime

// #include <wasm.h>
//
// wasm_module_t* go_module_new(wasm_store_t *store, uint8_t *bytes, size_t len) {
//    wasm_byte_vec_t vec;
//    vec.data = bytes;
//    vec.size = len;
//    return wasm_module_new(store, &vec);
// }
//
// bool go_module_validate(wasm_store_t *store, uint8_t *bytes, size_t len) {
//    wasm_byte_vec_t vec;
//    vec.data = bytes;
//    vec.size = len;
//    return wasm_module_validate(store, &vec);
// }
import "C"
import "runtime"
import "unsafe"
import "errors"

type Module struct {
	_ptr  *C.wasm_module_t
	Store *Store
}

// Compiles a new `Module` from the `wasm` provided with the given configuration
// in `store`.
func NewModule(store *Store, wasm []byte) (*Module, error) {
	if len(wasm) == 0 {
		return nil, errors.New("failed to compile module")
	}
	// We can't create the `wasm_byte_vec_t` here and pass it in because
	// that runs into the error of "passed a pointer to a pointer" because
	// the vec itself is passed by pointer and it contains a pointer to
	// `wasm`. To work around this we insert some C shims above and call
	// them.
	ptr := C.go_module_new(store.ptr(), (*C.uint8_t)(unsafe.Pointer(&wasm[0])), C.size_t(len(wasm)))
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasm)

	if ptr == nil {
		return nil, errors.New("failed to compile module")
	} else {
		return mkModule(ptr, store), nil
	}
}

// Validates whether `wasm` would be a valid wasm module according to the
// configuration in `store`
func ModuleValidate(store *Store, wasm []byte) bool {
	if len(wasm) == 0 {
		return false
	}
	ret := C.go_module_validate(store.ptr(), (*C.uint8_t)(unsafe.Pointer(&wasm[0])), C.size_t(len(wasm)))
	runtime.KeepAlive(store)
	runtime.KeepAlive(wasm)
	return bool(ret)
}

func mkModule(ptr *C.wasm_module_t, store *Store) *Module {
	module := &Module{_ptr: ptr, Store: store}
	runtime.SetFinalizer(module, func(module *Module) {
		C.wasm_module_delete(module._ptr)
	})
	return module
}

func (m *Module) ptr() *C.wasm_module_t {
	ret := m._ptr
	maybeGC()
	return ret
}

type importTypeList struct {
	vec C.wasm_importtype_vec_t
}

// Returns a list of `ImportType` items which are the items imported by this
// module and are required for instantiation.
func (m *Module) Imports() []*ImportType {
	imports := &importTypeList{}
	C.wasm_module_imports(m.ptr(), &imports.vec)
	runtime.KeepAlive(m)
	runtime.SetFinalizer(imports, func(imports *importTypeList) {
		C.wasm_importtype_vec_delete(&imports.vec)
	})

	ret := make([]*ImportType, int(imports.vec.size))
	base := unsafe.Pointer(imports.vec.data)
	var ptr *C.wasm_importtype_t
	for i := 0; i < int(imports.vec.size); i++ {
		ptr := *(**C.wasm_importtype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkImportType(ptr, imports)
		ret[i] = ty
	}
	return ret
}

type exportTypeList struct {
	vec C.wasm_exporttype_vec_t
}

// Returns a list of `ExportType` items which are the items that will
// be exported by this module after instantiation.
func (m *Module) Exports() []*ExportType {
	exports := &exportTypeList{}
	C.wasm_module_exports(m.ptr(), &exports.vec)
	runtime.KeepAlive(m)
	runtime.SetFinalizer(exports, func(exports *exportTypeList) {
		C.wasm_exporttype_vec_delete(&exports.vec)
	})

	ret := make([]*ExportType, int(exports.vec.size))
	base := unsafe.Pointer(exports.vec.data)
	var ptr *C.wasm_exporttype_t
	for i := 0; i < int(exports.vec.size); i++ {
		ptr := *(**C.wasm_exporttype_t)(unsafe.Pointer(uintptr(base) + unsafe.Sizeof(ptr)*uintptr(i)))
		ty := mkExportType(ptr, exports)
		ret[i] = ty
	}
	return ret
}
